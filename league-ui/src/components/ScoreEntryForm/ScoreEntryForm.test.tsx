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
  tableNumber: null,
}

describe('ScoreEntryForm', () => {
  it('renders score buttons for player1 wins', () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: '3 : 0' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '3 : 1' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '3 : 2' })).toBeInTheDocument()
  })

  it('renders score buttons for player2 wins', () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: '0 : 3' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '1 : 3' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '2 : 3' })).toBeInTheDocument()
  })

  it('calls onSubmit with correct args when a score button is clicked', async () => {
    const onSubmit = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={onSubmit} onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: '3 : 1' }))
    expect(onSubmit).toHaveBeenCalledWith(3, 1, false, false)
  })

  it('calls onSubmit with correct args for player2 win button', async () => {
    const onSubmit = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={onSubmit} onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: '1 : 3' }))
    expect(onSubmit).toHaveBeenCalledWith(1, 3, false, false)
  })

  it('renders cancel button', () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('calls onClose when cancel button is clicked', async () => {
    const onClose = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={onClose} />)
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('renders walkover W-L and L-W buttons', () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'W-L' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'L-W' })).toBeInTheDocument()
  })

  it('calls onSubmit with withdraw flags when W-L is clicked', async () => {
    const onSubmit = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={onSubmit} onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'W-L' }))
    expect(onSubmit).toHaveBeenCalledWith(3, 0, false, true)
  })

  it('calls onSubmit with withdraw flags when L-W is clicked', async () => {
    const onSubmit = vi.fn()
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={onSubmit} onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'L-W' }))
    expect(onSubmit).toHaveBeenCalledWith(0, 3, true, false)
  })

  it('disables all buttons when loading is true', () => {
    render(<ScoreEntryForm match={match} gamesToWin={3} onSubmit={vi.fn()} onClose={vi.fn()} loading={true} />)
    const scoreButton = screen.getByRole('button', { name: '3 : 0' })
    expect(scoreButton).toBeDisabled()
  })
})
