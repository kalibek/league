import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Badge } from './Badge'

describe('Badge', () => {
  it('renders DRAFT badge', () => {
    render(<Badge variant="DRAFT" />)
    expect(screen.getByText('Draft')).toBeInTheDocument()
  })

  it('renders IN_PROGRESS badge', () => {
    render(<Badge variant="IN_PROGRESS" />)
    expect(screen.getByText('In Progress')).toBeInTheDocument()
  })

  it('renders DONE badge', () => {
    render(<Badge variant="DONE" />)
    expect(screen.getByText('Done')).toBeInTheDocument()
  })

  it('renders DNS badge', () => {
    render(<Badge variant="DNS" />)
    expect(screen.getByText('DNS')).toBeInTheDocument()
  })

  it('renders custom label', () => {
    render(<Badge variant="DRAFT" label="Custom" />)
    expect(screen.getByText('Custom')).toBeInTheDocument()
  })
})
