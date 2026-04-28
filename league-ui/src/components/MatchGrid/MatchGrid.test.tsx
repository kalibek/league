import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { MatchGrid } from './MatchGrid'
import type { GroupPlayer, Match } from '../../types'

const makePlayer = (id: number, firstName: string): GroupPlayer => ({
  groupPlayerId: id,
  groupId: 1,
  userId: id,
  seed: id,
  place: 0,
  points: 0,
  tiebreakPoints: 0,
  advances: false,
  recedes: false,
  isNonCalculated: false,
  user: { userId: id, firstName, lastName: 'X', email: `${id}@test.com`, currentRating: 1500, deviation: 200, volatility: 0.06, isAdmin: false },
})

const doneMatch = (p1: number, p2: number, s1: number, s2: number): Match => ({
  matchId: p1 * 10 + p2,
  groupId: 1,
  groupPlayer1Id: p1,
  groupPlayer2Id: p2,
  score1: s1,
  score2: s2,
  withdraw1: false,
  withdraw2: false,
  status: 'DONE',
})

const renderGrid = (players: GroupPlayer[], matches: Match[], onScoreClick?: (m: Match) => void) =>
  render(
    <MemoryRouter>
      <MatchGrid players={players} matches={matches} onScoreClick={onScoreClick} />
    </MemoryRouter>
  )

describe('MatchGrid', () => {
  it('renders player names in round match rows', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    renderGrid(players, [doneMatch(1, 2, 3, 1)])
    expect(screen.getAllByText('Alice X').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Bob X').length).toBeGreaterThan(0)
  })

  it('shows score pill for a done match', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    const matches = [doneMatch(1, 2, 3, 1)]
    renderGrid(players, matches)
    expect(screen.getByText('3 : 1')).toBeInTheDocument()
  })

  it('shows — for unplayed matches', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    const pendingMatch: Match = {
      matchId: 12,
      groupId: 1,
      groupPlayer1Id: 1,
      groupPlayer2Id: 2,
      score1: null,
      score2: null,
      withdraw1: false,
      withdraw2: false,
      status: 'DRAFT',
    }
    renderGrid(players, [pendingMatch])
    const dashes = screen.getAllByText('—')
    expect(dashes.length).toBeGreaterThan(0)
  })

  it('calls onScoreClick when score pill button is clicked', async () => {
    const handler = vi.fn()
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    const matches: Match[] = [
      {
        matchId: 12,
        groupId: 1,
        groupPlayer1Id: 1,
        groupPlayer2Id: 2,
        score1: null,
        score2: null,
        withdraw1: false,
        withdraw2: false,
        status: 'DRAFT',
      },
    ]
    renderGrid(players, matches, handler)
    const btns = screen.getAllByRole('button')
    await userEvent.click(btns[0])
    expect(handler).toHaveBeenCalledOnce()
  })

  it('does not show non-calculated players', () => {
    const players = [
      makePlayer(1, 'Alice'),
      { ...makePlayer(2, 'Guest'), isNonCalculated: true },
    ]
    renderGrid(players, [])
    const guestLinks = screen.queryAllByText('Guest X')
    expect(guestLinks.length).toBe(0)
  })

  it('renders round header labels', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    renderGrid(players, [doneMatch(1, 2, 3, 0)])
    // The round badge "1" should be present in the round header
    const roundLabels = screen.getAllByText('1')
    expect(roundLabels.length).toBeGreaterThan(0)
  })
})
