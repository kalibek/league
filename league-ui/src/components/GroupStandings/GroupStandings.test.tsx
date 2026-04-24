import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { GroupStandings } from './GroupStandings'
import type { GroupPlayer, Match } from '../../types'

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
  user: { userId: 1, firstName: 'Alice', lastName: 'Smith', email: 'a@b.com', currentRating: 1500, deviation: 200, volatility: 0.06 },
  ...overrides,
})

const noMatch: Match[] = []

describe('GroupStandings', () => {
  it('renders players sorted by place', () => {
    const players: GroupPlayer[] = [
      basePlayer({ groupPlayerId: 2, place: 2, user: { userId: 2, firstName: 'Bob', lastName: 'Jones', email: 'b@c.com', currentRating: 1400, deviation: 200, volatility: 0.06 } }),
      basePlayer({ groupPlayerId: 1, place: 1 }),
    ]
    render(<GroupStandings players={players} matches={noMatch} />)
    const rows = screen.getAllByRole('row')
    // header + 2 data rows
    expect(rows).toHaveLength(3)
    // First data row should be Alice (place 1)
    expect(rows[1]).toHaveTextContent('Alice')
  })

  it('shows advance indicator for advancing players', () => {
    const players = [basePlayer({ advances: true })]
    render(<GroupStandings players={players} matches={noMatch} />)
    expect(screen.getByTitle('Advances')).toBeInTheDocument()
  })

  it('shows recede indicator for receding players', () => {
    const players = [basePlayer({ recedes: true })]
    render(<GroupStandings players={players} matches={noMatch} />)
    expect(screen.getByTitle('Recedes')).toBeInTheDocument()
  })

  it('shows non-calculated players as guests without place', () => {
    const players = [basePlayer({ isNonCalculated: true })]
    render(<GroupStandings players={players} matches={noMatch} />)
    expect(screen.getByText('(guest)')).toBeInTheDocument()
    // Place column should be —
    const cells = screen.getAllByText('—')
    expect(cells.length).toBeGreaterThan(0)
  })

  it('shows DNS players with strikethrough and DNS badge', () => {
    const gpId = 10
    const players = [
      basePlayer({
        groupPlayerId: gpId,
        user: { userId: gpId, firstName: 'DNS', lastName: 'Player', email: 'd@e.com', currentRating: 1500, deviation: 200, volatility: 0.06 },
      }),
    ]
    const matches: Match[] = [
      {
        matchId: 1,
        groupId: 1,
        groupPlayer1Id: gpId,
        groupPlayer2Id: 2,
        score1: null,
        score2: null,
        withdraw1: true,
        withdraw2: false,
        status: 'DONE',
      },
    ]
    render(<GroupStandings players={players} matches={matches} />)
    expect(screen.getByText('DNS')).toBeInTheDocument()
  })
})
