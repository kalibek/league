import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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

  it('renders children by default', () => {
    render(
      <GroupCard division="A" groupNo={1} status="DRAFT">
        <p>child content</p>
      </GroupCard>
    )
    expect(screen.getByText('child content')).toBeInTheDocument()
  })

  it('renders collapsed when defaultCollapsed=true', () => {
    render(
      <GroupCard division="A" groupNo={1} status="DRAFT" collapsible defaultCollapsed>
        <p>hidden content</p>
      </GroupCard>
    )
    expect(screen.queryByText('hidden content')).not.toBeInTheDocument()
  })

  it('expands when header is clicked in collapsible mode', async () => {
    render(
      <GroupCard division="A" groupNo={1} status="DRAFT" collapsible defaultCollapsed>
        <p>toggled content</p>
      </GroupCard>
    )
    expect(screen.queryByText('toggled content')).not.toBeInTheDocument()
    await userEvent.click(screen.getByText(/Division A — Group 1/))
    expect(screen.getByText('toggled content')).toBeInTheDocument()
  })

  it('collapses when collapseSignal increases', () => {
    const { rerender } = render(
      <GroupCard division="A" groupNo={1} status="DRAFT" collapsible collapseSignal={0}>
        <p>visible</p>
      </GroupCard>
    )
    expect(screen.getByText('visible')).toBeInTheDocument()
    rerender(
      <GroupCard division="A" groupNo={1} status="DRAFT" collapsible collapseSignal={1}>
        <p>visible</p>
      </GroupCard>
    )
    expect(screen.queryByText('visible')).not.toBeInTheDocument()
  })
})
