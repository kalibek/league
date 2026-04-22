# Frontend Implementation Plan — Table Tennis League UI

## Stack
- React 19 + TypeScript (Vite, already scaffolded)
- React Router v6 in data mode
- Tailwind CSS v4
- Axios for API calls
- Vitest + React Testing Library for unit tests
- Cypress for integration tests
- ESLint + Prettier

---

## Confirmed Decisions
| # | Decision |
|---|---------|
| A1 | OAuth handled server-side; frontend shows Google/Facebook/Apple buttons, each redirecting to `/auth/login?provider=google|facebook|apple` |
| A2 | No global state library (no Redux/Zustand); Auth state via Context, server state via custom hooks |
| A3 | React Router loaders used for data fetching on route entry |
| A4 | WebSocket connection is per-event, created only on LiveViewPage |
| A5 | Tailwind v4 (`@tailwindcss/vite` plugin, no `tailwind.config.ts` needed) |
| A6 | DNS status derived from match withdraw flags; no `didNotShow` field on GroupPlayer |
| A7 | CSV import accepts: `first_name`, `last_name`, `email`, `initial_rating` (optional) |

---

## Packages to Install

```bash
# Runtime
npm install react-router-dom axios

# Tailwind (v4 uses Vite plugin, no config file)
npm install tailwindcss @tailwindcss/vite

# Testing
npm install -D vitest @vitest/coverage-v8 @testing-library/react @testing-library/user-event jsdom

# Cypress
npm install -D cypress

# Code quality
npm install -D prettier eslint-config-prettier
```

---

## Directory Structure

```
league-ui/src/
├── api/                          # Typed async functions — one file per resource
│   ├── client.ts                 # Axios instance: base URL, interceptors (401 → /login)
│   ├── auth.ts                   # getMe
│   ├── players.ts                # listPlayers, getPlayer, createPlayer, importCSV
│   ├── leagues.ts                # listLeagues, getLeague, createLeague, updateConfig, assignRole
│   ├── events.ts                 # listEvents, getEvent, createDraft, updateEventConfig, startEvent
│   ├── groups.ts                 # getGroup, finishGroup, addPlayer, markNoShow, setManualPlace
│   └── matches.ts                # updateMatchScore
├── hooks/                        # React hooks — wrap api functions, manage loading/error state
│   ├── useAuth.ts                # reads AuthContext
│   ├── usePlayers.ts
│   ├── useLeagues.ts
│   ├── useEvents.ts
│   ├── useGroups.ts
│   ├── useMatches.ts
│   └── useWebSocket.ts           # WebSocket subscription for live event updates
├── context/
│   └── AuthContext.tsx           # AuthProvider: current user, role helpers, logout
├── components/                   # Atomic, reusable UI components
│   ├── Button/
│   │   ├── Button.tsx
│   │   └── Button.test.tsx
│   ├── Input/
│   │   ├── Input.tsx
│   │   └── Input.test.tsx
│   ├── Select/
│   │   ├── Select.tsx
│   │   └── Select.test.tsx
│   ├── Modal/
│   │   ├── Modal.tsx
│   │   └── Modal.test.tsx
│   ├── Table/
│   │   ├── Table.tsx             # generic sortable table
│   │   └── Table.test.tsx
│   ├── Badge/
│   │   ├── Badge.tsx             # status badges (DRAFT, IN_PROGRESS, DONE)
│   │   └── Badge.test.tsx
│   ├── RatingDelta/
│   │   ├── RatingDelta.tsx       # +12 in green / -8 in red
│   │   └── RatingDelta.test.tsx
│   ├── PlayerCard/
│   │   ├── PlayerCard.tsx        # name + rating summary row
│   │   └── PlayerCard.test.tsx
│   ├── GroupStandings/
│   │   ├── GroupStandings.tsx    # standings table: place, name, pts, tiebreak, advances/recedes
│   │   └── GroupStandings.test.tsx
│   ├── MatchGrid/
│   │   ├── MatchGrid.tsx         # n×n round-robin grid; cells show score or are clickable
│   │   └── MatchGrid.test.tsx
│   ├── ScoreEntryForm/
│   │   ├── ScoreEntryForm.tsx    # score input modal for umpires
│   │   └── ScoreEntryForm.test.tsx
│   ├── GroupCard/
│   │   ├── GroupCard.tsx         # card wrapping GroupStandings + MatchGrid for live view
│   │   └── GroupCard.test.tsx
│   ├── LeagueConfigForm/
│   │   ├── LeagueConfigForm.tsx  # numberOfAdvances, numberOfRecedes, gamesToWin, groupSize
│   │   └── LeagueConfigForm.test.tsx
│   ├── CSVImport/
│   │   ├── CSVImport.tsx         # drag-and-drop file upload + preview + submit
│   │   └── CSVImport.test.tsx
│   └── PlacementOverride/
│       ├── PlacementOverride.tsx # drag-to-reorder for manual tiebreak resolution
│       └── PlacementOverride.test.tsx
├── pages/                        # Route-level components
│   ├── HomePage.tsx              # active leagues list + quick links
│   ├── LoginPage.tsx             # OAuth sign-in button
│   ├── PlayersPage.tsx           # top players table with sort/filter
│   ├── PlayerProfilePage.tsx     # player detail: rating chart, group history, match history
│   ├── PlayerCreatePage.tsx      # manual create form
│   ├── PlayerImportPage.tsx      # CSV import flow
│   ├── LeaguesPage.tsx           # all leagues list
│   ├── LeaguePage.tsx            # league detail: events list, roles management
│   ├── LeagueConfigPage.tsx      # edit league config (maintainer)
│   └── LiveViewPage.tsx          # live view: grid of all groups in an event (most complex)
├── router/
│   └── index.tsx                 # React Router route definitions with loaders/guards
├── types/
│   └── index.ts                  # TypeScript interfaces mirroring backend models
└── main.tsx
```

---

## Types (src/types/index.ts)

```typescript
export interface User {
  userId: number
  firstName: string
  lastName: string
  email: string
  currentRating: number
  deviation: number
  volatility: number
}

export interface LeagueConfig {
  numberOfAdvances: number
  numberOfRecedes: number
  gamesToWin: number
  groupSize: number
}

export interface League {
  leagueId: number
  title: string
  description: string
  configuration: LeagueConfig
  created: string
  lastUpdated: string
}

export type EventStatus = 'DRAFT' | 'IN_PROGRESS' | 'DONE'
export type GroupStatus = 'DRAFT' | 'IN_PROGRESS' | 'DONE'
export type MatchStatus = 'DRAFT' | 'IN_PROGRESS' | 'DONE'

export interface LeagueEvent {
  eventId: number
  leagueId: number
  status: EventStatus
  title: string
  startDate: string
  endDate: string
}

export interface Group {
  groupId: number
  eventId: number
  status: GroupStatus
  division: string   // 'Superleague' | 'A' | 'B' | ...
  groupNo: number
  scheduled: string
}

export interface GroupPlayer {
  groupPlayerId: number
  groupId: number
  userId: number
  seed: number
  place: number
  points: number
  tiebreakPoints: number
  advances: boolean
  recedes: boolean
  isNonCalculated: boolean
  // DNS status is not a field. A player is DNS when all their matches
  // have withdraw1 or withdraw2 = true for their position.
  // Use isDns(groupPlayerId, matches) helper to derive it.
  user?: User  // populated in detail views
}

export interface Match {
  matchId: number
  groupId: number
  groupPlayer1Id: number | null
  groupPlayer2Id: number | null
  score1: number | null
  score2: number | null
  withdraw1: boolean
  withdraw2: boolean
  status: MatchStatus
}

export interface RatingHistory {
  historyId: number
  userId: number
  matchId: number
  delta: number
  rating: number
  deviation: number
  volatility: number
}

export interface GroupDetail extends Group {
  players: GroupPlayer[]
  matches: Match[]
}

export interface EventDetail extends LeagueEvent {
  groups: GroupDetail[]
}

// WebSocket message types
export type WSMessageType = 
  | 'match_updated' 
  | 'group_finished' 
  | 'event_finished' 
  | 'manual_placement_required'

export interface WSMessage {
  type: WSMessageType
  groupId: number
  matchId?: number
  payload: unknown
}

export interface UserRole {
  userId: number
  leagueId: number
  roleName: 'player' | 'umpire' | 'maintainer'
}

// Helper: derive DNS status from match withdraw flags
export function isDns(groupPlayerId: number, matches: Match[]): boolean {
  const playerMatches = matches.filter(
    m => m.groupPlayer1Id === groupPlayerId || m.groupPlayer2Id === groupPlayerId
  )
  if (playerMatches.length === 0) return false
  return playerMatches.every(m =>
    m.groupPlayer1Id === groupPlayerId ? m.withdraw1 : m.withdraw2
  )
}
```

---

## API Layer (src/api/)

### client.ts
```typescript
import axios from 'axios'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? '/api/v1',
  withCredentials: true,  // for session cookies
})

// redirect to login on 401
client.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) window.location.href = '/login'
    return Promise.reject(err)
  }
)

export default client
```

### players.ts
```typescript
export const listPlayers = (params: { sort?: string; limit?: number; offset?: number }) =>
  client.get<User[]>('/players', { params })

export const getPlayer = (id: number) =>
  client.get<User & { ratingHistory: RatingHistory[]; groups: GroupDetail[] }>(`/players/${id}`)

export const createPlayer = (data: { firstName: string; lastName: string; email: string }) =>
  client.post<User>('/players', data)

export const importCSV = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return client.post<{ imported: number; skipped: number; errors: string[] }>('/players/import', form)
}
```

### leagues.ts, events.ts, groups.ts, matches.ts
Follow same pattern: typed functions using `client.get/post/put/delete`.

---

## Hooks Layer (src/hooks/)

Each hook wraps its API module and provides loading/error state. No caching library — React state only.

```typescript
// Example: hooks/useGroups.ts
export function useFinishGroup() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const finish = async (eventId: number, groupId: number) => {
    setLoading(true)
    setError(null)
    try {
      await finishGroup(eventId, groupId)
    } catch (e) {
      setError(extractErrorMessage(e))
    } finally {
      setLoading(false)
    }
  }

  return { finish, loading, error }
}
```

### useWebSocket.ts
```typescript
export function useEventWebSocket(eventId: number, onMessage: (msg: WSMessage) => void) {
  useEffect(() => {
    const ws = new WebSocket(`${WS_BASE_URL}/ws/events/${eventId}`)
    ws.onmessage = e => onMessage(JSON.parse(e.data) as WSMessage)
    ws.onerror = () => console.error('WebSocket error')
    return () => ws.close()
  }, [eventId])
}
```

---

## Auth Context (src/context/AuthContext.tsx)

```typescript
interface AuthContextValue {
  user: User | null
  roles: Record<number, string[]>  // leagueId → roles
  isMaintainer: (leagueId: number) => boolean
  isUmpire: (leagueId: number) => boolean
  logout: () => void
}
```

- Fetches `/auth/me` on app mount
- If 401 and not on `/login`, redirects to `/login`
- Provides `isMaintainer(leagueId)` and `isUmpire(leagueId)` helpers

---

## Router (src/router/index.tsx)

Using React Router v6 `createBrowserRouter` in data mode:

```typescript
const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,  // AuthProvider wrapper, nav bar
    children: [
      { index: true, element: <HomePage /> },
      { path: 'players', element: <PlayersPage /> },
      { path: 'players/new', element: <PlayerCreatePage /> },
      { path: 'players/import', element: <PlayerImportPage /> },
      { path: 'players/:id', element: <PlayerProfilePage /> },
      { path: 'leagues', element: <LeaguesPage /> },
      { path: 'leagues/new', element: <CreateLeaguePage /> },
      { path: 'leagues/:id', element: <LeaguePage /> },
      { path: 'leagues/:id/config', element: <LeagueConfigPage /> },
      { path: 'leagues/:id/events/:eid', element: <LiveViewPage /> },
    ]
  },
  { path: '/login', element: <LoginPage /> },
])
```

Route guards: components check `AuthContext.user` and redirect if needed.

---

## Page Specifications

### LiveViewPage (most complex)
State: `EventDetail` loaded from `/api/v1/leagues/:id/events/:eid`.
WebSocket: `useEventWebSocket(eventId, handleWSMessage)`.

```
handleWSMessage(msg):
  if match_updated → update match in local state by matchId
  if group_finished → update group status, show placement badges  
  if event_finished → show "Create Next Draft" button (Maintainer only)
  if manual_placement_required → open PlacementOverride modal for that group
```

Layout:
```
<EventHeader status badge, title, dates>
<GroupGrid columns={2}>
  for each group:
    <GroupCard division="A" groupNo={1}>
      <GroupStandings players={...} />         // shows place, name, pts, tiebreak, arrows for advance/recede
      <MatchGrid players={...} matches={...} onCellClick={openScoreEntry} />
      // Umpire controls:
      <Button onClick={finishGroup}>Finish Group</Button>
      // Maintainer controls:
      <Button onClick={addPlayer}>Add Player</Button>
      // per-player row: <Button onClick={markNoShow}>No Show</Button>
    </GroupCard>
</GroupGrid>
```

#### MatchGrid layout
NxN grid where row player vs column player. Diagonal is empty. Each cell shows:
- `—` if match not played
- `3:1` if played (score1:score2 from row player's perspective)
- For umpires: cell is a `<button>` that opens ScoreEntryForm

#### GroupStandings columns
| Place | Name | Pts | TB | W | L | ↑↓ |
|-------|------|-----|-----|---|---|-----|
Advance indicator `↑` shown on `advances=true` rows, `↓` on `recedes=true`.
`isNonCalculated` rows shown in italic with `(guest)` label, no place shown.
DNS rows (detected via `isDns(groupPlayerId, matches)` helper): shown with strikethrough name and a "DNS" badge; automatically appear at the bottom due to 0 points and large negative tiebreak.

### PlayerProfilePage
Sections:
1. Header: name, current rating, deviation
2. Rating chart: sparkline of `rating` over time from `ratingHistory`
3. League history: table of each event/group participated in — division, place, points
4. Match history: each match with opponent name, score, rating delta

### PlayersPage
Filterable/sortable table:
- Columns: rank, name, rating, deviation, recent delta (last 3 events)
- Sort by: rating (default), name
- Filter by: division (last event), rating range

### LoginPage
Shows three OAuth provider buttons, each linking to the backend auth redirect:
- "Sign in with Google" → `/auth/login?provider=google`
- "Sign in with Facebook" → `/auth/login?provider=facebook`
- "Sign in with Apple" → `/auth/login?provider=apple`

All are standard anchor tags (not JS navigation) so the browser follows the redirect fully. After OAuth completes, the backend redirects back to `/` with a JWT in a cookie.

### LeagueConfigPage
Form fields: `numberOfAdvances`, `numberOfRecedes`, `gamesToWin`, `groupSize`.
Shows warning: "Changing config will recreate the draft if an event is in DRAFT status."
Submit triggers `PUT /api/v1/leagues/:id/config` and if draft exists, also `PUT /api/v1/leagues/:id/events/:eid/config`.

### PlayerImportPage
1. Drag-and-drop zone (accepts `.csv`)
2. CSV preview table (first 5 rows)
3. Column mapping UI: dropdowns to map CSV columns to `first_name`, `last_name`, `email`, `initial_rating`
   - `initial_rating` mapping is optional; unmapped rows use the server default of 1500
   - `deviation` and `volatility` are never shown — always use server defaults
4. "Import" button → calls `POST /api/v1/players/import`
5. Result summary: "Imported: 42, Skipped: 3 (duplicates by email), Errors: 1" with per-row error details

---

## Component Specifications

### Button
Props: `variant: 'primary' | 'secondary' | 'danger'`, `loading: boolean`, `disabled: boolean`, `onClick`, `type`, `children`.
Tailwind: primary = `bg-blue-600 hover:bg-blue-700 text-white`, danger = `bg-red-600...`.

### GroupStandings
Props: `players: GroupPlayer[]`, `matches: Match[]`.
Sorts by `place ASC` (0-placed players shown last).
Uses `advances`/`recedes` flags to show `↑`/`↓` indicators after draft is created.
Uses `isDns(p.groupPlayerId, matches)` to determine DNS styling (strikethrough + DNS badge).
`isNonCalculated` players shown in italic, no place number, no advance/recede indicator.

### MatchGrid
Props: `players: GroupPlayer[]`, `matches: Match[]`, `onScoreClick?: (match: Match) => void`.
Builds a lookup `{ [p1Id-p2Id]: Match }` for O(1) cell rendering.
Cell content: shows score from row player's perspective (swap if player is player2).

### ScoreEntryForm
Props: `match: Match`, `gamesToWin: number`, `onSubmit: (score1, score2) => void`, `onClose`.
Inputs: two number inputs (0 to gamesToWin), validates one must equal gamesToWin, other < gamesToWin.

### PlacementOverride
Props: `players: GroupPlayer[]` (the tied subset), `onConfirm: (orderedPlayerIds: number[]) => void`.
Renders a draggable list. On confirm, calls `PUT .../players/:pid/place` for each.

---

## Tailwind Setup

In `vite.config.ts`:
```typescript
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
})
```

In `src/index.css`:
```css
@import "tailwindcss";
```

---

## Testing Strategy

### Unit tests (Vitest)
- Location: sibling `.test.tsx` files in each component directory
- Setup: `vitest.config.ts` with `jsdom` environment, `@testing-library/react`
- Each component test covers: renders correctly, key user interactions, edge cases (empty state, loading state, error state)

Priority components to test first:
1. GroupStandings — placement logic display
2. MatchGrid — score display and cell clicks
3. ScoreEntryForm — validation rules
4. PlacementOverride — drag interaction

Example test pattern:
```typescript
// GroupStandings.test.tsx
describe('GroupStandings', () => {
  it('renders players sorted by place', () => { ... })
  it('shows advance indicator for advancing players', () => { ... })
  it('shows non-calculated players as guests without place', () => { ... })
  it('shows did-not-show players with strikethrough', () => { ... })
})
```

### Integration tests (Cypress)
Location: `cypress/e2e/`

Key flows to cover:
1. `umpire-score-entry.cy.ts` — log in as umpire, navigate to live view, enter score, verify standings update
2. `group-finish.cy.ts` — enter all scores, finish group, verify placements shown
3. `player-import.cy.ts` — upload CSV, verify imported players appear
4. `live-view-realtime.cy.ts` — two browser sessions; score entered in one appears in other without refresh

---

## Implementation Phases

### Phase 1: Setup (Day 1)
- Install all packages
- Configure Tailwind (vite plugin)
- Configure Vitest (`vitest.config.ts`, `src/setup.ts`)
- Configure Prettier + ESLint prettier compat
- Set up `src/types/index.ts`
- Set up `src/api/client.ts`

### Phase 2: Auth + Router skeleton (Day 1-2)
- `AuthContext.tsx` + `AuthProvider`
- `LoginPage` (OAuth redirect button)
- React Router setup with all routes (pages as placeholders)
- Auth guard pattern in `RootLayout`

### Phase 3: API + Hooks layer (Day 2-3)
- All `src/api/*.ts` files
- All `src/hooks/*.ts` files
- Unit test hooks with mocked axios

### Phase 4: Atomic components (Day 3-4)
- Button, Input, Select, Modal, Table, Badge, RatingDelta
- Each with tests

### Phase 5: Feature components (Day 4-5)
- GroupStandings, MatchGrid, ScoreEntryForm, GroupCard, PlacementOverride
- Each with tests

### Phase 6: Pages (Day 5-7)
- PlayersPage → PlayerProfilePage → PlayerCreatePage → PlayerImportPage
- LeaguesPage → LeaguePage → LeagueConfigPage
- LiveViewPage (last — most complex, depends on all components)

### Phase 7: WebSocket integration (Day 7)
- `useWebSocket.ts`
- Wire into LiveViewPage state updates

### Phase 8: Cypress tests (Day 8)
- Write key e2e flows
- Set up Cypress fixtures for test data
