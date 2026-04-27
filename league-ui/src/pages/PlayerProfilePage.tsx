import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { usePlayer } from '../hooks/usePlayers'
import { usePlayerEvents } from '../hooks/usePlayers'
import { RatingDelta } from '../components/RatingDelta/RatingDelta'
import { Badge } from '../components/Badge/Badge'
import { Button } from '../components/Button/Button'
import { formatDate } from '../hooks/utils'
import type { PlayerEventSummary, PlayerGroupSummary, PlayerMatchSummary } from '../types'

function MatchRow({ m }: { m: PlayerMatchSummary }) {
  const { t } = useTranslation()

  const scoreDisplay =
    m.status !== 'DONE'
      ? '— : —'
      : m.withdraw
      ? t('playerProfile.woDns')
      : m.oppWithdraw
      ? t('playerProfile.woDnsOpp')
      : `${m.myScore ?? '?'} : ${m.oppScore ?? '?'}`

  const resultBadge =
    m.status !== 'DONE' ? null : m.withdraw ? (
      <span style={{ fontSize: 11, padding: '2px 7px', borderRadius: 4, backgroundColor: '#f1f5f9', color: '#64748b', fontWeight: 600 }}>{t('playerProfile.woDns')}</span>
    ) : m.won === true ? (
      <span style={{ fontSize: 11, padding: '2px 7px', borderRadius: 4, backgroundColor: '#dcfce7', color: '#16a34a', fontWeight: 700 }}>W</span>
    ) : m.won === false ? (
      <span style={{ fontSize: 11, padding: '2px 7px', borderRadius: 4, backgroundColor: '#fee2e2', color: '#dc2626', fontWeight: 700 }}>L</span>
    ) : null

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid #f1f5f9', fontSize: 13 }}>
      <Link to={`/players/${m.opponentId}`}>
        <span style={{ color: 'var(--dark)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: 180 }}>
          {m.opponentName || '—'}
        </span>
      </Link>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
        <span style={{ color: '#64748b', fontVariantNumeric: 'tabular-nums' }}>{scoreDisplay}</span>
        {resultBadge}
        {m.ratingDelta != null && (
          <span style={{
            fontSize: 11,
            fontWeight: 700,
            color: m.ratingDelta > 0 ? '#16a34a' : m.ratingDelta < 0 ? '#dc2626' : '#94a3b8',
            minWidth: 36,
            textAlign: 'right',
          }}>
            {m.ratingDelta > 0 ? '+' : ''}{Math.round(m.ratingDelta)}
          </span>
        )}
      </div>
    </div>
  )
}

function GroupBlock({ g }: { g: PlayerGroupSummary }) {
  const { t } = useTranslation()
  return (
    <div style={{ borderRadius: 8, border: '1px solid var(--border)', backgroundColor: '#f8fafc', padding: 12 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
        <span style={{ fontSize: 12, fontWeight: 600, color: '#64748b' }}>
          {g.division} · {t('playerProfile.group', { no: g.groupNo })}
        </span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {g.place > 0 && (
            <span style={{ fontSize: 11, color: '#94a3b8' }}>
              {t('playerProfile.place', { place: g.place })} · {t('playerProfile.points', { points: g.points })}
            </span>
          )}
          {g.advances && (
            <span style={{ fontSize: 11, padding: '2px 8px', borderRadius: 4, backgroundColor: '#dbeafe', color: '#1e40af', fontWeight: 600 }}>
              {t('playerProfile.advances')}
            </span>
          )}
          {g.recedes && (
            <span style={{ fontSize: 11, padding: '2px 8px', borderRadius: 4, backgroundColor: '#fff7ed', color: '#c2410c', fontWeight: 600 }}>
              {t('playerProfile.recedes')}
            </span>
          )}
        </div>
      </div>
      <div>
        {(g.matches ?? []).length === 0 ? (
          <p style={{ fontSize: 12, color: '#94a3b8' }}>{t('playerProfile.noMatches')}</p>
        ) : (
          (g.matches ?? []).map((m) => <MatchRow key={m.matchId} m={m} />)
        )}
      </div>
    </div>
  )
}

function EventCard({ ev }: { ev: PlayerEventSummary }) {
  const { t } = useTranslation()
  return (
    <div style={{ borderRadius: 12, border: '1px solid var(--border)', backgroundColor: '#fff', padding: '18px 20px', boxShadow: '0 1px 4px rgba(11,60,93,0.04)' }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 10 }}>
        <div>
          <Link
              to={`/leagues/${ev.leagueId}/events/${ev.eventId}`}
              style={{ fontSize: 13, color: 'var(--orange)', fontWeight: 600, textDecoration: 'none' }}
              className="hover:opacity-80"
          >
            <p style={{ fontWeight: 700, color: 'var(--navy)', fontSize: 15, marginBottom: 3 }}>{ev.title}</p>
          </Link>
          <p style={{ fontSize: 12, color: '#94a3b8' }}>
            {formatDate(ev.startDate)} — {formatDate(ev.endDate)}
          </p>
        </div>
        <Badge variant={ev.status} />
      </div>

      {ev.ratingBefore != null && ev.ratingAfter != null && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          backgroundColor: '#f8fafc',
          borderRadius: 8,
          padding: '8px 12px',
          marginBottom: 12,
          fontSize: 13,
        }}>
          <span style={{ color: '#64748b' }}>{t('playerProfile.ratingLabel')}</span>
          <span style={{ fontWeight: 700, color: 'var(--navy)' }}>{Math.round(ev.ratingBefore)}</span>
          <span style={{ color: '#cbd5e1' }}>→</span>
          <span style={{ fontWeight: 700, color: 'var(--navy)' }}>{Math.round(ev.ratingAfter)}</span>
          {ev.ratingDelta !== 0 && (
            <RatingDelta delta={Math.round(ev.ratingDelta)} />
          )}
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {ev.groups.map((g) => (
          <GroupBlock key={g.groupId} g={g} />
        ))}
      </div>
    </div>
  )
}

export function PlayerProfilePage() {
  const { t } = useTranslation()
  const { id } = useParams<{ id: string }>()
  const userId = Number(id)
  const { player, loading: playerLoading, error } = usePlayer(userId)
  const { events, total, loadMore, loading: eventsLoading, hasMore } = usePlayerEvents(userId)

  if (playerLoading) return <div style={{ padding: '48px 16px', textAlign: 'center', color: '#94a3b8', fontSize: 14 }}>{t('playerProfile.loading')}</div>
  if (error) return <div style={{ padding: '48px 16px', textAlign: 'center', color: '#dc2626', fontSize: 14 }}>{error}</div>
  if (!player) return null

  const lastDelta = player.ratingHistory.length > 0 ? player.ratingHistory[0].delta : 0
  const prof = player.profile

  const profileFields: { label: string; value: string | null | undefined }[] = prof ? [
    { label: t('playerProfile.country'), value: prof.country?.name },
    { label: t('playerProfile.city'), value: prof.city?.name },
    { label: t('playerProfile.grip'), value: prof.grip ? prof.grip.charAt(0).toUpperCase() + prof.grip.slice(1) : null },
    { label: t('playerProfile.gender'), value: prof.gender ? prof.gender.charAt(0).toUpperCase() + prof.gender.slice(1) : null },
    { label: t('playerProfile.blade'), value: prof.blade?.name },
    { label: t('playerProfile.fhRubber'), value: prof.fhRubber?.name },
    { label: t('playerProfile.bhRubber'), value: prof.bhRubber?.name },
  ].filter(f => f.value) : []

  return (
    <div className="max-w-3xl mx-auto py-10 px-4">
      <Link
        to="/players"
        style={{ fontSize: 13, color: '#64748b', textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: 4, marginBottom: 24 }}
        className="hover:text-[#0B3C5D] transition-colors"
      >
        {t('playerProfile.backToPlayers')}
      </Link>

      {/* Player header card */}
      <div
        style={{
          borderRadius: 16,
          border: '1px solid var(--border)',
          backgroundColor: '#fff',
          padding: '24px 28px',
          marginBottom: 16,
          boxShadow: '0 2px 8px rgba(11,60,93,0.07)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <div>
            <h1 style={{ fontSize: 26, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.5px', marginBottom: 4 }}>
              {player.firstName} {player.lastName}
            </h1>
          </div>
          <div style={{ textAlign: 'right' }}>
            <p style={{ fontSize: 36, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-1px', lineHeight: 1 }}>
              {Math.round(player.currentRating)}
            </p>
            <p style={{ fontSize: 11, color: '#94a3b8', marginTop: 3 }}>±{Math.round(player.deviation)} RD</p>
            {lastDelta !== 0 && (
              <div style={{ marginTop: 6 }}>
                <RatingDelta delta={Math.round(lastDelta)} />
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Profile card */}
      <div
        style={{
          borderRadius: 12,
          border: '1px solid var(--border)',
          backgroundColor: '#fff',
          padding: '18px 24px',
          marginBottom: 28,
          boxShadow: '0 1px 4px rgba(11,60,93,0.04)',
        }}
      >
        {!prof || !prof.isComplete ? (
          <p style={{ fontSize: 13, color: '#94a3b8', fontStyle: 'italic' }}>{t('playerProfile.profileNotFilled')}</p>
        ) : (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '10px 32px' }}>
            {profileFields.map(({ label, value }) => (
              <div key={label}>
                <div style={{ fontSize: 10, fontWeight: 700, color: '#94a3b8', letterSpacing: '0.06em', textTransform: 'uppercase', marginBottom: 2 }}>
                  {label}
                </div>
                <div style={{ fontSize: 14, fontWeight: 500, color: 'var(--dark)' }}>{value}</div>
              </div>
            ))}
          </div>
        )}
      </div>

      <h2 style={{ fontSize: 11, fontWeight: 700, color: '#94a3b8', letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 14 }}>
        {t('playerProfile.eventHistory')} {total > 0 && <span style={{ fontWeight: 400 }}>({total})</span>}
      </h2>

      {events.length === 0 && !eventsLoading && (
        <p style={{ fontSize: 14, color: '#94a3b8', textAlign: 'center', padding: '32px 0' }}>{t('playerProfile.noEventsYet')}</p>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {events.map((ev) => (
          <EventCard key={ev.eventId} ev={ev} />
        ))}
      </div>

      {hasMore && (
        <div style={{ marginTop: 20, display: 'flex', justifyContent: 'center' }}>
          <Button variant="secondary" onClick={loadMore} loading={eventsLoading}>
            {t('playerProfile.loadOlderEvents')}
          </Button>
        </div>
      )}

      {eventsLoading && events.length === 0 && (
        <p style={{ fontSize: 13, color: '#94a3b8', textAlign: 'center', padding: '32px 0' }}>{t('playerProfile.loadingEvents')}</p>
      )}
    </div>
  )
}
