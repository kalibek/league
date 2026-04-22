# Backend Implementation Plan — Table Tennis League API

## Stack
- Go (module: `league-api`)
- Gin web framework
- PostgreSQL with sqlx
- go-migrate for schema migrations
- OAuth 2.0 authentication (provider TBD — see open questions)
- Glicko2 rating system (custom implementation in `pkg/glicko2`)

---

## Confirmed Decisions
| # | Decision |
|---|---------|
| A1 | `group_players.league_id` → corrected to `group_id` (FK to `groups`) |
| A2 | `matches` second `score1` → corrected to `score2` |
| A3 | DNS ("did not show") is represented via `withdraw1`/`withdraw2` on matches — no separate flag on `group_players` |
| A4 | `is_non_calculated boolean default false` added to `group_players` for replacement players |
| A5 | `users` table has `email varchar`; OAuth identity stored in separate `user_oauth_accounts` table |
| A6 | Glicko2 runs once per group when finished; all group matches are one rating period; per-match deltas stored for display |
| A7 | CSV columns: `first_name`, `last_name`, `email`, `initial_rating` (deviation + volatility always default) |
| A8 | Multiple leagues may be active simultaneously; concurrent `league_events` within the same league are prevented |
| A9 | Social OAuth providers: Google, Facebook, Apple |

---

## Directory Structure

```
league-api/
├── cmd/
│   └── server/
│       └── main.go              # init config, DB, router, start server
├── internal/
│   ├── config/
│   │   └── config.go            # load from env vars via os.Getenv
│   ├── db/
│   │   └── db.go                # sqlx.Connect, ping, expose *sqlx.DB
│   ├── model/                   # plain Go structs, db tags, json tags
│   │   ├── user.go
│   │   ├── league.go
│   │   ├── event.go
│   │   ├── group.go
│   │   ├── match.go
│   │   └── rating.go
│   ├── repository/              # data access layer — sqlx queries only, no business logic
│   │   ├── interfaces.go        # repository interface definitions
│   │   ├── postgres/            # postgres implementations
│   │   │   ├── user_repo.go
│   │   │   ├── league_repo.go
│   │   │   ├── event_repo.go
│   │   │   ├── group_repo.go
│   │   │   ├── match_repo.go
│   │   │   └── rating_repo.go
│   │   └── sqlite/              # sqlite in-memory implementations for tests
│   │       └── ...              # mirrors postgres/ structure
│   ├── service/                 # business logic — depends on repository interfaces
│   │   ├── auth_service.go      # OAuth token exchange, session creation/validation
│   │   ├── player_service.go    # player CRUD, CSV import, role management
│   │   ├── league_service.go    # league CRUD, event lifecycle, config updates
│   │   ├── group_service.go     # round-robin generation, placement calculation
│   │   ├── match_service.go     # score entry, no-show handling, score recalculation
│   │   ├── draft_service.go     # draft creation, advancement/recession logic
│   │   └── rating_service.go   # Glicko2 orchestration, rating history writes
│   ├── handler/                 # Gin handlers — parse request, call service, write response
│   │   ├── auth.go
│   │   ├── players.go
│   │   ├── leagues.go
│   │   ├── events.go
│   │   ├── groups.go
│   │   ├── matches.go
│   │   └── websocket.go
│   ├── middleware/
│   │   ├── auth.go              # validate session, extract user+roles, set in context
│   │   └── cors.go
│   └── ws/
│       └── hub.go               # WebSocket broadcast hub (register, unregister, broadcast)
├── pkg/
│   └── glicko2/
│       └── glicko2.go           # pure Glicko2 algorithm — no DB deps
├── migrations/
│   ├── 001_create_roles.up.sql
│   ├── 001_create_roles.down.sql
│   ├── 002_create_users.up.sql
│   ├── 002_create_users.down.sql
│   ├── 003_create_user_oauth_accounts.up.sql
│   ├── 003_create_user_oauth_accounts.down.sql
│   ├── 004_create_leagues.up.sql
│   ├── 004_create_leagues.down.sql
│   ├── 005_create_user_roles.up.sql
│   ├── 005_create_user_roles.down.sql
│   ├── 006_create_league_events.up.sql
│   ├── 006_create_league_events.down.sql
│   ├── 007_create_groups.up.sql
│   ├── 007_create_groups.down.sql
│   ├── 008_create_group_players.up.sql
│   ├── 008_create_group_players.down.sql
│   ├── 009_create_matches.up.sql
│   ├── 009_create_matches.down.sql
│   ├── 010_create_rating_history.up.sql
│   └── 010_create_rating_history.down.sql
├── go.mod
└── go.sum
```

---

## Database Migrations

### Migration 001 — roles
```sql
CREATE TABLE roles (
    role_id   SERIAL PRIMARY KEY,
    role_name VARCHAR(50) NOT NULL UNIQUE
);
INSERT INTO roles (role_name) VALUES ('player'), ('umpire'), ('maintainer');
```

### Migration 002 — users
```sql
CREATE TABLE users (
    user_id        BIGSERIAL PRIMARY KEY,
    first_name     VARCHAR(100) NOT NULL,
    last_name      VARCHAR(100) NOT NULL,
    email          VARCHAR(255) NOT NULL UNIQUE,
    current_rating DOUBLE PRECISION NOT NULL DEFAULT 1500,
    deviation      DOUBLE PRECISION NOT NULL DEFAULT 350,
    volatility     DOUBLE PRECISION NOT NULL DEFAULT 0.06,
    created        TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated   TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Migration 003a — user_oauth_accounts
OAuth identity is stored separately to support multiple providers per user in the future.
```sql
CREATE TABLE user_oauth_accounts (
    account_id   BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider     VARCHAR(20) NOT NULL,             -- 'google' | 'facebook' | 'apple'
    provider_sub VARCHAR(255) NOT NULL,            -- subject ID from provider
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_sub)
);
```
Renumber subsequent migrations: `003_create_leagues` → `004`, etc. (update filenames accordingly).

### Migration 003 — leagues
```sql
CREATE TABLE leagues (
    league_id    BIGSERIAL PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    configuration JSONB NOT NULL DEFAULT '{}',
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW()
);
```
Configuration JSON schema:
```json
{
  "numberOfAdvances": 2,
  "numberOfRecedes": 2,
  "gamesToWin": 3,
  "groupSize": 6
}
```

### Migration 004 — user_roles
```sql
CREATE TABLE user_roles (
    user_id    BIGINT NOT NULL REFERENCES users(user_id),
    role_id    INT NOT NULL REFERENCES roles(role_id),
    league_id  BIGINT NOT NULL REFERENCES leagues(league_id),
    created    TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id, league_id)
);
```

### Migration 005 — league_events
```sql
CREATE TYPE event_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'DONE');
CREATE TABLE league_events (
    event_id     BIGSERIAL PRIMARY KEY,
    league_id    BIGINT NOT NULL REFERENCES leagues(league_id),
    status       event_status NOT NULL DEFAULT 'DRAFT',
    title        VARCHAR(255) NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE NOT NULL,
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Migration 006 — groups
```sql
CREATE TYPE group_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'DONE');
CREATE TABLE groups (
    group_id     BIGSERIAL PRIMARY KEY,
    event_id     BIGINT NOT NULL REFERENCES league_events(event_id),
    status       group_status NOT NULL DEFAULT 'DRAFT',
    division     VARCHAR(10) NOT NULL,    -- e.g. 'A', 'B', 'Superleague'
    group_no     INT NOT NULL,            -- 1-based; 0 for Superleague
    scheduled    TIMESTAMP NOT NULL,
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Migration 007 — group_players
```sql
CREATE TABLE group_players (
    group_player_id   BIGSERIAL PRIMARY KEY,
    group_id          BIGINT NOT NULL REFERENCES groups(group_id),   -- corrected from league_id
    user_id           BIGINT NOT NULL REFERENCES users(user_id),
    seed              SMALLINT NOT NULL DEFAULT 0,
    place             SMALLINT NOT NULL DEFAULT 0,     -- 0 = not yet placed
    points            SMALLINT NOT NULL DEFAULT 0,
    tiebreak_points   SMALLINT NOT NULL DEFAULT 0,
    advances          BOOLEAN NOT NULL DEFAULT FALSE,
    recedes           BOOLEAN NOT NULL DEFAULT FALSE,
    is_non_calculated BOOLEAN NOT NULL DEFAULT FALSE,  -- replacement players skip rating+placement
    created           TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, user_id)
);
```
DNS ("did not show") status is derived at query time: a player is considered DNS when all their matches in the group have their withdraw flag set (`withdraw1` or `withdraw2` depending on their position).

### Migration 008 — matches
```sql
CREATE TYPE match_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'DONE');
CREATE TABLE matches (
    match_id          BIGSERIAL PRIMARY KEY,
    group_id          BIGINT NOT NULL REFERENCES groups(group_id),
    group_player1_id  BIGINT REFERENCES group_players(group_player_id),
    group_player2_id  BIGINT REFERENCES group_players(group_player_id),
    score1            SMALLINT,    -- games won by player 1
    score2            SMALLINT,    -- games won by player 2 (corrected from duplicate score1)
    withdraw1         BOOLEAN NOT NULL DEFAULT FALSE,
    withdraw2         BOOLEAN NOT NULL DEFAULT FALSE,
    status            match_status NOT NULL DEFAULT 'DRAFT',
    created           TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated      TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Migration 009 — rating_history
```sql
CREATE TABLE rating_history (
    history_id   BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(user_id),
    match_id     BIGINT NOT NULL REFERENCES matches(match_id),
    delta        DOUBLE PRECISION NOT NULL,
    rating       DOUBLE PRECISION NOT NULL,
    deviation    DOUBLE PRECISION NOT NULL,
    volatility   DOUBLE PRECISION NOT NULL
);
```

---

## Models (internal/model/)

### user.go
```go
type User struct {
    UserID        int64     `db:"user_id"        json:"userId"`
    FirstName     string    `db:"first_name"     json:"firstName"`
    LastName      string    `db:"last_name"      json:"lastName"`
    Email         string    `db:"email"          json:"email"`
    CurrentRating float64   `db:"current_rating" json:"currentRating"`
    Deviation     float64   `db:"deviation"      json:"deviation"`
    Volatility    float64   `db:"volatility"     json:"volatility"`
    Created       time.Time `db:"created"        json:"created"`
    LastUpdated   time.Time `db:"last_updated"   json:"lastUpdated"`
}

// OAuthAccount links a provider identity to a user; stored in user_oauth_accounts.
type OAuthAccount struct {
    AccountID   int64  `db:"account_id"`
    UserID      int64  `db:"user_id"`
    Provider    string `db:"provider"`     // "google" | "facebook" | "apple"
    ProviderSub string `db:"provider_sub"` // subject ID from the provider
}

type Role struct {
    RoleID   int    `db:"role_id"   json:"roleId"`
    RoleName string `db:"role_name" json:"roleName"`
}

type UserRole struct {
    UserID   int64 `db:"user_id"   json:"userId"`
    RoleID   int   `db:"role_id"   json:"roleId"`
    LeagueID int64 `db:"league_id" json:"leagueId"`
}
```

### league.go
```go
type LeagueConfig struct {
    NumberOfAdvances int `json:"numberOfAdvances"`
    NumberOfRecedes  int `json:"numberOfRecedes"`
    GamesToWin       int `json:"gamesToWin"`
    GroupSize        int `json:"groupSize"`
}

type League struct {
    LeagueID     int64          `db:"league_id"    json:"leagueId"`
    Title        string         `db:"title"        json:"title"`
    Description  string         `db:"description"  json:"description"`
    Config       LeagueConfig   `db:"configuration" json:"configuration"`
    Created      time.Time      `db:"created"      json:"created"`
    LastUpdated  time.Time      `db:"last_updated" json:"lastUpdated"`
}
```

### event.go
```go
type EventStatus string
const (
    EventDraft      EventStatus = "DRAFT"
    EventInProgress EventStatus = "IN_PROGRESS"
    EventDone       EventStatus = "DONE"
)

type LeagueEvent struct {
    EventID     int64       `db:"event_id"    json:"eventId"`
    LeagueID    int64       `db:"league_id"   json:"leagueId"`
    Status      EventStatus `db:"status"      json:"status"`
    Title       string      `db:"title"       json:"title"`
    StartDate   time.Time   `db:"start_date"  json:"startDate"`
    EndDate     time.Time   `db:"end_date"    json:"endDate"`
    Created     time.Time   `db:"created"     json:"created"`
    LastUpdated time.Time   `db:"last_updated" json:"lastUpdated"`
}
```

### group.go
```go
type GroupStatus string
const (
    GroupDraft      GroupStatus = "DRAFT"
    GroupInProgress GroupStatus = "IN_PROGRESS"
    GroupDone       GroupStatus = "DONE"
)

type Group struct {
    GroupID     int64       `db:"group_id"    json:"groupId"`
    EventID     int64       `db:"event_id"    json:"eventId"`
    Status      GroupStatus `db:"status"      json:"status"`
    Division    string      `db:"division"    json:"division"`
    GroupNo     int         `db:"group_no"    json:"groupNo"`
    Scheduled   time.Time   `db:"scheduled"   json:"scheduled"`
    Created     time.Time   `db:"created"     json:"created"`
    LastUpdated time.Time   `db:"last_updated" json:"lastUpdated"`
}

type GroupPlayer struct {
    GroupPlayerID   int64     `db:"group_player_id"   json:"groupPlayerId"`
    GroupID         int64     `db:"group_id"          json:"groupId"`
    UserID          int64     `db:"user_id"           json:"userId"`
    Seed            int16     `db:"seed"              json:"seed"`
    Place           int16     `db:"place"             json:"place"`
    Points          int16     `db:"points"            json:"points"`
    TiebreakPoints  int16     `db:"tiebreak_points"   json:"tiebreakPoints"`
    Advances        bool      `db:"advances"          json:"advances"`
    Recedes         bool      `db:"recedes"           json:"recedes"`
    IsNonCalculated bool      `db:"is_non_calculated" json:"isNonCalculated"`
    // DNS status is not stored here; derive it from matches: player is DNS when
    // all their matches in the group have withdraw1 or withdraw2 set to true.
    Created         time.Time `db:"created"           json:"created"`
    LastUpdated     time.Time `db:"last_updated"      json:"lastUpdated"`
}
```

### match.go
```go
type MatchStatus string
const (
    MatchDraft      MatchStatus = "DRAFT"
    MatchInProgress MatchStatus = "IN_PROGRESS"
    MatchDone       MatchStatus = "DONE"
)

type Match struct {
    MatchID         int64       `db:"match_id"          json:"matchId"`
    GroupID         int64       `db:"group_id"          json:"groupId"`
    GroupPlayer1ID  *int64      `db:"group_player1_id"  json:"groupPlayer1Id"`
    GroupPlayer2ID  *int64      `db:"group_player2_id"  json:"groupPlayer2Id"`
    Score1          *int16      `db:"score1"            json:"score1"`
    Score2          *int16      `db:"score2"            json:"score2"`
    Withdraw1       bool        `db:"withdraw1"         json:"withdraw1"`
    Withdraw2       bool        `db:"withdraw2"         json:"withdraw2"`
    Status          MatchStatus `db:"status"            json:"status"`
    Created         time.Time   `db:"created"           json:"created"`
    LastUpdated     time.Time   `db:"last_updated"      json:"lastUpdated"`
}
```

### rating.go
```go
type RatingHistory struct {
    HistoryID  int64   `db:"history_id" json:"historyId"`
    UserID     int64   `db:"user_id"    json:"userId"`
    MatchID    int64   `db:"match_id"   json:"matchId"`
    Delta      float64 `db:"delta"      json:"delta"`
    Rating     float64 `db:"rating"     json:"rating"`
    Deviation  float64 `db:"deviation"  json:"deviation"`
    Volatility float64 `db:"volatility" json:"volatility"`
}
```

---

## Repository Interfaces (internal/repository/interfaces.go)

```go
type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Create(ctx context.Context, u *model.User) (int64, error)
    List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error)
    UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error
}

type OAuthAccountRepository interface {
    GetByProviderSub(ctx context.Context, provider, sub string) (*model.OAuthAccount, error)
    Create(ctx context.Context, a *model.OAuthAccount) error
    ListByUser(ctx context.Context, userID int64) ([]model.OAuthAccount, error)
}

type LeagueRepository interface {
    GetByID(ctx context.Context, id int64) (*model.League, error)
    List(ctx context.Context) ([]model.League, error)
    Create(ctx context.Context, l *model.League) (int64, error)
    UpdateConfig(ctx context.Context, id int64, config model.LeagueConfig) error
    AssignRole(ctx context.Context, ur model.UserRole) error
    RemoveRole(ctx context.Context, userID, leagueID int64, roleID int) error
    GetUserRoles(ctx context.Context, userID, leagueID int64) ([]model.UserRole, error)
}

type EventRepository interface {
    GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error)
    ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error)
    Create(ctx context.Context, e *model.LeagueEvent) (int64, error)
    UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error
}

type GroupRepository interface {
    GetByID(ctx context.Context, id int64) (*model.Group, error)
    ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error)
    Create(ctx context.Context, g *model.Group) (int64, error)
    UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error
    GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error)
    AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error)
    UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error
    // DNS status is derived from matches; no MarkNoShow here.
    // Use MatchRepository.SetWithdraw to mark all matches for a DNS player.
}

type MatchRepository interface {
    GetByID(ctx context.Context, id int64) (*model.Match, error)
    ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error)
    Create(ctx context.Context, m *model.Match) (int64, error)
    UpdateScore(ctx context.Context, id int64, score1, score2 int16) error
    UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error
    BulkCreate(ctx context.Context, matches []model.Match) error
    // SetWithdraw marks one side of a match as a forfeit.
    // position must be 1 or 2; sets the corresponding withdraw flag and match status=DONE.
    SetWithdraw(ctx context.Context, matchID int64, position int) error
}

type RatingRepository interface {
    InsertHistory(ctx context.Context, rh *model.RatingHistory) error
    GetByUser(ctx context.Context, userID int64) ([]model.RatingHistory, error)
    DeleteByGroup(ctx context.Context, groupID int64) error  // for recalculation
}
```

---

## Service Layer

### auth_service.go
Supported providers: **Google**, **Facebook**, **Apple**.

Each provider requires its own OAuth app credentials (client ID + secret) configured via env vars:
`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `FACEBOOK_CLIENT_ID`, `FACEBOOK_CLIENT_SECRET`, `APPLE_CLIENT_ID`, `APPLE_TEAM_ID`, `APPLE_KEY_ID`, `APPLE_PRIVATE_KEY`.

Apple Sign In differs from Google/Facebook: it uses Sign In with Apple (SIWA) which sends a JWT `id_token` instead of an access token, and only reveals the user's email on the very first authorization. The backend must verify the Apple JWT signature using Apple's public keys.

Responsibilities:
- `GetAuthURL(provider string) string` — return provider-specific OAuth redirect URL with state param (CSRF)
- `HandleCallback(provider, code, state string) (*User, string, error)`:
  1. Validate state (anti-CSRF)
  2. Exchange code for access token (or verify Apple `id_token` JWT)
  3. Fetch user profile from provider (name, email, provider subject ID)
  4. Look up `user_oauth_accounts` by `(provider, provider_sub)`
  5. If found → load user; if not found → create user + oauth_account (or link to existing user by email)
  6. Issue a signed JWT with `userID` claim
- `ValidateToken(token string) (userID int64, err error)`
- JWT signed with `JWT_SECRET` env var, 30-day expiry

### player_service.go
Responsibilities:
- `CreatePlayer(firstName, lastName, email string) (*User, error)` — Maintainer action
- `ImportCSV(r io.Reader) (ImportResult, error)` — parse CSV rows, validate, bulk-insert users
  - CSV columns: `first_name` (required), `last_name` (required), `email` (required), `initial_rating` (optional float, defaults to 1500)
  - `deviation` and `volatility` always use Glicko2 defaults (350, 0.06) regardless of CSV content
  - Return counts: imported, skipped (duplicates by email), errors per row with row number and message
- `GetProfile(userID int64) (*PlayerProfile, error)` — user + all group participations + rating history

### league_service.go
Responsibilities:
- `CreateLeague(creatorID int64, title, description string, config LeagueConfig) (*League, error)`
  - Creates league
  - Assigns creator as Maintainer
  - Assigns creator as Player
- `UpdateConfig(leagueID int64, config LeagueConfig) error`
- `AssignRole(leagueID, targetUserID int64, roleName string) error` — Maintainer only
- `RemoveRole(leagueID, targetUserID int64, roleName string) error`

### event_service.go (extracted from league_service for clarity)
Responsibilities:
- `CreateDraftEvent(leagueID int64, ...) (*LeagueEvent, error)`
  - **Guard**: check there is no existing event with status `DRAFT` or `IN_PROGRESS` for this league; return error if found (concurrent events within a league are prevented)
  - Creates new LeagueEvent with status=DRAFT
- `StartEvent(eventID int64) error` — set status to IN_PROGRESS
- `UpdateEventConfig(eventID int64, config LeagueConfig) error` — delegates to `DraftService.RecreateDraft`

### group_service.go
Responsibilities:
- `GenerateRoundRobin(groupID int64) error`
  - Fetch group players
  - Create n*(n-1)/2 match records in DRAFT status
  - Skip non-calculated players in match generation
- `CalculatePlacements(groupID int64) (needsManual []int64, err error)`
  - Implements full tiebreak algorithm (see Business Logic section)
  - Returns list of groupPlayerIDs that need manual ordering
- `SetManualPlace(groupPlayerID int64, place int16) error` — Umpire override for tiebreak ties
- `AddNonCalculatedPlayer(groupID, userID int64) error`
  - Validates user not in another group in same event
  - Inserts as is_non_calculated=true

### match_service.go
Responsibilities:
- `UpdateScore(matchID int64, score1, score2 int16) error`
  - Validate scores: winner must reach gamesToWin, loser score < gamesToWin
  - Mark match DONE
  - Recalculate points and tiebreak_points for both group_players
  - Trigger placement recalculation for group
  - Broadcast via WebSocket hub
- `MarkNoShow(groupPlayerID int64) error`
  - Fetch all matches in the group where this groupPlayerID is `group_player1_id` or `group_player2_id`
  - For each match: call `MatchRepository.SetWithdraw(matchID, position)` (position=1 if player is player1, 2 if player2)
  - `SetWithdraw` also sets scores: DNS player gets 0 games, opponent gets `gamesToWin` games, match → DONE
  - Recalculate points for all affected group_players (DNS player scores 0 pts per match, opponent scores 2 pts)
  - Trigger placement recalculation for group
  - Broadcast via WebSocket
  - DNS detection query: `SELECT gp.* FROM group_players gp WHERE gp.group_id = $1 AND NOT EXISTS (SELECT 1 FROM matches m WHERE (m.group_player1_id = gp.group_player_id AND NOT m.withdraw1) OR (m.group_player2_id = gp.group_player_id AND NOT m.withdraw2))`

### draft_service.go
Responsibilities:
- `CreateDraft(leagueID, finishedEventID int64) (*LeagueEvent, error)`
  - Requires all groups in finishedEventID to be DONE
  - Creates new LeagueEvent (status=DRAFT, title=next month)
  - For each group in finished event, using `numberOfAdvances`/`numberOfRecedes` config:
    - Top N players → advance to higher group (or stay if already top group A0/Superleague)
    - Bottom N players → recede to lower group (or stay if already lowest)
    - Middle players → stay in same group
  - Creates group records and group_player records for new event
  - Generates round-robin match stubs (DRAFT status) for each group
- `RecreateDraft(eventID int64, newConfig LeagueConfig) error`
  - Only allowed if event status = DRAFT
  - Delete all group_players and matches for the event
  - Delete all groups for the event
  - Re-run draft creation with new config
  - Used when config changes before event starts

### rating_service.go
Responsibilities:
- `CalculateGroupRatings(groupID int64) error`
  - Fetch all DONE matches in group
  - For each player in group:
    - Collect all opponents and match outcomes
    - Run Glicko2 with all matches as one rating period
    - Compute rating delta
  - Update user current_rating, deviation, volatility
  - Insert rating_history records for each match with the per-match share of delta
  - Skip is_non_calculated players
- `RecalculateGroupRatings(groupID int64) error` — called when score edited after group is DONE
  - Delete existing rating_history records for all matches in group
  - Re-run CalculateGroupRatings

---

## Business Logic: Placement Algorithm

```
CalculatePlacements(groupID):
  players = getGroupPlayers(groupID, excludeNonCalculated=true, excludeNoShow=false)
  
  // Phase 1: sort by points DESC
  groups = groupByEqualPoints(players)
  
  for each tiedGroup in groups:
    if len(tiedGroup) == 1:
      assign next place
    elif len(tiedGroup) == 2:
      winner = headToHeadWinner(tiedGroup[0], tiedGroup[1])
      assign places by winner first
    else:
      // Phase 2: calculate tiebreak points
      for each player in tiedGroup:
        tiebreak = sum(score1) - sum(score2)  // games won minus games lost across all matches
      subGroups = groupByEqualTiebreak(tiedGroup)
      
      for each subGroup in subGroups:
        if len(subGroup) == 1:
          assign next place
        elif len(subGroup) == 2:
          winner = headToHeadWinner(subGroup[0], subGroup[1])
          assign places by winner first
        else:
          // Manual intervention required
          return needsManual = subGroup.playerIDs

  return needsManual=[], err=nil
```

DNS players (all matches have withdraw flag set for their position): placed last. Since their withdraw matches score 0 games won and `gamesToWin` games lost, their points = 0 and tiebreak_points will be a large negative number, naturally sorting them to the bottom. No special-casing needed in the placement algorithm; withdraw-based scoring handles it.

---

## Glicko2 Implementation (pkg/glicko2/)

```go
type Player struct {
    Rating     float64  // μ
    Deviation  float64  // φ
    Volatility float64  // σ
}

type MatchResult struct {
    Opponent Player
    Score    float64  // 1.0 = win, 0.0 = loss
}

// Calculate returns new rating, deviation, volatility for one rating period
func Calculate(player Player, results []MatchResult) Player
```

Constants: τ (system constant) = 0.5 (configurable), ε (convergence) = 0.000001.
Algorithm follows Glickman's 2012 paper step-by-step.

---

## API Routes

### Auth
```
GET  /auth/login           → redirect to OAuth provider
GET  /auth/callback        → exchange code, issue session, redirect to frontend
POST /auth/logout          → clear session
GET  /auth/me              → current user + role map
```

### Players
```
GET  /api/v1/players               → list players (query: sort=rating|name, limit, offset)
POST /api/v1/players               → create player [Maintainer]
GET  /api/v1/players/:id           → player profile (rating history, group history)
POST /api/v1/players/import        → CSV import [Maintainer] (multipart/form-data)
```

### Leagues
```
GET  /api/v1/leagues               → list all leagues
POST /api/v1/leagues               → create league [authenticated]
GET  /api/v1/leagues/:id           → league detail + current event summary
PUT  /api/v1/leagues/:id/config    → update config [Maintainer]
POST /api/v1/leagues/:id/roles     → assign role [Maintainer] body: {userId, role}
DELETE /api/v1/leagues/:id/roles/:userId/:role → remove role [Maintainer]
```

### Events
```
GET  /api/v1/leagues/:id/events            → list events for league
POST /api/v1/leagues/:id/events            → create draft for next month [Maintainer]
GET  /api/v1/leagues/:id/events/:eid       → event detail (all groups + players + matches)
PUT  /api/v1/leagues/:id/events/:eid/config → update event config, recreate draft [Maintainer]
POST /api/v1/leagues/:id/events/:eid/start → start event: set status IN_PROGRESS [Maintainer]
```

### Groups
```
GET  /api/v1/events/:eid/groups/:gid                             → group detail
POST /api/v1/events/:eid/groups/:gid/finish                      → finish group [Umpire]
POST /api/v1/events/:eid/groups/:gid/players                     → add non-calculated player [Maintainer]
PUT  /api/v1/events/:eid/groups/:gid/players/:pid/no-show        → mark no-show [Maintainer]
PUT  /api/v1/events/:eid/groups/:gid/players/:pid/place          → manual place [Umpire]
```

### Matches
```
PUT  /api/v1/groups/:gid/matches/:mid      → update score [Umpire]
```

### WebSocket
```
WS   /ws/events/:eid                       → subscribe to live event updates
```

WebSocket message format:
```json
{
  "type": "match_updated" | "group_finished" | "event_finished" | "manual_placement_required",
  "groupId": 123,
  "matchId": 456,
  "payload": { ... }
}
```

---

## Authentication Middleware (internal/middleware/auth.go)

- Extract JWT from `Authorization: Bearer <token>` header or `session` cookie
- Decode and validate signature
- Load user from DB if not in JWT claims
- Attach `userID` and `roles map[leagueID][]string` to Gin context
- Role check helpers: `RequireAuth()`, `RequireMaintainer(leagueID)`, `RequireUmpire(leagueID)`

---

## WebSocket Hub (internal/ws/hub.go)

```go
type Hub struct {
    // eventID → set of client connections
    rooms map[int64]map[*Client]bool
    // channels
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
}

type Client struct {
    eventID int64
    conn    *websocket.Conn
    send    chan []byte
}

func (h *Hub) Run()  // main goroutine: select on channels
func (h *Hub) BroadcastToEvent(eventID int64, msg Message)
```

Gin handler upgrades HTTP to WebSocket, registers client with Hub, starts read/write pumps.

---

## Testing Strategy

### Unit Tests (service layer)
- File: `internal/service/*_test.go`
- Pattern: table-driven tests
```go
func TestCalculatePlacements(t *testing.T) {
    tests := []struct {
        name     string
        players  []model.GroupPlayer
        matches  []model.Match
        expected []int16  // expected place order by seed
    }{
        {name: "clear winner", ...},
        {name: "two-way tie head-to-head", ...},
        {name: "three-way tiebreak", ...},
        {name: "manual placement required", ...},
    }
    for _, tc := range tests { ... }
}
```
- Mock repository interfaces using hand-written test doubles (no mock libraries)

### Integration Tests (repository layer)
- Use SQLite in-memory via `modernc.org/sqlite` (CGO-free)
- Apply same migrations to SQLite (with minor SQL dialect adjustments)
- Test actual SQL queries against SQLite
- File: `internal/repository/sqlite/*_test.go`

### Test data
- Use helper `func newTestGroup(t *testing.T, n int) (group, players, matches)` builders

---

## Implementation Phases

### Phase 1: Project skeleton (Day 1)
- `go mod init league-api`
- Dependencies: `github.com/gin-gonic/gin`, `github.com/jmoiron/sqlx`, `github.com/golang-migrate/migrate/v4`, `github.com/gorilla/websocket`, `github.com/golang-jwt/jwt/v5`, `lib/pq`
- `cmd/server/main.go`: load config, connect DB, run migrations, start Gin
- `internal/config/config.go`: env var loading (DB_URL, PORT, OAUTH_*, JWT_SECRET)
- `internal/db/db.go`: sqlx connect + ping

### Phase 2: Migrations + Models (Day 1-2)
- Write all 9 migration pairs
- Write all model structs

### Phase 3: Repository layer (Day 2-3)
- Implement all repository interfaces for Postgres
- Write SQLite equivalents for test use
- Write integration tests for each repository

### Phase 4: Glicko2 (Day 3)
- Implement algorithm in `pkg/glicko2/glicko2.go`
- Unit tests with known input/output from Glickman's paper examples

### Phase 5: Services (Day 4-6)
- Implement all services with unit tests
- Focus order: auth → league → group → match → draft → rating

### Phase 6: HTTP handlers + router (Day 6-7)
- Wire Gin routes to services
- Implement auth middleware (JWT)
- Implement CORS middleware

### Phase 7: WebSocket (Day 7)
- Implement Hub
- Wire into match score update flow

### Phase 8: Integration + polish (Day 8)
- End-to-end manual testing with Postgres
- Error handling consistency
- API response shape consistency
