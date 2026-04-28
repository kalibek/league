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

  it('toggles sort direction when clicking the same column twice', async () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    const nameHeader = screen.getByText('Name')
    // First click: sort asc
    await userEvent.click(nameHeader)
    let cells = screen.getAllByRole('cell')
    expect(cells[0].textContent).toBe('Alice')
    // Second click: sort desc — Bob should be first
    await userEvent.click(nameHeader)
    cells = screen.getAllByRole('cell')
    expect(cells[0].textContent).toBe('Bob')
  })

  it('sorts by a different column', async () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    const ratingHeader = screen.getByText('Rating')
    await userEvent.click(ratingHeader)
    const cells = screen.getAllByRole('cell')
    // Asc by rating: Alice (1500) first, Bob (1600) second
    expect(cells[1].textContent).toBe('1500')
  })

  it('shows sort indicator arrow on active column', async () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    await userEvent.click(screen.getByText('Name'))
    expect(screen.getByText('↑')).toBeInTheDocument()
  })

  it('shows desc arrow after second click', async () => {
    render(<Table columns={columns} rows={rows} rowKey={(r) => r.id} />)
    await userEvent.click(screen.getByText('Name'))
    await userEvent.click(screen.getByText('Name'))
    expect(screen.getByText('↓')).toBeInTheDocument()
  })

  it('ignores click on non-sortable column', async () => {
    const nonSortableColumns: Column<Row>[] = [
      { key: 'name', header: 'Name', render: (r) => r.name, sortable: false },
    ]
    render(<Table columns={nonSortableColumns} rows={rows} rowKey={(r) => r.id} />)
    // Should not throw
    await userEvent.click(screen.getByText('Name'))
    expect(screen.getByText('Bob')).toBeInTheDocument()
  })
})
