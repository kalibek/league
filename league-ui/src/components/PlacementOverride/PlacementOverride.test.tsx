import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PlacementOverride } from './PlacementOverride'
import type { GroupPlayer } from '../../types'

const makePlayer = (id: number, name: string): GroupPlayer => ({
  groupPlayerId: id,
  groupId: 1,
  userId: id,
  seed: id,
  place: 0,
  points: 4,
  tiebreakPoints: 1,
  advances: false,
  recedes: false,
  isNonCalculated: false,
  user: { userId: id, firstName: name, lastName: 'X', email: `${id}@t.com`, currentRating: 1500, deviation: 200, volatility: 0.06, isAdmin: false },
})

describe('PlacementOverride', () => {
  it('renders all players', () => {
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    render(<PlacementOverride players={players} onConfirm={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText(/Alice/)).toBeInTheDocument()
    expect(screen.getByText(/Bob/)).toBeInTheDocument()
  })

  it('calls onConfirm with ordered player ids', async () => {
    const onConfirm = vi.fn()
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    render(<PlacementOverride players={players} onConfirm={onConfirm} onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: /confirm placement/i }))
    expect(onConfirm).toHaveBeenCalledWith([1, 2])
  })

  it('calls onClose when cancel is clicked', async () => {
    const onClose = vi.fn()
    render(<PlacementOverride players={[makePlayer(1, 'A')]} onConfirm={vi.fn()} onClose={onClose} />)
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('reorders via move up/down buttons', async () => {
    const onConfirm = vi.fn()
    const players = [makePlayer(1, 'Alice'), makePlayer(2, 'Bob')]
    render(<PlacementOverride players={players} onConfirm={onConfirm} onClose={vi.fn()} />)

    // Move Bob up
    const moveUpBtns = screen.getAllByLabelText('Move up')
    await userEvent.click(moveUpBtns[1]) // second player's move-up

    await userEvent.click(screen.getByRole('button', { name: /confirm placement/i }))
    expect(onConfirm).toHaveBeenCalledWith([2, 1])
  })
})
