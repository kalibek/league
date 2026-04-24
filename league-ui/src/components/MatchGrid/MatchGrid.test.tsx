import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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
  user: { userId: id, firstName, lastName: 'X', email: `${id}@test.com`, currentRating: 1500, deviation: 200, volatility: 0.06 },
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

describe('MatchGrid', () => {
  it('renders player names in header', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    render(<MatchGrid players={players} matches={[]} />)
    expect(screen.getAllByText('Alice').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Bob').length).toBeGreaterThan(0)
  })

  it('shows score from row player perspective', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    const matches = [doneMatch(1, 2, 3, 1)]
    render(<MatchGrid players={players} matches={matches} />)
    // Alice row vs Bob column: Alice won 3:1
    expect(screen.getByText('3:1')).toBeInTheDocument()
    // Bob row vs Alice column: Bob lost 1:3
    expect(screen.getByText('1:3')).toBeInTheDocument()
  })

  it('shows — for unplayed matches', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    render(<MatchGrid players={players} matches={[]} />)
    const dashes = screen.getAllByText('—')
    expect(dashes.length).toBeGreaterThan(0)
  })

  it('calls onScoreClick when umpire clicks a cell', async () => {
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
    render(<MatchGrid players={players} matches={matches} onScoreClick={handler} />)
    const btns = screen.getAllByRole('button')
    await userEvent.click(btns[0])
    expect(handler).toHaveBeenCalledOnce()
  })

  it('does not show non-calculated players', () => {
    const players = [
      makePlayer(1, 'Alice'),
      { ...makePlayer(2, 'Guest'), isNonCalculated: true },
    ]
    render(<MatchGrid players={players} matches={[]} />)
    const headers = screen.queryAllByText('Guest')
    // Guest should not appear as a column header
    expect(headers.length).toBe(0)
  })
})
