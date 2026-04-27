import { type CSSProperties } from 'react'
import type { GroupPlayer, Match } from '../../types'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

interface MatchGridProps {
  players: GroupPlayer[]
  matches: Match[]
  onScoreClick?: (match: Match) => void
}

interface Round {
  id: number
  matches: Match[]
}

// ─── Score pill helpers ───────────────────────────────────────────────────────

function getRoundMatchState(m: Match): 'pending' | 'done' | 'walkover' {
  if (m.status !== 'DONE') return 'pending'
  if (m.withdraw1 || m.withdraw2) return 'walkover'
  return 'done'
}

function getRoundScoreLabel(m: Match): string {
  const state = getRoundMatchState(m)
  if (state === 'pending') return '—'
  if (state === 'walkover') {
    if (m.withdraw1 && m.withdraw2) return 'W/O'
    return m.withdraw1 ? 'L-W' : 'W-L'  // from player1's perspective
  }
  return m.score1 !== null && m.score2 !== null ? `${m.score1} : ${m.score2}` : '—'
}

// ─── Component ────────────────────────────────────────────────────────────────

export function MatchGrid({ players, matches, onScoreClick }: MatchGridProps) {
  const { t } = useTranslation()
  const calcPlayers = [...players.filter((p) => !p.isNonCalculated)].sort((a, b) => a.seed - b.seed)

  const matchLookup = new Map<string, Match>()
  for (const m of matches) {
    if (m.groupPlayer1Id !== null && m.groupPlayer2Id !== null) {
      matchLookup.set(`${m.groupPlayer1Id}-${m.groupPlayer2Id}`, m)
    }
  }

  const getMatchById = (id1: number, id2: number): Match | undefined =>
    matchLookup.get(`${id1}-${id2}`) ?? matchLookup.get(`${id2}-${id1}`)

  const getPlayerById = (id: number | null): GroupPlayer | undefined =>
    calcPlayers.find((p) => p.groupPlayerId === id)

  const playerName = (p: GroupPlayer) =>
    p.user ? `${p.user.firstName} ${p.user.lastName}` : `#${p.userId}`

  const getRounds = (ps: GroupPlayer[]): Round[] => {
    let n = ps.length
    if (n < 2) return []
    let seats = ps.map((p) => p.groupPlayerId)
    if (n % 2 !== 0) { seats.push(-1); n++ }
    const rounds: Round[] = []
    for (let r = 0; r < n - 1; r++) {
      const round: Round = { id: r + 1, matches: [] }
      for (let i = 0; i < n / 2; i++) {
        const a = seats[i], b = seats[n - 1 - i]
        if (a !== -1 && b !== -1) {
          const m = getMatchById(a, b)
          if (m) round.matches.push(m)
        }
      }
      seats = [seats[0], seats[n - 1], ...seats.slice(1, n - 1)]
      rounds.push(round)
    }
    return rounds
  }

  if (calcPlayers.length === 0) {
    return <p style={{ fontSize: 13, color: '#94a3b8', padding: '4px 0' }}>{t('matchGrid.noPlayers')}</p>
  }

  const rounds = getRounds(calcPlayers)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>

      {/* ── Round tables ─────────────────────────────────────────────────── */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {rounds.map((r) => (
          <div
            key={r.id}
            style={{
              border: '1px solid var(--border)',
              overflow: 'hidden',
              background: '#fff',
            }}
          >
            {/* Round header strip */}
            <div style={{
              padding: '6px 14px',
              background: '#f8fafc',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              gap: 7,
            }}>
              <span style={{
                fontSize: 10,
                fontWeight: 700,
                letterSpacing: '0.09em',
                textTransform: 'uppercase',
                color: '#94a3b8',
              }}>
                {t('matchGrid.round')}
              </span>
              <span style={{
                fontSize: 11,
                fontWeight: 700,
                color: '#fff',
                background: 'var(--navy)',
                borderRadius: 4,
                padding: '1px 8px',
                letterSpacing: '0.02em',
              }}>
                {r.id}
              </span>
            </div>

            {/* Rows */}
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <tbody>
                {r.matches.map((m, idx) => {
                  const p1 = getPlayerById(m.groupPlayer1Id)
                  const p2 = getPlayerById(m.groupPlayer2Id)
                  if (!p1 || !p2) return null

                  const state = getRoundMatchState(m)
                  const clickable = !!onScoreClick
                  const label = getRoundScoreLabel(m)

                  // Score pill appearance
                  let pillBg = 'transparent'
                  let pillColor = clickable ? '#FF7A00' : '#94a3b8'
                  let pillBorder = `1.5px dashed ${clickable ? '#FF7A00' : '#cbd5e1'}`
                  let pillWeight = 500

                  if (state === 'done') {
                    pillBg = '#f1f5f9'
                    pillColor = '#334155'
                    pillBorder = 'none'
                    pillWeight = 700
                  } else if (state === 'walkover') {
                    pillBg = '#fef3c7'
                    pillColor = '#92400e'
                    pillBorder = 'none'
                    pillWeight = 600
                  }

                  const pillStyle: CSSProperties = {
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    minWidth: 62,
                    padding: '3px 10px',
                    borderRadius: 6,
                    fontSize: 12,
                    fontWeight: pillWeight,
                    fontVariantNumeric: 'tabular-nums',
                    background: pillBg,
                    color: pillColor,
                    border: pillBorder,
                    cursor: clickable ? 'pointer' : 'default',
                    transition: 'opacity 0.1s',
                    whiteSpace: 'nowrap',
                  }

                  const nameLinkStyle: CSSProperties = {
                    color: 'var(--navy)',
                    fontWeight: 500,
                    fontSize: 13,
                    textDecoration: 'none',
                  }

                  return (
                    <tr
                      key={m.matchId}
                      style={{ borderTop: idx === 0 ? 'none' : '1px solid #f1f5f9' }}
                    >
                      {/* Index */}
                      <td style={{
                        width: 32,
                        padding: '10px 0 10px 14px',
                        textAlign: 'center',
                        fontSize: 11,
                        fontWeight: 600,
                        color: '#e2e8f0',
                        verticalAlign: 'middle',
                      }}>
                        {idx + 1}
                      </td>

                      {/* Player 1 */}
                      <td style={{ padding: '10px 6px', verticalAlign: 'middle', width: '38%' }}>
                        <Link to={`/players/${p1.userId}`} style={nameLinkStyle} className="hover:underline">
                          {playerName(p1)}
                        </Link>
                      </td>

                      {/* vs */}
                      <td style={{
                        width: 22,
                        textAlign: 'center',
                        fontSize: 10,
                        fontWeight: 700,
                        color: '#e2e8f0',
                        letterSpacing: '0.05em',
                        verticalAlign: 'middle',
                        padding: '10px 2px',
                        userSelect: 'none',
                      }}>
                        vs
                      </td>

                      {/* Player 2 */}
                      <td style={{ padding: '10px 6px', verticalAlign: 'middle', width: '38%' }}>
                        <Link to={`/players/${p2.userId}`} style={nameLinkStyle} className="hover:underline">
                          {playerName(p2)}
                        </Link>
                      </td>

                      {/* Score pill */}
                      <td style={{ padding: '10px 14px 10px 4px', textAlign: 'right', verticalAlign: 'middle' }}>
                        {clickable && onScoreClick ? (
                          <button
                            onClick={() => onScoreClick(m)}
                            style={pillStyle}
                            aria-label={t('matchGrid.enterScore', { p1: playerName(p1), p2: playerName(p2) })}
                          >
                            {label}
                          </button>
                        ) : (
                          <span style={pillStyle}>{label}</span>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        ))}
      </div>

    </div>
  )
}
