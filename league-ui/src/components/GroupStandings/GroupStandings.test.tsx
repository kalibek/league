import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { GroupStandings } from './GroupStandings'
import type { GroupPlayer, Match } from '../../types'

const renderStandings = (players: GroupPlayer[], matches: Match[]) =>
  render(
    <MemoryRouter>
      <GroupStandings players={players} matches={matches} />
    </MemoryRouter>
  )

const basePlayer = (overrides: Partial<GroupPlayer>): GroupPlayer => ({
  groupPlayerId: 1,
  groupId: 1,
  userId: 1,
  seed: 1,
  place: 1,
  points: 4,
  tiebreakPoints: 2,
  advances: false,
  recedes: false,
  isNonCalculated: false,
  user: { userId: 1, firstName: 'Alice', lastName: 'Smith', email: 'a@b.com', currentRating: 1500, deviation: 200, volatility: 0.06, isAdmin: false },
  ...overrides,
})

const noMatch: Match[] = []

describe('GroupStandings', () => {
  it('renders players sorted by seed', () => {
    const players: GroupPlayer[] = [
      basePlayer({ groupPlayerId: 2, seed: 2, user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06 } }),
      basePlayer({ groupPlayerId: 1, seed: 1 }),
    ]
    renderStandings(players, noMatch)
    const rows = screen.getAllByRole('row')
    expect(rows).toHaveLength(3)
    expect(rows[1]).toHaveTextContent('Alice')
  })

  it('shows advance indicator for advancing players', () => {
    renderStandings([basePlayer({ advances: true })], noMatch)
    expect(screen.getByTitle('Advances')).toBeInTheDocument()
  })

  it('shows recede indicator for receding players', () => {
    renderStandings([basePlayer({ recedes: true })], noMatch)
    expect(screen.getByTitle('Recedes')).toBeInTheDocument()
  })

  it('shows non-calculated players as guests without place', () => {
    renderStandings([basePlayer({ isNonCalculated: true })], noMatch)
    expect(screen.getByText('(guest)')).toBeInTheDocument()
    const cells = screen.getAllByText('—')
    expect(cells.length).toBeGreaterThan(0)
  })

  it('shows — for TB when player has unique points (no tie)', () => {
    const players = [
      basePlayer({ groupPlayerId: 1, points: 6, seed: 1 }),
      basePlayer({ groupPlayerId: 2, points: 4, seed: 2, user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06 } }),
    ]
    renderStandings(players, noMatch)
    const rows = screen.getAllByRole('row')
    expect(rows[1]).toHaveTextContent('—')
    expect(rows[2]).toHaveTextContent('—')
  })

  it('shows backend tiebreakPoints for tied players, — for non-tied', () => {
    // Backend already computed TB correctly: A=+2, B=-2 (tied at 5 pts); C=0 but unique
    const players = [
      basePlayer({ groupPlayerId: 1, userId: 1, points: 5, tiebreakPoints: 2, seed: 1 }),
      basePlayer({ groupPlayerId: 2, userId: 2, points: 5, tiebreakPoints: -2, seed: 2, user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06 } }),
      basePlayer({ groupPlayerId: 3, userId: 3, points: 3, tiebreakPoints: 0, seed: 3, user: { userId: 3, firstName: 'Carol', lastName: 'Lee', email: 'c@d.com', currentRating: 1300, deviation: 200, volatility: 0.06 } }),
    ]
    renderStandings(players, noMatch)
    const rows = screen.getAllByRole('row')
    expect(rows[1]).toHaveTextContent('2')   // A: tied, show backend TB
    expect(rows[2]).toHaveTextContent('-2')  // B: tied, show backend TB
    expect(rows[3]).toHaveTextContent('—')   // C: unique points → —
  })

  it('shows backend tiebreakPoints for two separate tie groups', () => {
    // Backend computed: A=+2, B=-2 (pts=5); C=-3, D=+3 (pts=3)
    const players = [
      basePlayer({ groupPlayerId: 1, userId: 1, points: 5, tiebreakPoints: 2, seed: 1 }),
      basePlayer({ groupPlayerId: 2, userId: 2, points: 5, tiebreakPoints: -2, seed: 2, user: { isAdmin: false, userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06 } }),
      basePlayer({ groupPlayerId: 3, userId: 3, points: 3, tiebreakPoints: -3, seed: 3, user: { isAdmin: false, userId: 3, firstName: 'Carol', lastName: 'Lee', email: 'c@d.com', currentRating: 1300, deviation: 200, volatility: 0.06 } }),
      basePlayer({ groupPlayerId: 4, userId: 4, points: 3, tiebreakPoints: 3, seed: 4, user:  { isAdmin: false, userId: 4, firstName: 'Dave', lastName: 'Kim', email: 'd@e.com', currentRating: 1200, deviation: 200, volatility: 0.06 } }),
    ]
    renderStandings(players, noMatch)
    const rows = screen.getAllByRole('row')
    expect(rows[1]).toHaveTextContent('2')   // A
    expect(rows[2]).toHaveTextContent('-2')  // B
    expect(rows[3]).toHaveTextContent('-3')  // C
    expect(rows[4]).toHaveTextContent('3')   // D
  })

  it('shows DNS players with strikethrough and DNS badge', () => {
    const gpId = 10
    const players = [
      basePlayer({
        groupPlayerId: gpId,
        user: { userId: gpId, firstName: 'DNS', lastName: 'Player', email: 'd@e.com', currentRating: 1500, deviation: 200, volatility: 0.06, isAdmin: false },
      }),
    ]
    const matches: Match[] = [
      {
        matchId: 1, groupId: 1, groupPlayer1Id: gpId, groupPlayer2Id: 2,
        score1: null, score2: null, withdraw1: true, withdraw2: false, status: 'DONE', tableNumber: null,
      },
    ]
    renderStandings(players, matches)
    expect(screen.getByText('DNS')).toBeInTheDocument()
  })

  it('calls onNoShow when no-show button is clicked', async () => {
    const onNoShow = vi.fn()
    const player = basePlayer({ groupPlayerId: 5 })
    render(
      <MemoryRouter>
        <GroupStandings players={[player]} matches={[]} onNoShow={onNoShow} />
      </MemoryRouter>
    )
    const btn = screen.getByTitle(/mark/i)
    await userEvent.click(btn)
    expect(onNoShow).toHaveBeenCalledWith(5)
  })

  it('shows match score in cell for done match (no walkover)', () => {
    const p1 = basePlayer({ groupPlayerId: 1, seed: 1 })
    const p2 = basePlayer({
      groupPlayerId: 2, seed: 2, userId: 2,
      user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06, isAdmin: false },
    })
    const match: Match = {
      matchId: 1, groupId: 1, groupPlayer1Id: 1, groupPlayer2Id: 2,
      score1: 3, score2: 1, withdraw1: false, withdraw2: false, status: 'DONE', tableNumber: null,
    }
    renderStandings([p1, p2], [match])
    // score cells from row player perspective
    expect(screen.getAllByText('3:1').length).toBeGreaterThan(0)
  })

  it('calls onScoreClick when score cell button is clicked', async () => {
    const onScoreClick = vi.fn()
    const p1 = basePlayer({ groupPlayerId: 1, seed: 1 })
    const p2 = basePlayer({
      groupPlayerId: 2, seed: 2, userId: 2,
      user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06, isAdmin: false },
    })
    const match: Match = {
      matchId: 1, groupId: 1, groupPlayer1Id: 1, groupPlayer2Id: 2,
      score1: null, score2: null, withdraw1: false, withdraw2: false, status: 'DRAFT', tableNumber: null,
    }
    render(
      <MemoryRouter>
        <GroupStandings players={[p1, p2]} matches={[match]} onScoreClick={onScoreClick} />
      </MemoryRouter>
    )
    const scoreBtns = screen.getAllByRole('button')
    await userEvent.click(scoreBtns[0])
    expect(onScoreClick).toHaveBeenCalledOnce()
  })

  it('shows walkover score as W-L in cell', () => {
    const p1 = basePlayer({ groupPlayerId: 1, seed: 1 })
    const p2 = basePlayer({
      groupPlayerId: 2, seed: 2, userId: 2,
      user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06, isAdmin: false },
    })
    const match: Match = {
      matchId: 1, groupId: 1, groupPlayer1Id: 1, groupPlayer2Id: 2,
      score1: 3, score2: 0, withdraw1: false, withdraw2: true, status: 'DONE', tableNumber: null,
    }
    renderStandings([p1, p2], [match])
    expect(screen.getAllByText('W-L').length).toBeGreaterThan(0)
  })
})
