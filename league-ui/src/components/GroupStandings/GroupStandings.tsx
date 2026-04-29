import type { GroupPlayer, Match } from '../../types'
import { isDns } from '../../types'
import { Badge } from '../Badge/Badge'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

interface GroupStandingsProps {
  players: GroupPlayer[]
  matches: Match[]
  onNoShow?: (groupPlayerId: number) => void
  onScoreClick?: (match: Match) => void
}

// Returns null for players with a unique points total (no tie → show '—').
// Returns the backend tiebreakPoints for players in a tied group.
function buildTiebreakDisplayMap(players: GroupPlayer[]): Map<number, number | null> {
  const calcPlayers = players.filter((p) => !p.isNonCalculated)
  const byPoints = new Map<number, GroupPlayer[]>()
  for (const p of calcPlayers) {
    const group = byPoints.get(p.points) ?? []
    group.push(p)
    byPoints.set(p.points, group)
  }
  const result = new Map<number, number | null>()
  for (const group of byPoints.values()) {
    for (const p of group) {
      result.set(p.groupPlayerId, group.length >= 2 ? p.tiebreakPoints : null)
    }
  }
  return result
}

export function GroupStandings({ players, matches, onNoShow, onScoreClick }: GroupStandingsProps) {
  const { t } = useTranslation()
  const sorted = [...players].sort((a, b) => a.seed - b.seed)
  const tiebreakMap = buildTiebreakDisplayMap(players)

  const playerName = (p: GroupPlayer) =>
      p.user ? `${p.user.firstName} ${p.user.lastName} (${Math.round(p.user.currentRating)})` : `#${p.userId}`

  // Build a lookup: `${p1Id}-${p2Id}` (canonical: smaller id first) → Match
  const matchLookup = new Map<string, Match>()
  for (const m of matches) {
    if (m.groupPlayer1Id !== null && m.groupPlayer2Id !== null) {
      const key = `${m.groupPlayer1Id}-${m.groupPlayer2Id}`
      matchLookup.set(key, m)
    }
  }

  const getMatch = (rowPlayer: GroupPlayer, colPlayer: GroupPlayer): Match | undefined => {
    const key1 = `${rowPlayer.groupPlayerId}-${colPlayer.groupPlayerId}`
    const key2 = `${colPlayer.groupPlayerId}-${rowPlayer.groupPlayerId}`
    return matchLookup.get(key1) ?? matchLookup.get(key2)
  }

  const cellContent = (rowPlayer: GroupPlayer, colPlayer: GroupPlayer) => {
    const m = getMatch(rowPlayer, colPlayer)
    if (!m) return '—'
    if (m.score1 === null || m.score2 === null) return '—'

    // Show score from row player's perspective.
    const isP1 = m.groupPlayer1Id === rowPlayer.groupPlayerId
    const s1 = isP1 ? m.score1 : m.score2
    const s2 = isP1 ? m.score2 : m.score1
    // Walkover: W-L / L-W only when a player withdrew or DNS.
    const rowWithdrew = isP1 ? m.withdraw1 : m.withdraw2
    if (m.withdraw1 || m.withdraw2) return rowWithdrew ? 'L-W' : 'W-L'
    return `${s1}:${s2}`
  }

  const p1wins = (rowPlayer: GroupPlayer, colPlayer: GroupPlayer) => {
    const m = getMatch(rowPlayer, colPlayer)
    if (!m) return false
    if (m.score1 === null || m.score2 === null) return false
    return (m.groupPlayer1Id === rowPlayer.groupPlayerId &&
    m.score1 > m.score2) ||
        (m.groupPlayer2Id === rowPlayer.groupPlayerId &&
    m.score1 < m.score2)
  }

  const wins = (p: GroupPlayer) =>
    matches.filter(
      (m) =>
        m.status === 'DONE' &&
        !m.withdraw1 &&
        !m.withdraw2 &&
        ((m.groupPlayer1Id === p.groupPlayerId && m.score1 !== null && m.score2 !== null && m.score1 > m.score2) ||
          (m.groupPlayer2Id === p.groupPlayerId && m.score1 !== null && m.score2 !== null && m.score2 > m.score1))
    ).length

  const losses = (p: GroupPlayer) =>
    matches.filter(
      (m) =>
        m.status === 'DONE' &&
        !m.withdraw1 &&
        !m.withdraw2 &&
        ((m.groupPlayer1Id === p.groupPlayerId && m.score1 !== null && m.score2 !== null && m.score1 < m.score2) ||
          (m.groupPlayer2Id === p.groupPlayerId && m.score1 !== null && m.score2 !== null && m.score2 < m.score1))
    ).length
  const tdStyle: React.CSSProperties = {
    textAlign: 'center',
    borderLeft: '1px solid var(--border)',
  }
  const thStyle: React.CSSProperties = {
    padding: '10px 12px',
    fontSize: 10,
    fontWeight: 700,
    letterSpacing: '0.06em',
    textTransform: 'uppercase',
    color: '#64748b',
    backgroundColor: '#f8fafc',
    borderBottom: '1px solid var(--border)',
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm text-left">
        <thead>
          <tr>
            <th style={{ ...thStyle, width: 40 }}>#</th>
            <th style={thStyle}>{t('groupStandings.player')}</th>

            <th style={{ ...thStyle, textAlign: 'center' }}>{t('groupStandings.place')}</th>
            <th style={{ ...thStyle, textAlign: 'center' }}>{t('groupStandings.ptsWL')}</th>
            <th style={{ ...thStyle, textAlign: 'center' }}>{t('groupStandings.tb')}</th>
            <th style={{ ...thStyle, textAlign: 'center' }}>{t('groupStandings.move')}</th>
            {sorted.map((p, i) => (
                <th
                    key={p.groupPlayerId}
                    style={ {...thStyle, borderLeft: '1px solid var(--border)',textAlign: 'center' }}
                    title={playerName(p)}
                >
                  #{i+1}
                </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sorted.map((p, i) => {
            const dns = isDns(p.groupPlayerId, matches)
            const name = playerName(p)

            return (
              <tr
                key={p.groupPlayerId}
                style={{
                  backgroundColor: i % 2 === 0 ? '#fff' : '#fafbfd',
                  borderBottom: '1px solid var(--border)',
                  opacity: dns ? 0.6 : 1,
                }}
              >
                <td style={{ padding: '10px 12px', color: '#94a3b8', fontSize: 12 }}>
                  {p.seed}
                </td>

                <td style={{ padding: '10px 12px' }}>
                  <div className="flex items-center gap-1.5">
                    <span
                      style={{
                        fontWeight: 500,
                        color: p.isNonCalculated ? '#94a3b8' : 'var(--dark)',
                        fontStyle: p.isNonCalculated ? 'italic' : 'normal',
                        textDecoration: dns ? 'line-through' : 'none',
                      }}
                    >
                      <Link to={`/players/${p.userId}`}>{name}</Link>
                    </span>
                    {p.isNonCalculated && (
                      <span style={{ fontSize: 11, color: '#94a3b8' }}>{t('groupStandings.guest')}</span>
                    )}
                    {dns && <Badge variant="DNS" />}
                    {onNoShow && !dns && (
                      <button
                        onClick={() => onNoShow(p.groupPlayerId)}
                        style={{ color: '#cbd5e1', marginLeft: 4, lineHeight: 1, fontSize: 12 }}
                        className="hover:text-red-500 transition-colors"
                        title={t('groupStandings.noShow', { name })}
                        aria-label={t('groupStandings.noShow', { name })}
                      >
                        ✕
                      </button>
                    )}
                  </div>
                </td>

                <td style={{ padding: '10px 12px', textAlign: 'center', fontWeight: 700 }}>
                  {p.isNonCalculated ? '—' : p.place > 0 ? p.place : '—'}
                </td>
                <td style={{ padding: '10px 12px', textAlign: 'center' }}>
                  {p.isNonCalculated ? '—' : p.points}
                  (
                  <span style={{ color: '#16a34a'}} >{p.isNonCalculated ? '—' : wins(p)}</span>
                  -
                  <span style={{ color: '#dc2626'}} >{p.isNonCalculated ? '—' : losses(p)}</span>
                  )
                </td>
                <td style={{ padding: '10px 12px', textAlign: 'center', color: '#64748b' }}>
                  {p.isNonCalculated ? '—' : (tiebreakMap.get(p.groupPlayerId) ?? '—')}
                </td>
                <td style={{ padding: '10px 12px', textAlign: 'center' }}>
                  {p.advances && !p.isNonCalculated && (
                    <span style={{ color: '#16a34a', fontWeight: 700, fontSize: 16 }} title="Advances">↑</span>
                  )}
                  {p.recedes && !p.isNonCalculated && (
                    <span style={{ color: '#dc2626', fontWeight: 700, fontSize: 16 }} title="Recedes">↓</span>
                  )}
                </td>
                {sorted.map((colPlayer) => {
                  if (colPlayer.groupPlayerId === p.groupPlayerId) {
                    return (
                        <td
                            key={colPlayer.groupPlayerId}
                            style={{...tdStyle, backgroundColor: '#94a3b8' }}
                        >
                          ×
                        </td>
                    )
                  }
                  const content = cellContent(p, colPlayer)
                  const clickable = !!onScoreClick

                  const m = getMatch(p, colPlayer)
                  const inProgress = m?.status === 'IN_PROGRESS'

                  return (
                    <td
                      key={colPlayer.groupPlayerId}
                      style={{
                        ...tdStyle,
                        backgroundColor: inProgress ? '#fef9c3' : undefined,
                        color:
                          content === '—'
                            ? '#94a3b8'
                            : p1wins(p, colPlayer)
                              ? '#16a34a'
                              : '#dc2626',
                      }}
                    >
                      {inProgress && m?.tableNumber !== null && m?.tableNumber !== undefined && (
                        <span style={{
                          display: 'block',
                          fontSize: 9,
                          fontWeight: 700,
                          color: '#92400e',
                          lineHeight: 1,
                          marginBottom: 2,
                        }}>
                          T{m.tableNumber}
                        </span>
                      )}
                      {clickable && onScoreClick ? (
                        <button
                          onClick={() => m && onScoreClick(m)}
                          aria-label={t('groupStandings.enterScore', { p1: playerName(p), p2: playerName(colPlayer) })}
                        >
                          {content}
                        </button>
                      ) : (
                        <span>{content}</span>
                      )}

                    </td>
                  )
                })}
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
