import { useState } from 'react'
import { Link } from 'react-router-dom'
import { usePlayers } from '../hooks/usePlayers'
import { Table, type Column } from '../components/Table/Table'
import type { User } from '../types'
import { useAuth } from '../hooks/useAuth'

export function PlayersPage() {
  const [sort, setSort] = useState<'rating' | 'name'>('rating')
  const [query, setQuery] = useState('')
  const { players, loading, error } = usePlayers({ q: query, sort, limit: 100, offset: 0 })
  const { isAdmin, roles } = useAuth()
  const canManage = isAdmin || Object.values(roles).some((r) => r.includes('maintainer'))

  const columns: Column<User>[] = [
    {
      key: 'rank',
      header: '#',
      render: (p) => (
        <span style={{ color: '#94a3b8', fontSize: 12, fontWeight: 600 }}>{players.indexOf(p) + 1}</span>
      ),
    },
    {
      key: 'name',
      header: 'Player',
      render: (p) => (
        <Link
          to={`/players/${p.userId}`}
          style={{ fontWeight: 600, color: 'var(--navy)', textDecoration: 'none' }}
          className="hover:text-[#FF7A00] transition-colors"
        >
          {p.firstName} {p.lastName}
        </Link>
      ),
      sortable: true,
      sortValue: (p) => `${p.lastName} ${p.firstName}`,
    },
    {
      key: 'rating',
      header: 'Rating',
      render: (p) => (
        <span style={{ fontWeight: 700, fontSize: 15, color: 'var(--navy)' }}>{Math.round(p.currentRating)}</span>
      ),
      sortable: true,
      sortValue: (p) => p.currentRating,
    },
    {
      key: 'deviation',
      header: 'RD',
      render: (p) => (
        <span style={{ fontSize: 12, color: '#94a3b8' }}>±{Math.round(p.deviation)}</span>
      ),
    },
  ]

  const sortBtnStyle = (active: boolean): React.CSSProperties => ({
    fontSize: 13,
    fontWeight: active ? 700 : 500,
    color: active ? 'var(--orange)' : '#64748b',
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    padding: '4px 8px',
    borderRadius: 6,
    backgroundColor: active ? 'rgba(255,122,0,0.08)' : 'transparent',
  })

  return (
    <div className="max-w-5xl mx-auto py-10 px-4">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 style={{ fontSize: 28, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.5px' }}>
            Players
          </h1>
          <p style={{ fontSize: 14, color: '#64748b', marginTop: 4 }}>
            {players.length > 0 ? `${players.length} registered players` : 'Registered player rankings'}
          </p>
        </div>
        {canManage && (
          <div style={{ display: 'flex', gap: 8 }}>
            <Link
              to="/players/import"
              style={{
                fontSize: 13,
                fontWeight: 600,
                color: 'var(--navy)',
                border: '1.5px solid var(--navy)',
                borderRadius: 8,
                padding: '8px 16px',
                textDecoration: 'none',
                backgroundColor: 'transparent',
              }}
              className="hover:bg-[#0B3C5D] hover:text-white transition-colors"
            >
              Import CSV
            </Link>
            <Link
              to="/players/new"
              style={{
                fontSize: 13,
                fontWeight: 700,
                color: '#fff',
                backgroundColor: 'var(--orange)',
                borderRadius: 8,
                padding: '8px 16px',
                textDecoration: 'none',
              }}
              className="hover:opacity-90 transition-opacity"
            >
              + Add Player
            </Link>
          </div>
        )}
      </div>

      {/* Search + sort */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'row',
          gap: 12,
          marginBottom: 20,
          alignItems: 'center',
          flexWrap: 'wrap',
        }}
      >
        <div style={{ position: 'relative', flex: 1, minWidth: 220 }}>
          <svg
            style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', width: 16, height: 16, color: '#94a3b8', pointerEvents: 'none' }}
            fill="none" stroke="currentColor" viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search by name, city, blade…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            style={{
              width: '100%',
              paddingLeft: 38,
              paddingRight: query ? 36 : 14,
              paddingTop: 10,
              paddingBottom: 10,
              border: '1.5px solid var(--border)',
              borderRadius: 8,
              fontSize: 14,
              color: 'var(--dark)',
              backgroundColor: '#fff',
              outline: 'none',
            }}
            className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
          />
          {query && (
            <button
              style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', background: 'none', border: 'none', cursor: 'pointer', fontSize: 13 }}
              onClick={() => setQuery('')}
              aria-label="Clear search"
            >
              ✕
            </button>
          )}
        </div>

        {!query && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexShrink: 0 }}>
            <span style={{ fontSize: 12, color: '#94a3b8', fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase' }}>Sort:</span>
            <button style={sortBtnStyle(sort === 'rating')} onClick={() => setSort('rating')}>Rating</button>
            <button style={sortBtnStyle(sort === 'name')} onClick={() => setSort('name')}>Name</button>
          </div>
        )}
      </div>

      {loading && <p style={{ color: '#94a3b8', fontSize: 13, marginBottom: 8 }}>Searching…</p>}
      {error && (
        <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '10px 14px', fontSize: 13, marginBottom: 16 }}>
          {error}
        </div>
      )}

      <Table
        columns={columns}
        rows={players}
        rowKey={(p) => p.userId}
        emptyMessage={query ? `No players match "${query}"` : 'No players found'}
      />
    </div>
  )
}
