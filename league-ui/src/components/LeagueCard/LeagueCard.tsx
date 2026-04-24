import { Link } from 'react-router-dom'
import type { League } from '../../types'

export function LeagueCard({ league }: { league: League }) {
  const maintainers = league.maintainers ?? []
  const eventCount = league.eventCount ?? 0

  return (
    <Link
      to={`/leagues/${league.leagueId}`}
      style={{
        display: 'block',
        backgroundColor: '#fff',
        borderRadius: 12,
        border: '1px solid var(--border)',
        padding: '18px 20px',
        textDecoration: 'none',
        boxShadow: '0 1px 4px rgba(11,60,93,0.06)',
      }}
      className="hover:border-[#FF7A00] hover:shadow-md transition-all"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h2 style={{ fontWeight: 700, fontSize: 16, color: 'var(--navy)', marginBottom: 4, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {league.title}
          </h2>
          {league.description && (
            <p style={{ fontSize: 13, color: '#64748b', overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 1, WebkitBoxOrient: 'vertical' }}>
              {league.description}
            </p>
          )}
        </div>
        <span
          style={{
            flexShrink: 0,
            fontSize: 11,
            fontWeight: 700,
            color: 'var(--navy)',
            backgroundColor: 'rgba(11,60,93,0.08)',
            borderRadius: 9999,
            padding: '4px 12px',
            whiteSpace: 'nowrap',
          }}
        >
          {eventCount} {eventCount === 1 ? 'event' : 'events'}
        </span>
      </div>

      {maintainers.length > 0 && (
        <div style={{ marginTop: 12, display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
          <span style={{ fontSize: 11, color: '#94a3b8' }}>Maintained by</span>
          {maintainers.map((m) => (
            <span
              key={m.userId}
              style={{
                fontSize: 11,
                fontWeight: 600,
                color: '#475569',
                backgroundColor: '#f1f5f9',
                borderRadius: 4,
                padding: '2px 8px',
              }}
            >
              {m.firstName} {m.lastName}
            </span>
          ))}
        </div>
      )}
    </Link>
  )
}
