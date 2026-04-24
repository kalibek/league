import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ScoreEntryForm } from './ScoreEntryForm'
import type { Match } from '../../types'

const match: Match = {
  matchId: 1,
  groupId: 1,
  groupPlayer1Id: 1,
  groupPlayer2Id: 2,
  score1: null,
  score2: null,
  withdraw1: false,
  withdraw2: false,
  status: 'DRAFT',
}

describe('ScoreEntryForm', () => {
  it('renders score inputs and buttons', () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByLabelText(/Player 1 games/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Player 2 games/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /save score/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('calls onSubmit with valid scores', async () => {
    const onSubmit = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={onSubmit} onClose={vi.fn()} />)

    await userEvent.clear(screen.getByLabelText(/Player 1 games/i))
    await userEvent.type(screen.getByLabelText(/Player 1 games/i), '3')
    await userEvent.clear(screen.getByLabelText(/Player 2 games/i))
    await userEvent.type(screen.getByLabelText(/Player 2 games/i), '1')

    await userEvent.click(screen.getByRole('button', { name: /save score/i }))
    expect(onSubmit).toHaveBeenCalledWith(3, 1)
  })

  it('shows error for invalid scores (neither reaches gamesToWin)', async () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} />)

    await userEvent.clear(screen.getByLabelText(/Player 1 games/i))
    await userEvent.type(screen.getByLabelText(/Player 1 games/i), '2')
    await userEvent.clear(screen.getByLabelText(/Player 2 games/i))
    await userEvent.type(screen.getByLabelText(/Player 2 games/i), '1')

    await userEvent.click(screen.getByRole('button', { name: /save score/i }))
    expect(screen.getByText(/one score must equal 3/i)).toBeInTheDocument()
  })

  it('calls onClose when cancel is clicked', async () => {
    const onClose = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={onClose} />)
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })
})
