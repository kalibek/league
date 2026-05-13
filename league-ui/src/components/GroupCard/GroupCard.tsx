import { type ReactNode, useState } from 'react'
import { Link } from 'react-router-dom'
import type { GroupStatus } from '../../types'
import { Badge } from '../Badge/Badge'
import { useTranslation } from 'react-i18next'

interface GroupCardProps {
  division: string
  groupNo: number
  status: GroupStatus
  children: ReactNode
  collapsible?: boolean
  defaultCollapsed?: boolean
  collapseSignal?: number
  groupViewUrl?: string
}

export function GroupCard({
  division,
  groupNo,
  status,
  children,
  collapsible = false,
  defaultCollapsed = false,
  collapseSignal = 0,
  groupViewUrl,
}: GroupCardProps) {
  const { t } = useTranslation()
  const [seenSignal, setSeenSignal] = useState(collapseSignal)
  const [collapsed, setCollapsed] = useState(defaultCollapsed)

  if (collapsible && collapseSignal !== seenSignal) {
    setSeenSignal(collapseSignal)
    if (collapseSignal > 0) setCollapsed(true)
  }

  const title =
    division === 'S' || division === 'Superleague'
      ? 'Superleague'
      : t('groupCard.divisionGroup', { division, no: groupNo })

  return (
    <div
      className="rounded-xl overflow-hidden shadow-sm"
      style={{ border: '1px solid var(--border)' }}
    >
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
          <span
            style={{
              color: 'rgba(255,255,255,0.7)',
              fontSize: 18,
              lineHeight: 1,
              transition: 'transform 0.2s',
              transform: collapsed ? 'rotate(-90deg)' : 'rotate(0deg)',
              display: 'inline-block',
              flexShrink: 0,
            }}
          >
            ▾
          </span>
        )}
        <h3
          style={{ color: '#fff', fontWeight: 600, fontSize: 13, letterSpacing: '-0.1px', flex: 1 }}
        >
          {title}
        </h3>
        {groupViewUrl && (
          <Link
            to={groupViewUrl}
            onClick={(e) => e.stopPropagation()}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 20,
              height: 20,
              flexShrink: 0,
              transition: 'opacity 0.2s',
            }}
            onMouseEnter={(e) => {
              ;(e.currentTarget as HTMLElement).style.opacity = '1'
            }}
            onMouseLeave={(e) => {
              ;(e.currentTarget as HTMLElement).style.opacity = '0.6'
            }}
            title="View group details"
          >
            <svg
              width="12"
              height="12"
              viewBox="0 0 12 12"
              fill="none"
              style={{
                color: 'rgba(255,255,255,0.6)',
              }}
            >
              <path
                d="M2 10L10 2M10 2H5M10 2V7"
                stroke="currentColor"
                strokeWidth="1.2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </Link>
        )}
        <Badge variant={status} />
      </div>
      {!collapsed && <div className="p-4 bg-white">{children}</div>}
    </div>
  )
}
