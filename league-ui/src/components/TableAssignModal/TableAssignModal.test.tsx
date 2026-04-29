import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TableAssignModal } from './TableAssignModal'
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

describe('TableAssignModal', () => {
  it('renders a select with available tables', () => {
    render(
      <TableAssignModal
        match={match}
        numberOfTables={5}
        tablesInUse={[2, 4]}
        onSubmit={vi.fn()}
        onClose={vi.fn()}
      />
    )
    const options = screen.getAllByRole('option')
    const values = options.map((o) => (o as HTMLOptionElement).value)
    expect(values).toContain('1')
    expect(values).toContain('3')
    expect(values).toContain('5')
    expect(values).not.toContain('2')
    expect(values).not.toContain('4')
  })

  it('calls onSubmit with selected table number when submitted', async () => {
    const onSubmit = vi.fn()
    render(
      <TableAssignModal
        match={match}
        numberOfTables={3}
        tablesInUse={[]}
        onSubmit={onSubmit}
        onClose={vi.fn()}
      />
    )
    await userEvent.click(screen.getByRole('button', { name: /assign/i }))
    expect(onSubmit).toHaveBeenCalledWith(1)
  })

  it('calls onClose when cancel is clicked', async () => {
    const onClose = vi.fn()
    render(
      <TableAssignModal
        match={match}
        numberOfTables={3}
        tablesInUse={[]}
        onSubmit={vi.fn()}
        onClose={onClose}
      />
    )
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('shows occupied message when all tables are in use', () => {
    render(
      <TableAssignModal
        match={match}
        numberOfTables={2}
        tablesInUse={[1, 2]}
        onSubmit={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.queryByRole('combobox')).toBeNull()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('disables submit button when loading', () => {
    render(
      <TableAssignModal
        match={match}
        numberOfTables={3}
        tablesInUse={[]}
        onSubmit={vi.fn()}
        onClose={vi.fn()}
        loading={true}
      />
    )
    const assignBtn = screen.getByRole('button', { name: /assign/i })
    expect(assignBtn).toBeDisabled()
  })
})
