import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RatingDelta } from './RatingDelta'

describe('RatingDelta', () => {
  it('shows positive delta in green with + sign', () => {
    render(<RatingDelta delta={12} />)
    const el = screen.getByText('+12')
    expect(el.className).toContain('text-green-600')
  })

  it('shows negative delta in red', () => {
    render(<RatingDelta delta={-8} />)
    const el = screen.getByText('-8')
    expect(el.className).toContain('text-red-600')
  })

  it('shows zero delta as neutral', () => {
    render(<RatingDelta delta={0} />)
    const el = screen.getByText('±0')
    expect(el.className).toContain('text-gray-400')
  })
})
