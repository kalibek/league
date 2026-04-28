import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RatingDelta } from './RatingDelta'

describe('RatingDelta', () => {
  it('shows positive delta with + sign', () => {
    render(<RatingDelta delta={12} />)
    expect(screen.getByText('+12')).toBeInTheDocument()
  })

  it('shows positive delta in green color', () => {
    render(<RatingDelta delta={12} />)
    const el = screen.getByText('+12')
    expect(el).toHaveStyle({ color: '#16a34a' })
  })

  it('shows negative delta without + sign', () => {
    render(<RatingDelta delta={-8} />)
    expect(screen.getByText('-8')).toBeInTheDocument()
  })

  it('shows negative delta in red color', () => {
    render(<RatingDelta delta={-8} />)
    const el = screen.getByText('-8')
    expect(el).toHaveStyle({ color: '#dc2626' })
  })

  it('shows zero delta as ±0', () => {
    render(<RatingDelta delta={0} />)
    expect(screen.getByText('±0')).toBeInTheDocument()
  })

  it('shows zero delta in neutral gray color', () => {
    render(<RatingDelta delta={0} />)
    const el = screen.getByText('±0')
    expect(el).toHaveStyle({ color: '#94a3b8' })
  })
})
