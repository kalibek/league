# CLAUDE.md

Guide Claude Code (claude.ai/code) for this repo.

## Project Overview

Full-stack table tennis league platform. Players compete in monthly round-robin leagues, divisions (A/B/C), groups. Matches feed Glicko2 rating. Month-end drafts promote/relegate players.

## Repository Structure

```
league/
├── league-ui/        # React 19 + Vite + TypeScript SPA
├── league-api/       # Go REST API + WebSocket (not yet implemented)
├── plan.md           # Business requirements, data model, scoring rules
├── backend-plan.md   # Go architecture, DB schema, confirmed decisions
└── frontend-plan.md  # React architecture, component specs, confirmed decisions
```

`plan.md`, `backend-plan.md`, `frontend-plan.md` = implementation roadmaps. Read before any arch decision.

## Frontend (league-ui)

**Stack**: React 19, Vite, TypeScript, Tailwind v4, Axios, React Router v7

**Commands** (from `league-ui/`):
```bash
npm run dev       # Start Vite dev server
npm run build     # tsc -b && vite build
npm run lint      # ESLint
npm run preview   # Preview production build
```

**Planned directory layout** (per `frontend-plan.md`):
- `src/api/` — Axios client and per-resource request functions
- `src/hooks/` — data-fetch hooks (no global state; Router loaders + local state)
- `src/context/` — Auth context only
- `src/components/` — shared UI components (Button, Input, Table, Badge, etc.)
- `src/pages/` — Login, Players, Leagues, LiveView, Profile, LeagueConfig
- `src/router/` — React Router config with loaders
- `src/types/` — TypeScript interfaces (User, League, Event, Group, Match, Rating)

**Key confirmed decisions** (from `frontend-plan.md`):
- OAuth server-side (no PKCE); frontend redirects to `/api/auth/{provider}`
- No global state — React Router loaders + local `useState`
- Tailwind v4 (not v3)

## Backend (league-api)

**Stack**: Go, Gin, PostgreSQL, sqlx, go-migrate, WebSocket (gorilla/websocket)

**Planned commands** (from `backend-plan.md`):
```bash
go build ./cmd/server   # Build server binary
go test ./...           # Run all tests
```

**Planned directory layout** (per `backend-plan.md`):
- `cmd/server/` — entry point
- `internal/config/` — env-var loading
- `internal/handler/` — Gin route handlers (HTTP + WebSocket)
- `internal/service/` — business logic (auth, player, league, group, match, draft, rating)
- `internal/repository/` — PostgreSQL queries via sqlx
- `internal/model/` — Go structs matching DB schema
- `migrations/` — SQL migration files (go-migrate)

**Key confirmed decisions** (from `backend-plan.md`):
- Three-layer: handler → service → repository (no logic in handlers or repositories)
- sqlx with raw SQL (no ORM)
- Glicko2 as custom internal package
- OAuth 2.0 via Google/Facebook/Apple; JWT sessions
- WebSocket hub for live match score broadcasting
- CSV import for bulk player registration

## Domain Rules (from plan.md)

- **Scoring**: Win = 2 pts, Loss = 1 pt; tiebreaks by head-to-head then point differential
- **Roles**: Player, Umpire, Maintainer (Maintainer manage everything)
- **Draft**: Monthly — top players advance, bottom recede between groups
- **Rating**: Glicko2 recalculated after each event; initial rating 1500, RD 350, volatility 0.06