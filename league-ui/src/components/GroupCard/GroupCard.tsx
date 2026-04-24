import { useState, type ReactNode } from 'react'
import type { GroupStatus } from '../../types'
import { Badge } from '../Badge/Badge'

interface GroupCardProps {
  division: string
  groupNo: number
  status: GroupStatus
  children: ReactNode
  collapsible?: boolean
  defaultCollapsed?: boolean
  collapseSignal?: number
}

export function GroupCard({
  division,
  groupNo,
  status,
  children,
  collapsible = false,
  defaultCollapsed = false,
  collapseSignal = 0,
}: GroupCardProps) {
  const [seenSignal, setSeenSignal] = useState(collapseSignal)
  const [collapsed, setCollapsed] = useState(defaultCollapsed)

  if (collapsible && collapseSignal !== seenSignal) {
    setSeenSignal(collapseSignal)
    if (collapseSignal > 0) setCollapsed(true)
  }

  const title = division === 'Superleague' ? 'Superleague' : `Division ${division} — Group ${groupNo}`

  return (
    <div className="rounded-xl overflow-hidden shadow-sm" style={{ border: '1px solid var(--border)' }}>
      <div
        className="flex items-center gap-3 px-4 py-3"
        style={{
          backgroundColor: 'var(--navy)',
          borderBottom: collapsed ? 'none' : '1px solid rgba(255,255,255,0.1)',
          cursor: collapsible ? 'pointer' : 'default',
          userSelect: 'none',
        }}
        onClick={collapsible ? () => setCollapsed((c) => !c) : undefined}
      >
        {collapsible && (
          <span style={{
            color: 'rgba(255,255,255,0.7)',
            fontSize: 18,
            lineHeight: 1,
            transition: 'transform 0.2s',
            transform: collapsed ? 'rotate(-90deg)' : 'rotate(0deg)',
            display: 'inline-block',
            flexShrink: 0,
          }}>
            ▾
          </span>
        )}
        <h3 style={{ color: '#fff', fontWeight: 600, fontSize: 13, letterSpacing: '-0.1px', flex: 1 }}>{title}</h3>
        <Badge variant={status} />
      </div>
      {!collapsed && <div className="p-4 bg-white">{children}</div>}
    </div>
  )
}
