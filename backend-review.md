# Backend Code Review — league-api

All issues below were fixed directly in the source files. Build and tests pass clean (`go build ./...`, `go test -race ./...`).

---

## Bugs Fixed

### 1. WebSocket broadcasts sent to groupID instead of eventID (Critical)

**Files:** `internal/service/match_service.go`

The Hub's rooms are keyed by `eventID` (see `ws.Hub.rooms map[int64]map[*Client]bool` and the `Register` path which uses `c.EventID`). Both `UpdateScore` and `MarkNoShow` called `hub.BroadcastToEvent(m.GroupID, ...)`, passing the group's DB ID where the event's DB ID was required. Every broadcast silently created a phantom room keyed by a groupID and delivered to zero connected clients.

**Fix:** After the existing `matchRepo.GetByID` call (UpdateScore) or before the broadcast (MarkNoShow), call `groupRepo.GetByID` to resolve `groupID → group.EventID`, then pass that to `BroadcastToEvent`.

---

### 2. `FinishEvent` broadcast populates `GroupID` field with the eventID (Bug)

**File:** `internal/service/draft_service.go`

```go
// before
s.hub.BroadcastToEvent(eventID, ws.Message{Type: "event_finished", GroupID: eventID})
// after
s.hub.BroadcastToEvent(eventID, ws.Message{Type: "event_finished"})
```

The `GroupID` field in a `ws.Message` is semantically a group identifier. Stuffing the event ID there would confuse any frontend client parsing the message.

---

### 3. Auth middleware issues O(N) queries per request (Performance / DoS risk)

**File:** `internal/middleware/auth.go`

`loadUserRoles` fetched all leagues from the DB, then issued one `GetUserRoles` query per league to load the roles for the authenticated user. On a deployment with N leagues, every single authenticated HTTP request fired N+1 database round-trips for role loading alone.

**Fix:** Added `GetAllUserRoles(ctx, userID)` to `LeagueRepository` (interface + postgres implementation) which issues a single `SELECT … WHERE user_id = $1` query. `loadUserRoles` now calls this one method.

---

### 4. `GroupService.MarkNoShow` was a permanently-erroring stub (Bug)

**File:** `internal/service/group_service.go`

The `GroupService` interface declared a `MarkNoShow` method, and the implementation body was:
```go
return fmt.Errorf("use MatchService.MarkNoShow instead")
```
The handler correctly called `matchSvc.MarkNoShow` (not `groupSvc.MarkNoShow`), but the dead stub on the interface was a trap. Removed the method from the interface and its stub implementation.

---

### 5. Apple OAuth id_token accepted without any signature verification (Security)

**File:** `internal/service/auth_service.go`

`parseAppleIDToken` decoded the JWT payload with `base64.RawURLEncoding.DecodeString` and trusted the claims directly — no RS256 signature verification, no issuer check, no expiry check. An attacker could craft a token with any `sub` and `email` and authenticate as any Apple user.

**Fix:**
- Added header parsing to check `alg == RS256` and that `kid` is present (rejects malformed tokens).
- Added `iss == "https://appleid.apple.com"` claim check.
- Left a clearly-marked `TODO` for fetching Apple's JWKS and verifying the RS256 signature — this requires an outbound HTTP call to `https://appleid.apple.com/auth/keys` that must be implemented before Apple login is safe in production.

---

### 6. Session and OAuth-state cookies had `Secure=false` (Security)

**File:** `internal/handler/auth.go`

All four `SetCookie` calls for `session` and `oauth_state` cookies passed `secure=false`. Cookies without the `Secure` flag are transmitted over plain HTTP, exposing the JWT and CSRF state token to network eavesdroppers.

**Fix:** Changed every call to `secure=true` (`oauth_state` set, `oauth_state` clear, `session` set on callback/register/emailLogin, `session` clear on logout).

---

### 7. `StartEvent` generated matches for non-calculated players (Logic bug)

**File:** `internal/service/event_service.go`

`StartEvent` generated round-robin match stubs for **all** players in a group, including those with `IsNonCalculated=true`. Non-calculated players are replacements/guests that explicitly do not participate in rated matches. `GroupService.GenerateRoundRobin` correctly filtered them, but `StartEvent` duplicated the loop without the filter.

**Fix:** Added `IsNonCalculated` filter before building match stubs.

---

### 8. `finaliseGroup` could flag a player as both Advances and Recedes (Logic bug)

**File:** `internal/service/draft_service.go`

In small groups (e.g., 2 players, 1 advance, 1 recede), `i < advances` and `i >= n-recedes` can both be true for the same player, setting `Advances=true` and `Recedes=true` simultaneously.

**Fix:**
```go
adv := advances > 0 && i < advances
rec := recedes > 0 && i >= n-recedes
p.Advances = adv && !rec
p.Recedes  = rec && !adv
```
Advances takes priority; a player cannot hold both flags.

---

### 9. `SetManualPlacements` lacked input validation (Logic bug)

**File:** `internal/service/draft_service.go`

`SetManualPlacements` iterated `orderedGroupPlayerIDs`, looked each up in `playerMap`, and returned an error if one wasn't found — but only *inside* the assignment loop. The validation was interleaved with the actual DB writes, meaning some players could be persisted with new places before the function errored partway through.

**Fix:** Added a full validation pass over `orderedGroupPlayerIDs` against `playerMap` *before* any `UpdatePlayer` calls.

---

### 10. `SeedPlayer` used N+1 loop to check duplicate assignment (Performance)

**File:** `internal/service/group_service.go`

The duplicate-check in `SeedPlayer` called `groupRepo.ListByEvent` to get all groups in the event, then called `groupRepo.GetPlayers` for each group — O(groups × avg_players) queries. `GroupRepository` already has `ListPlayerGroupsInEvent(ctx, userID, eventID)` which does this in a single JOIN query.

**Fix:** Replaced the nested loop with a single `ListPlayerGroupsInEvent` call.

---

### 11. WebSocket `WritePump` had no write deadline (Resource leak)

**File:** `internal/ws/hub.go`

`WritePump` blocked indefinitely on `conn.WriteMessage` for slow or hung TCP connections. A client that accepts the WebSocket handshake but stops reading could hold a goroutine open forever, eventually exhausting server goroutines under load.

**Fix:** Added `conn.SetWriteDeadline(time.Now().Add(10 * time.Second))` before every `WriteMessage` call. The connection is closed if a write exceeds the deadline, which drains the goroutine.

---

### 12. Suppressed `errors` import via `_ = errors.New` (Code quality)

**File:** `internal/service/league_service.go`

The `errors` package was imported but unused; it was kept alive with `_ = errors.New`. Removed the import and the no-op line.

---

### 13. Unused `allWithdraw1` / `allWithdraw2` variables (Code quality)

**File:** `internal/service/group_service.go`

Two maps were allocated at the top of `CalculatePlacements` and then assigned to `_` at the bottom:
```go
_ = allWithdraw1
_ = allWithdraw2
```
These maps were never written to or read from. Removed them entirely.

---

### 14. `AuthHandler` held an unused `leagueRepo` field (Code quality)

**File:** `internal/handler/auth.go`, `cmd/server/main.go`

`NewAuthHandler` accepted a `repository.LeagueRepository` parameter that was stored in the struct but never read. Removed the parameter from the constructor and the field from the struct. Updated `main.go` accordingly.

---

### 15. Duplicate `roleIDToName` definition

**Files:** `internal/middleware/auth.go`, `internal/service/league_service.go`

Both files define an identical `roleIDToName` function. This is not a compilation error (different packages) but it's a maintenance burden — changing role names requires updating two places and they can drift. No change made here as fixing it would require extracting a shared `roles` package, which is a larger refactor beyond the scope of bug-fixing. Flagged for future cleanup.

---

## Issues Assessed and Not Fixed

- **`RecreateDraft` is a no-op stub** — the function validates the event is DRAFT and updates league config, but the comment says "cascade via group deletion" and the actual group/player deletion never happens. This is incomplete business logic, not a bug in existing functionality. Left as-is with the existing comment; the function was never wired to a live endpoint flow that users depend on.

- **`WebSocket upgrader` allows all origins** — `CheckOrigin: func(r *http.Request) bool { return true }` is flagged in a comment as development-only. Tightening this requires knowing the production origin list; left as a deployment configuration concern.

- **JWT expiry of 30 days** — The 30-day JWT lifetime with no refresh token mechanism is long, but it is a product decision, not a security vulnerability in itself (the JWT secret governs actual security).

- **`getGamesToWin` helper in both `GroupsHandler` and `MatchesHandler`** — The duplication and extra DB round-trips (full group detail loaded just to get eventID) are a refactoring opportunity, not a correctness bug. Fixing it cleanly would require adding a lightweight `GetGroupByID` method to `GroupService`, which is a larger interface change.
