# Transaction Support Implementation Plan

## Context

The backend performs multi-step database writes in service-layer functions with no atomicity guarantees. A failure mid-way leaves the database in a corrupt/inconsistent state:

- `CreateDraft` creates an event then groups then players — a crash between steps leaves orphaned events with no groups
- `StartEvent` creates matches then updates group/event status — partial completion leaves groups IN_PROGRESS but event still DRAFT
- `finaliseGroup` updates N player rows then group status — partial player updates corrupt standings
- `Register` / `HandleCallback` create a user then set password/OAuth link — user exists but cannot authenticate
- `CreateLeague` creates a league then assigns roles — league exists with no owner
- Rating recalculation deletes history then rewrites it — interruption leaves no rating history at all

Currently **zero** transaction wrappers exist in the service layer. Only `matchRepo.BulkCreate` uses an internal transaction. No `BeginTx`, `WithTx`, or UnitOfWork pattern is defined anywhere.

---

## Approach: Context-propagated transactions

Store the active `*sqlx.Tx` in the request context. Each repo method checks context for a live transaction; if found, uses it instead of the connection pool. Services wrap multi-step operations in a `RunInTx` helper.

**Why this approach:**
- Repo method signatures stay unchanged — no interface churn
- Services remain testable with the current mock pattern (mocks ignore tx in context)
- Incremental adoption: wrap one function at a time
- Both `*sqlx.DB` and `*sqlx.Tx` satisfy the same internal `DBTX` interface

---

## Files to create / modify

### New file: `internal/db/tx.go`

```go
package db

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/jmoiron/sqlx"
)

type txKey struct{}

// DBTX is satisfied by both *sqlx.DB and *sqlx.Tx.
type DBTX interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// InjectTx stores tx in ctx (used inside RunInTx).
func InjectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
    return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx retrieves the active transaction from ctx; nil if none.
func ExtractTx(ctx context.Context) *sqlx.Tx {
    tx, _ := ctx.Value(txKey{}).(*sqlx.Tx)
    return tx
}

// RunInTx begins a transaction, injects it into ctx, calls fn(txCtx),
// and commits on success or rolls back on error/panic.
func RunInTx(ctx context.Context, db *sqlx.DB, fn func(txCtx context.Context) error) error {
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    txCtx := InjectTx(ctx, tx)
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()
    if err := fn(txCtx); err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
```

### Modified: all repo structs in `internal/repository/postgres/`

Each repo struct changes `db *sqlx.DB` → `pool *sqlx.DB` and adds a private helper:

```go
func (r *groupRepo) db(ctx context.Context) db.DBTX {
    if tx := db.ExtractTx(ctx); tx != nil {
        return tx
    }
    return r.pool
}
```

All method bodies replace `r.db.X(ctx, ...)` → `r.db(ctx).X(ctx, ...)`.

Affected repos (every repo whose methods are called inside the transactions below):
- `postgres/group_repo.go` — `UpdatePlayer`, `UpdateStatus`, `Create`, `AddPlayer`, `GetPlayers`, `GetByID`, `ListByEvent`, `ResetGroupPlayers`, `GetPlayersByMovement`
- `postgres/match_repo.go` — `GetByID`, `UpdateScore`, `UpdateStatus`, `SetWithdraw`, `BulkCreate`, `ListByGroup`
- `postgres/event_repo.go` — `Create`, `UpdateStatus`, `GetByID`, `ListByLeague`
- `postgres/league_repo.go` — `Create`, `AssignRole`, `GetByID`
- `postgres/user_repo.go` — `Create`, `SetPasswordHash`, `UpdateRating`, `ResetAllRatings`, `GetByID`, `GetByEmail`
- `postgres/oauth_repo.go` — `Create`, `GetByProviderSub`
- `postgres/rating_repo.go` — `InsertHistory`, `DeleteByGroup`, `DeleteAll`
- `postgres/profile_repo.go` — `UpsertProfile`

`matchRepo.BulkCreate` already has an internal transaction — replace it to use the context-propagated tx when one is active; otherwise behave as today.

### Modified: service constructors (inject `*sqlx.DB`)

Services that start transactions need access to `*sqlx.DB`. Add a `db *sqlx.DB` field to each:

| Service | File | Functions to wrap |
|---------|------|-------------------|
| `draftService` | `draft_service.go` | `CreateDraft`, `finaliseGroup`, `SetManualPlacements` |
| `eventService` | `event_service.go` | `StartEvent` |
| `authService` | `auth_service.go` | `HandleCallback`, `Register` |
| `leagueService` | `league_service.go` | `CreateLeague` |
| `ratingService` | `rating_service.go` | `CalculateGroupRatings`, `RecalculateGroupRatings`, `RecalculateAllRatings` |
| `matchService` | `match_service.go` | `UpdateScore`, `SetMatchWalkover`, `MarkNoShow` |
| `groupService` | `group_service.go` | `CalculatePlacements` |
| `profileService` | `profile_service.go` | `UpsertProfile` |

Update corresponding `NewXxxService` constructors to accept `*sqlx.DB` as the first parameter. Update `cmd/server/main.go` to pass `database` to each.

---

## Transaction boundaries (priority order)

### 1. `draftService.CreateDraft` — CRITICAL
Wraps: `eventRepo.Create` → per-group `groupRepo.Create` → per-player `groupRepo.AddPlayer`

Risk without tx: orphaned event + empty groups persisted forever with no cleanup path.

### 2. `draftService.finaliseGroup` — CRITICAL
Wraps: `ratingSvc.CalculateGroupRatings` → N×`groupRepo.UpdatePlayer` (advances/recedes) → `groupRepo.UpdateStatus`

Note: `ratingSvc.CalculateGroupRatings` must also run inside the same tx; pass `txCtx` through.

### 3. `eventService.StartEvent` — CRITICAL
Wraps: per-group loop of `matchRepo.BulkCreate` + `groupRepo.UpdateStatus` → `eventRepo.UpdateStatus`

### 4. `authService.HandleCallback` — HIGH
Wraps: `userRepo.Create` → `oauthRepo.Create`

### 5. `authService.Register` — HIGH
Wraps: `userRepo.Create` → `userRepo.SetPasswordHash`

### 6. `leagueService.CreateLeague` — HIGH
Wraps: `leagueRepo.Create` → `leagueRepo.AssignRole` (×2)

### 7. `ratingService.RecalculateGroupRatings` — HIGH
Wraps: `ratingRepo.DeleteByGroup` → `CalculateGroupRatings` (history inserts + user rating updates)

### 8. `ratingService.RecalculateAllRatings` — HIGH
Wraps: `ratingRepo.DeleteAll` → `userRepo.ResetAllRatings` → all-events recalculation loop

### 9. `matchService.UpdateScore` / `SetMatchWalkover` / `MarkNoShow` — HIGH
Wraps: `matchRepo.UpdateScore` → `matchRepo.UpdateStatus` (or `SetWithdraw`) → `recalcGroupPoints` (N×`groupRepo.UpdatePlayer`)

### 10. `draftService.SetManualPlacements` — MEDIUM
Wraps: N×`groupRepo.UpdatePlayer` (place assignment) → `finaliseGroup` (already transactional after step 2)

### 11. `groupService.CalculatePlacements` — MEDIUM
Wraps: all nested `groupRepo.UpdatePlayer` calls in tiebreak resolution loops

### 12. `profileService.UpsertProfile` — LOW
Wraps: `userRepo.UpdateName` → `profileRepo.UpsertProfile`

---

## Usage pattern in service methods

```go
func (s *draftService) CreateDraft(ctx context.Context, leagueID, finishedEventID int64) (*model.LeagueEvent, error) {
    // reads happen outside tx (no need to hold lock during validation)
    groups, err := s.groupRepo.ListByEvent(ctx, finishedEventID)
    // ... validation ...

    var newEvent *model.LeagueEvent
    err = db.RunInTx(ctx, s.db, func(txCtx context.Context) error {
        eventID, err := s.eventRepo.Create(txCtx, newEvt)        // uses tx
        if err != nil { return err }
        for _, div := range divisions {
            gid, err := s.groupRepo.Create(txCtx, grp)            // uses tx
            if err != nil { return err }
            for _, uid := range players {
                _, err = s.groupRepo.AddPlayer(txCtx, gp)         // uses tx
                if err != nil { return err }
            }
        }
        newEvent, err = s.eventRepo.GetByID(txCtx, eventID)
        return err
    })
    return newEvent, err
}
```

---

## Test impact

Existing unit tests use hand-rolled mocks that implement repo interfaces. The mocks never touch `*sqlx.Tx`, so they automatically satisfy the new behaviour (context tx check is a no-op when no tx is injected). **No mock changes required.**

Service constructors gain a `*sqlx.DB` parameter. Test code that constructs services directly must pass `nil` for `db` (safe — `nil` is only dereferenced inside `RunInTx`, which tests never call directly).

---

## Verification

```bash
# Backend compiles
cd league-api && go build ./...

# All tests still pass
cd league-api && go test ./...

# Manual smoke test (requires running DB):
# 1. Start server: go run ./cmd/server
# 2. Create league → assign roles → verify: if role assign fails, league must not exist
# 3. Create draft → introduce artificial failure mid-group → verify no orphaned event in DB
# 4. Start event → check all groups flip to IN_PROGRESS atomically
```

---

## Out of scope

- `RecreateDraft` — function is incomplete (deletion logic missing). Fix the function first, then add tx.
- Handler/middleware layer — transactions are service-layer concerns only.
- Distributed transactions — single PostgreSQL instance, standard `BEGIN`/`COMMIT` is sufficient.
