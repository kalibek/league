import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LeagueConfigForm } from './LeagueConfigForm'
import type { LeagueConfig } from '../../types'

const config: LeagueConfig = {
  numberOfAdvances: 2,
  numberOfRecedes: 2,
  gamesToWin: 3,
  groupSize: 6,
}

describe('LeagueConfigForm', () => {
  it('renders all config fields', () => {
    render(<LeagueConfigForm initial={config} onSubmit={vi.fn()} />)
    expect(screen.getByLabelText(/advances per group/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/recedes per group/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/games to win/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/group size/i)).toBeInTheDocument()
  })

  it('shows draft warning when showDraftWarning=true', () => {
    render(<LeagueConfigForm initial={config} onSubmit={vi.fn()} showDraftWarning />)
    expect(screen.getByText(/changing config will recreate/i)).toBeInTheDocument()
  })

  it('calls onSubmit with updated config', async () => {
    const onSubmit = vi.fn()
    render(<LeagueConfigForm initial={config} onSubmit={onSubmit} />)

    const groupSizeInput = screen.getByLabelText(/group size/i)
    await userEvent.clear(groupSizeInput)
    await userEvent.type(groupSizeInput, '8')

    await userEvent.click(screen.getByRole('button', { name: /save configuration/i }))
    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ groupSize: 8 }))
  })
})
