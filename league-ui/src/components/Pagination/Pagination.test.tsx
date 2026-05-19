import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Pagination } from './Pagination'

describe('Pagination', () => {
  it('renders nothing when total fits on one page', () => {
    const { container } = render(
      <Pagination page={1} pageSize={10} total={5} onPageChange={() => {}} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders correct number of pages', () => {
    const onPageChange = vi.fn()
    render(
      <Pagination page={1} pageSize={10} total={100} onPageChange={onPageChange} />
    )
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('highlights current page', () => {
    const { rerender } = render(
      <Pagination page={1} pageSize={10} total={100} onPageChange={() => {}} />
    )
    let currentBtn = screen.getByRole('button', { name: '1' })
    expect(currentBtn).toHaveStyle({ backgroundColor: 'var(--orange)' })

    rerender(
      <Pagination page={3} pageSize={10} total={100} onPageChange={() => {}} />
    )
    currentBtn = screen.getByRole('button', { name: '3' })
    expect(currentBtn).toHaveStyle({ backgroundColor: 'var(--orange)' })
  })

  it('calls onPageChange with correct page on button click', async () => {
    const user = userEvent.setup()
    const onPageChange = vi.fn()
    render(
      <Pagination page={1} pageSize={10} total={100} onPageChange={onPageChange} />
    )
    const page2Btn = screen.getByRole('button', { name: '2' })
    await user.click(page2Btn)
    expect(onPageChange).toHaveBeenCalledWith(2)
  })

  it('disables prev button on page 1', () => {
    render(
      <Pagination page={1} pageSize={10} total={100} onPageChange={() => {}} />
    )
    const prevBtn = screen.getByLabelText('Previous page')
    expect(prevBtn).toBeDisabled()
  })

  it('disables next button on last page', () => {
    const onPageChange = vi.fn()
    render(
      <Pagination page={10} pageSize={10} total={100} onPageChange={onPageChange} />
    )
    const nextBtn = screen.getByLabelText('Next page')
    expect(nextBtn).toBeDisabled()
  })

  it('calls onPageChange with next page on next button click', async () => {
    const user = userEvent.setup()
    const onPageChange = vi.fn()
    render(
      <Pagination page={2} pageSize={10} total={100} onPageChange={onPageChange} />
    )
    const nextBtn = screen.getByLabelText('Next page')
    await user.click(nextBtn)
    expect(onPageChange).toHaveBeenCalledWith(3)
  })

  it('calls onPageChange with prev page on prev button click', async () => {
    const user = userEvent.setup()
    const onPageChange = vi.fn()
    render(
      <Pagination page={3} pageSize={10} total={100} onPageChange={onPageChange} />
    )
    const prevBtn = screen.getByLabelText('Previous page')
    await user.click(prevBtn)
    expect(onPageChange).toHaveBeenCalledWith(2)
  })

  it('shows ellipsis when there are many pages', () => {
    render(
      <Pagination page={5} pageSize={10} total={500} onPageChange={() => {}} />
    )
    const ellipsis = screen.getAllByText('…')
    expect(ellipsis.length).toBeGreaterThan(0)
  })

  it('always shows first and last page buttons', () => {
    render(
      <Pagination page={50} pageSize={10} total={1000} onPageChange={() => {}} />
    )
    const firstBtn = screen.getByRole('button', { name: '1' })
    const lastBtn = screen.getByRole('button', { name: '100' })
    expect(firstBtn).toBeInTheDocument()
    expect(lastBtn).toBeInTheDocument()
  })
})
