import { Link } from 'react-router-dom'
import { useLeagues } from '../hooks/useLeagues'
import { usePlayers } from '../hooks/usePlayers'
import { LeagueCard } from '../components/LeagueCard/LeagueCard'
import { Table, type Column } from '../components/Table/Table'
import type { User } from '../types'

function SectionHeader({ title, to, label }: { title: string; to: string; label: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
      <h2 style={{ fontSize: 20, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.3px' }}>
        {title}
      </h2>
      <Link
        to={to}
        style={{ fontSize: 13, color: 'var(--orange)', fontWeight: 600, textDecoration: 'none' }}
        className="hover:opacity-80 transition-opacity"
      >
        {label} →
      </Link>
    </div>
  )
}

export function HomePage() {
  const { leagues, loading: leaguesLoading, error: leaguesError } = useLeagues()
  const { players, loading: playersLoading, error: playersError } = usePlayers({ sort: 'rating', limit: 10, offset: 0 })

  const topLeagues = leagues.slice(0, 3)

  const playerColumns: Column<User>[] = [
    {
      key: 'rank',
      header: '#',
      render: (p) => (
        <span style={{ color: '#94a3b8', fontSize: 12, fontWeight: 600 }}>
          {players.indexOf(p) + 1}
        </span>
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
    },
    {
      key: 'rating',
      header: 'Rating',
      render: (p) => (
        <span style={{ fontWeight: 700, fontSize: 15, color: 'var(--navy)' }}>
          {Math.round(p.currentRating)}
        </span>
      ),
    },
    {
      key: 'deviation',
      header: 'RD',
      render: (p) => (
        <span style={{ fontSize: 12, color: '#94a3b8' }}>±{Math.round(p.deviation)}</span>
      ),
    },
  ]

  return (
    <div className="max-w-4xl mx-auto py-10 px-4" style={{ display: 'flex', flexDirection: 'column', gap: 48 }}>

      {/* Leagues section */}
      <section>
        <SectionHeader title="Leagues" to="/leagues" label="View all" />

        {leaguesLoading && <p style={{ color: '#94a3b8', fontSize: 14 }}>Loading…</p>}
        {leaguesError && (
          <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '10px 14px', fontSize: 13 }}>
            {leaguesError}
          </div>
        )}
        {!leaguesLoading && leagues.length === 0 && (
          <div style={{ textAlign: 'center', padding: '40px 0', color: '#94a3b8' }}>
            <div style={{ fontSize: 40, marginBottom: 10 }}>🏓</div>
            <p style={{ fontSize: 15, fontWeight: 600, color: '#64748b' }}>No leagues yet.</p>
          </div>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {topLeagues.map((league) => (
            <LeagueCard key={league.leagueId} league={league} />
          ))}
        </div>

        {leagues.length > 3 && (
          <div style={{ marginTop: 12 }}>
            <Link
              to="/leagues"
              style={{
                display: 'block',
                textAlign: 'center',
                padding: '11px',
                borderRadius: 10,
                border: '1.5px dashed var(--border)',
                fontSize: 13,
                fontWeight: 600,
                color: '#64748b',
                textDecoration: 'none',
                backgroundColor: '#fff',
              }}
              className="hover:border-[#FF7A00] hover:text-[#FF7A00] transition-colors"
            >
              Show {leagues.length - 3} more league{leagues.length - 3 !== 1 ? 's' : ''}
            </Link>
          </div>
        )}
      </section>

      {/* Players section */}
      <section>
        <SectionHeader title="Top Players" to="/players" label="View all" />

        {playersLoading && <p style={{ color: '#94a3b8', fontSize: 14 }}>Loading…</p>}
        {playersError && (
          <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '10px 14px', fontSize: 13 }}>
            {playersError}
          </div>
        )}

        <Table
          columns={playerColumns}
          rows={players}
          rowKey={(p) => p.userId}
          emptyMessage="No players found"
        />

        {!playersLoading && players.length > 0 && (
          <div style={{ marginTop: 12 }}>
            <Link
              to="/players"
              style={{
                display: 'block',
                textAlign: 'center',
                padding: '11px',
                borderRadius: 10,
                border: '1.5px dashed var(--border)',
                fontSize: 13,
                fontWeight: 600,
                color: '#64748b',
                textDecoration: 'none',
                backgroundColor: '#fff',
              }}
              className="hover:border-[#FF7A00] hover:text-[#FF7A00] transition-colors"
            >
              Show more players
            </Link>
          </div>
        )}
      </section>

    </div>
  )
}
