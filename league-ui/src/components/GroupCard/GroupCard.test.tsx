import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { GroupCard } from './GroupCard'

describe('GroupCard', () => {
  it('renders division and group number', () => {
    render(
      <GroupCard division="A" groupNo={1} status="IN_PROGRESS">
        <p>content</p>
      </GroupCard>
    )
    expect(screen.getByText(/Division A — Group 1/)).toBeInTheDocument()
  })

  it('renders Superleague title', () => {
    render(
      <GroupCard division="Superleague" groupNo={0} status="DRAFT">
        <p>content</p>
      </GroupCard>
    )
    expect(screen.getByText('Superleague')).toBeInTheDocument()
  })

  it('shows status badge', () => {
    render(
      <GroupCard division="B" groupNo={2} status="DONE">
        <p>x</p>
      </GroupCard>
    )
    expect(screen.getByText('Done')).toBeInTheDocument()
  })
})
