import { Link } from 'react-router-dom'
import { useLeagues } from '../hooks/useLeagues'
import { useAuth } from '../hooks/useAuth'
import { LeagueCard } from '../components/LeagueCard/LeagueCard'

export function LeaguesPage() {
  const { leagues, loading, error } = useLeagues()
  const { isAdmin } = useAuth()

  return (
    <div className="max-w-4xl mx-auto py-10 px-4">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 style={{ fontSize: 28, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.5px' }}>
            All Leagues
          </h1>
          <p style={{ fontSize: 14, color: '#64748b', marginTop: 4 }}>
            Browse and join table tennis leagues
          </p>
        </div>
        {isAdmin && (
          <Link
            to="/leagues/new"
            style={{
              backgroundColor: 'var(--orange)',
              color: '#fff',
              fontWeight: 700,
              fontSize: 14,
              padding: '10px 20px',
              borderRadius: 8,
              textDecoration: 'none',
            }}
            className="hover:opacity-90 transition-opacity"
          >
            + New League
          </Link>
        )}
      </div>

      {loading && (
        <div style={{ color: '#94a3b8', textAlign: 'center', padding: '48px 0', fontSize: 14 }}>
          Loading…
        </div>
      )}
      {error && (
        <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '12px 16px', fontSize: 14, marginBottom: 16 }}>
          {error}
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {leagues.map((league) => (
          <LeagueCard key={league.leagueId} league={league} />
        ))}
      </div>

      {!loading && leagues.length === 0 && (
        <div style={{ textAlign: 'center', padding: '64px 0', color: '#94a3b8' }}>
          <div style={{ fontSize: 48, marginBottom: 12 }}>🏆</div>
          <p style={{ fontSize: 16, fontWeight: 600, color: '#64748b' }}>No leagues yet.</p>
        </div>
      )}
    </div>
  )
}
