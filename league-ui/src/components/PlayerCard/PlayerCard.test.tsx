import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PlayerCard } from './PlayerCard'
import type { User } from '../../types'

const player: User = {
  userId: 1,
  firstName: 'Alice',
  lastName: 'Smith',
  email: 'alice@example.com',
  currentRating: 1532.7,
  deviation: 45.3,
  volatility: 0.06,
}

describe('PlayerCard', () => {
  it('renders player name', () => {
    render(<PlayerCard player={player} />)
    expect(screen.getByText('Alice Smith')).toBeInTheDocument()
  })

  it('renders rounded rating', () => {
    render(<PlayerCard player={player} />)
    expect(screen.getByText('1533')).toBeInTheDocument()
  })

  it('renders email', () => {
    render(<PlayerCard player={player} />)
    expect(screen.getByText('alice@example.com')).toBeInTheDocument()
  })
})
