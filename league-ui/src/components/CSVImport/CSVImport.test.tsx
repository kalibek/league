import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CSVImport } from './CSVImport'

describe('CSVImport', () => {
  it('renders drop zone with instructions', () => {
    render(<CSVImport onImport={vi.fn()} />)
    expect(screen.getByText(/drag and drop a csv/i)).toBeInTheDocument()
  })

  it('shows file name after selection', async () => {
    render(<CSVImport onImport={vi.fn()} />)
    const file = new File(['first_name,last_name,email\nAlice,Smith,a@b.com'], 'players.csv', { type: 'text/csv' })
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await userEvent.upload(input, file)
    expect(screen.getByText('players.csv')).toBeInTheDocument()
  })

  it('shows import button after file selection', async () => {
    render(<CSVImport onImport={vi.fn()} />)
    const file = new File(['a,b,c'], 'test.csv', { type: 'text/csv' })
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await userEvent.upload(input, file)
    expect(screen.getByRole('button', { name: /import players/i })).toBeInTheDocument()
  })

  it('calls onImport when button clicked', async () => {
    const onImport = vi.fn().mockResolvedValue({ imported: 3, skipped: 0, errors: [] })
    render(<CSVImport onImport={onImport} />)
    const file = new File(['a,b,c'], 'test.csv', { type: 'text/csv' })
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await userEvent.upload(input, file)
    await userEvent.click(screen.getByRole('button', { name: /import players/i }))
    expect(onImport).toHaveBeenCalledWith(file)
  })
})
