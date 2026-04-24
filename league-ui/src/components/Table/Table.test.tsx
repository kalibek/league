import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Table, type Column } from './Table'

interface Row {
  id: number
  name: string
  rating: number
}

const columns: Column<Row>[] = [
  { key: 'name', header: 'Name', render: (r) => r.name, sortable: true, sortValue: (r) => r.name },
  { key: 'rating', header: 'Rating', render: (r) => r.rating, sortable: true, sortValue: (r) => r.rating },
]

const rows: Row[] = [
  { id: 1, name: 'Bob', rating: 1600 },
  { id: 2, name: 'Alice', rating: 1500 },
]

describe('Table', () => {
  it('renders headers', () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Rating')).toBeInTheDocument()
  })

  it('renders rows', () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    expect(screen.getByText('Alice')).toBeInTheDocument()
    expect(screen.getByText('Bob')).toBeInTheDocument()
  })

  it('shows empty message when no rows', () => {
    render(<Table columns={columns} rows={[]} rowKey={(r) => r.id} emptyMessage="No players" />)
    expect(screen.getByText('No players')).toBeInTheDocument()
  })

  it('sorts by column on click', async () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    const nameHeader = screen.getByText('Name')
    await userEvent.click(nameHeader)
    const cells = screen.getAllByRole('cell')
    // After ascending sort by name: Alice first, Bob second
    expect(cells[0].textContent).toBe('Alice')
  })
})
