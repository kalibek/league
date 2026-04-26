# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Structure

Monorepo with two subprojects:
- `league-api/` — Go backend (Gin, PostgreSQL, sqlx, Glicko2 ratings)
- `league-ui/` — React frontend (Vite, TypeScript, Tailwind v4, React Router v7)

## Backend (league-api)

### Commands

```bash
# Start Postgres (use podman, not docker)
cd league-api && podman compose up -d

# Run server (auto-runs migrations on start)
cd league-api && go run ./cmd/server

# Run all tests
cd league-api && go test ./...

# Run single test
cd league-api && go test ./internal/service/ -run TestGroupService

# Run glicko2 tests
cd league-api && go test ./pkg/glicko2/
```

### Architecture

```
cmd/server/main.go       — wires everything: repos → services → handlers → gin router
internal/config/         — env-based config (DATABASE_URL, PORT, JWT_SECRET, etc.)
internal/db/             — sqlx connection + golang-migrate migrations
internal/model/          — domain structs (User, League, LeagueEvent, Group, Match, etc.)
internal/repository/     — interfaces.go defines all repo interfaces; postgres/ has implementations
internal/service/        — business logic; each service depends only on repo interfaces
internal/handler/        — gin handlers; thin layer calling services
internal/middleware/      — JWT auth (Auth), admin check (RequireAdmin), CORS
internal/ws/             — WebSocket hub (gorilla/websocket); broadcasts to rooms keyed by eventID
pkg/glicko2/             — standalone Glicko2 rating implementation
migrations/              — numbered SQL migration files (golang-migrate format)
```

**Key service interactions:**
- `DraftService` orchestrates end-of-event promotion/relegation: reads finished event → reseeds players across groups → creates next event draft; depends on `MatchService`, `RatingService`, `GroupService`
- `MatchService` updates scores, recalculates group standings, broadcasts via `ws.Hub`
- `GroupService` generates round-robin schedules, calculates placements (handles three-way ties requiring manual umpire resolution)
- `RatingService` applies Glicko2 to group finishers; rating history stored per group finish

**Route groups:**
- `/api/v1/auth/*` — login/logout/register (Google OAuth + email)
- `/api/v1/public/*` — read-only, no auth
- `/api/v1/secured/*` — requires JWT cookie
- `/api/v1/admin/*` — requires JWT + `is_admin=true`
- `/ws/events/:eid` — WebSocket for live event updates

**Default config** (all overridable via env):
- `DATABASE_URL`: `postgres://test:test@localhost:5432/test?sslmode=disable`
- `PORT`: `8080`
- `FRONTEND_URL`: `http://localhost:5173`

## Frontend (league-ui)

### Commands

```bash
cd league-ui

# Dev server (proxies to localhost:8080)
npm run dev

# Type check (must run from league-ui/, not repo root)
npx tsc --noEmit

# Run unit tests
npx vitest

# Run single test file
npx vitest src/components/GroupStandings/GroupStandings.test.tsx

# Lint
npm run lint
```

### Architecture

```
src/api/         — axios API clients; client.ts sets baseURL to VITE_API_URL, redirects 401 on /secured/ routes
src/hooks/       — data-fetching hooks (useLeagues, useGroups, useMatches, etc.) + useWebSocket
src/context/     — AuthContext (current user)
src/pages/       — route-level components
src/components/  — reusable UI (each has co-located .test.tsx)
src/router/      — React Router v7 browser router + RootLayout
src/types/       — shared TypeScript types
```

**WebSocket**: `useWebSocket` connects to `ws://…/ws/events/:eid`; `LiveViewPage` uses it for real-time match score updates.

**API base URL**: set via `VITE_API_URL` env var; defaults to `/api/v1`. Dev script hardcodes `http://localhost:8080/api/v1`.

**Testing**: Vitest + jsdom + @testing-library/react. Setup file at `src/setup.ts`. No Cypress integration tests are wired into CI.
