import { useCallback, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { formatDate } from '../hooks/utils'
import { useEvent, useFinishEvent, useUpdateEventDetails } from '../hooks/useEvents'
import {
  useResetMatchScore,
  useSetTableNumber,
  useTablesInUse,
  useUpdateMatchScore,
} from '../hooks/useMatches'
import { useLeague } from '../hooks/useLeagues'
import {
  useFinishGroup,
  useReopenGroup,
  useSetManualPlace,
  useSetPlayerStatus,
} from '../hooks/useGroups'
import { useEventWebSocket } from '../hooks/useWebSocket'
import { useAuth } from '../hooks/useAuth'
import { GroupCard } from '../components/GroupCard/GroupCard'
import { GroupStandings } from '../components/GroupStandings/GroupStandings'
import { MatchGrid } from '../components/MatchGrid/MatchGrid'
import { Modal } from '../components/Modal/Modal'
import { ScoreEntryForm } from '../components/ScoreEntryForm/ScoreEntryForm'
import { TableAssignModal } from '../components/TableAssignModal/TableAssignModal'
import { PlacementOverride } from '../components/PlacementOverride/PlacementOverride'
import { Button } from '../components/Button/Button'
import { Badge } from '../components/Badge/Badge'
import type { EventDetail, GroupDetail, GroupPlayer, Match, WSMessage } from '../types'
import { groupTitle, groupSortKey } from '../utils/group'

export function LiveViewPage() {
  const { t } = useTranslation()
  const { id, eid } = useParams<{ id: string; eid: string }>()
  const leagueId = Number(id)
  const eventId = Number(eid)

  const { event, setEvent, loading, error, refresh: refreshEvent } = useEvent(leagueId, eventId)
  const { league } = useLeague(leagueId)
  const { isMaintainer, isUmpire } = useAuth()

  const { update: updateScore, loading: scoreSaving } = useUpdateMatchScore()
  const { reset: resetScore, loading: scoreResetting } = useResetMatchScore()
  const { assign: assignTable, loading: tableSaving } = useSetTableNumber()
  const { tablesInUse, refresh: refreshTablesInUse } = useTablesInUse(eventId)
  const { finish: finishGroup, loading: finishing } = useFinishGroup()
  const { reopen: reopenGroup, loading: reopening } = useReopenGroup()
  const { setPlace, loading: placing } = useSetManualPlace()
  const { setStatus: setPlayerStatus } = useSetPlayerStatus()
  const {
    finish: finishEventAction,
    loading: finishingEvent,
    error: finishEventError,
  } = useFinishEvent()
  const { update: updateDetails, loading: savingDetails, error: detailsError } = useUpdateEventDetails()

  const [scoreModal, setScoreModal] = useState<{
    match: Match
    groupId: number
    player1Name: string
    player2Name: string
  } | null>(null)
  const [tableModal, setTableModal] = useState<{
    match: Match
    groupId: number
  } | null>(null)
  const [placementModal, setPlacementModal] = useState<{
    players: GroupPlayer[]
    groupId: number
    eventId: number
  } | null>(null)
  const [dnsConfirm, setDnsConfirm] = useState<{
    groupId: number
    groupPlayerId: number
    playerName: string
    scoredMatchCount: number
  } | null>(null)
  const [collapseSignal, setCollapseSignal] = useState(0)
  const [collapsedDivisions, setCollapsedDivisions] = useState<Set<string>>(new Set())
  const [editingDetails, setEditingDetails] = useState(false)
  const [editTitle, setEditTitle] = useState('')
  const [editStartDate, setEditStartDate] = useState('')
  const [editEndDate, setEditEndDate] = useState('')

  const toggleDivision = (div: string) => {
    setCollapsedDivisions((prev) => {
      const next = new Set(prev)
      if (next.has(div)) next.delete(div)
      else next.add(div)
      return next
    })
  }

  const gamesToWin = league?.configuration?.gamesToWin ?? 3
  const numberOfTables = league?.configuration?.numberOfTables ?? 0

  // WebSocket handler — update local state on live messages.
  const handleWSMessage = useCallback(
    (msg: WSMessage) => {
      if (!event) return

      setEvent((prev: EventDetail | null) => {
        if (!prev) return prev

        if (msg.type === 'match_updated') {
          refreshTablesInUse()
          return {
            ...prev,
            groups: prev.groups.map((g) => {
              if (g.groupId !== msg.groupId) return g
              return {
                ...g,
                matches: g.matches.map((m) => {
                  const payload = msg.payload as {
                    matchId: number
                    score1: number
                    score2: number
                    withdraw1: boolean
                    withdraw2: boolean
                  }
                  if (m.matchId !== payload?.matchId) return m
                  return {
                    ...m,
                    score1: payload.score1 ?? m.score1,
                    score2: payload.score2 ?? m.score2,
                    withdraw1: payload.withdraw1 ?? m.withdraw1,
                    withdraw2: payload.withdraw2 ?? m.withdraw2,
                    status: 'DONE' as const,
                  }
                }),
              }
            }),
          }
        }

        if (msg.type === 'group_finished') {
          return {
            ...prev,
            groups: prev.groups.map((g) =>
              g.groupId === msg.groupId ? { ...g, status: 'DONE' as const } : g
            ),
          }
        }

        if (msg.type === 'event_finished') {
          return { ...prev, status: 'DONE' as const }
        }

        if (msg.type === 'table_assigned') {
          const payload = msg.payload as { matchId: number; tableNumber: number }
          refreshTablesInUse()
          return {
            ...prev,
            groups: prev.groups.map((g) => {
              if (g.groupId !== msg.groupId) return g
              return {
                ...g,
                matches: g.matches.map((m) => {
                  if (m.matchId !== payload?.matchId) return m
                  return { ...m, status: 'IN_PROGRESS' as const, tableNumber: payload.tableNumber }
                }),
              }
            }),
          }
        }

        if (msg.type === 'manual_placement_required') {
          const grp = prev.groups.find((g) => g.groupId === msg.groupId)
          if (grp) {
            const tiedPlayerIds = (msg.payload as { playerIds: number[] })?.playerIds ?? []
            const tiedPlayers = grp.players.filter((p) => tiedPlayerIds.includes(p.groupPlayerId))
            if (tiedPlayers.length > 0) {
              setPlacementModal({ players: tiedPlayers, groupId: grp.groupId, eventId })
            }
          }
        }

        if (msg.type === 'player_dns') {
          const payload = msg.payload as { groupPlayerId: number; deletedMatchIds: number[] }
          return {
            ...prev,
            groups: prev.groups.map((g) => {
              if (g.groupId !== msg.groupId) return g
              return {
                ...g,
                players: g.players.map((p) =>
                  p.groupPlayerId === payload.groupPlayerId ? { ...p, playerStatus: 'dns' as const } : p
                ),
                matches: g.matches.filter((m) => !payload.deletedMatchIds.includes(m.matchId)),
              }
            }),
          }
        }

        if (msg.type === 'player_active') {
          const payload = msg.payload as { groupPlayerId: number; newMatches: Match[] }
          return {
            ...prev,
            groups: prev.groups.map((g) => {
              if (g.groupId !== msg.groupId) return g
              return {
                ...g,
                players: g.players.map((p) =>
                  p.groupPlayerId === payload.groupPlayerId ? { ...p, playerStatus: 'active' as const } : p
                ),
                matches: [...g.matches, ...payload.newMatches],
              }
            }),
          }
        }

        return prev
      })
    },
    [event, eventId, setEvent, refreshTablesInUse]
  )

  const { connected } = useEventWebSocket(eventId, handleWSMessage, {
    enabled: event?.status === 'IN_PROGRESS',
  })

  const handleTableAssignSubmit = async (tableNumber: number) => {
    if (!tableModal) return
    const ok = await assignTable(tableModal.groupId, tableModal.match.matchId, tableNumber)
    if (ok) {
      setTableModal(null)
      refreshTablesInUse()
      setEvent((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          groups: prev.groups.map((g) => {
            if (g.groupId !== tableModal.groupId) return g
            return {
              ...g,
              matches: g.matches.map((m) =>
                m.matchId === tableModal.match.matchId
                  ? { ...m, status: 'IN_PROGRESS' as const, tableNumber }
                  : m
              ),
            }
          }),
        }
      })
    }
  }

  const handleScoreSubmit = async (
    score1: number,
    score2: number,
    withdraw1: boolean,
    withdraw2: boolean
  ) => {
    if (!scoreModal) return
    const ok = await updateScore(scoreModal.groupId, scoreModal.match.matchId, {
      score1,
      score2,
      withdraw1,
      withdraw2,
    })
    if (ok) {
      setScoreModal(null)
      refreshTablesInUse()
      // Optimistically update local state.
      setEvent((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          groups: prev.groups.map((g) => {
            if (g.groupId !== scoreModal.groupId) return g
            return {
              ...g,
              matches: g.matches.map((m) =>
                m.matchId === scoreModal.match.matchId
                  ? { ...m, score1, score2, withdraw1, withdraw2, status: 'DONE' as const }
                  : m
              ),
            }
          }),
        }
      })
    }
  }

  const handleClearScore = async () => {
    if (!scoreModal) return
    const ok = await resetScore(scoreModal.groupId, scoreModal.match.matchId)
    if (ok) {
      setScoreModal(null)
      refreshTablesInUse()
      setEvent((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          groups: prev.groups.map((g) => {
            if (g.groupId !== scoreModal.groupId) return g
            return {
              ...g,
              matches: g.matches.map((m) =>
                m.matchId === scoreModal.match.matchId
                  ? {
                      ...m,
                      score1: null,
                      score2: null,
                      withdraw1: false,
                      withdraw2: false,
                      tableNumber: null,
                      status: 'DRAFT' as const,
                    }
                  : m
              ),
            }
          }),
        }
      })
    }
  }

  const executeDnsToggle = async (
    groupId: number,
    groupPlayerId: number,
    currentStatus: 'active' | 'dns'
  ) => {
    const newStatus = currentStatus === 'dns' ? 'active' : 'dns'
    const result = await setPlayerStatus(eventId, groupId, groupPlayerId, newStatus)
    if (result === null) return

    setEvent((prev) => {
      if (!prev) return prev
      return {
        ...prev,
        groups: prev.groups.map((g) => {
          if (g.groupId !== groupId) return g
          if (newStatus === 'dns') {
            return {
              ...g,
              players: g.players.map((p) =>
                p.groupPlayerId === groupPlayerId ? { ...p, playerStatus: 'dns' } : p
              ),
              matches: g.matches.filter(
                (m) =>
                  !(
                    (m.groupPlayer1Id === groupPlayerId || m.groupPlayer2Id === groupPlayerId) &&
                    result.deletedMatchIds?.includes(m.matchId)
                  )
              ),
            }
          } else {
            return {
              ...g,
              players: g.players.map((p) =>
                p.groupPlayerId === groupPlayerId ? { ...p, playerStatus: 'active' } : p
              ),
              matches: [...g.matches, ...(result.newMatches ?? [])],
            }
          }
        }),
      }
    })
  }

  const handleSetPlayerStatus = async (
    groupId: number,
    groupPlayerId: number,
    currentStatus: 'active' | 'dns'
  ) => {
    const newStatus = currentStatus === 'dns' ? 'active' : 'dns'

    if (newStatus === 'dns') {
      const group = event?.groups.find((g) => g.groupId === groupId)
      if (group) {
        const player = group.players.find((p) => p.groupPlayerId === groupPlayerId)
        const scoredMatches = group.matches.filter(
          (m) =>
            (m.groupPlayer1Id === groupPlayerId || m.groupPlayer2Id === groupPlayerId) &&
            m.status === 'DONE'
        )

        if (scoredMatches.length > 0 && player) {
          setDnsConfirm({
            groupId,
            groupPlayerId,
            playerName: player.user
              ? `${player.user.firstName} ${player.user.lastName}`
              : `#${player.userId}`,
            scoredMatchCount: scoredMatches.length,
          })
          return
        }
      }
    }

    await executeDnsToggle(groupId, groupPlayerId, currentStatus)
  }

  const handleReopenGroup = async (groupId: number) => {
    const ok = await reopenGroup(eventId, groupId)
    if (ok) {
      setEvent((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          groups: prev.groups.map((g) =>
            g.groupId === groupId
              ? {
                  ...g,
                  status: 'IN_PROGRESS' as const,
                  players: g.players.map((p) => ({
                    ...p,
                    points: 0,
                    tiebreakPoints: 0,
                    place: 0,
                    advances: false,
                    recedes: false,
                  })),
                }
              : g
          ),
        }
      })
    }
  }

  const handleFinishGroup = async (groupId: number) => {
    const result = await finishGroup(eventId, groupId)
    if (result) {
      refreshEvent()
    }
  }

  const handleConfirmPlacement = async (orderedPlayerIds: number[]) => {
    if (!placementModal) return
    const ok = await setPlace(placementModal.eventId, placementModal.groupId, orderedPlayerIds)
    if (ok) {
      setPlacementModal(null)
      refreshEvent()
    }
  }

  const handleEditDetails = () => {
    if (!event) return
    setEditTitle(event.title)
    setEditStartDate(event.startDate.substring(0, 10))
    setEditEndDate(event.endDate.substring(0, 10))
    setEditingDetails(true)
  }

  const handleSaveDetails = async () => {
    const result = await updateDetails(leagueId, eventId, {
      title: editTitle,
      startDate: editStartDate,
      endDate: editEndDate,
    })
    if (result) {
      setEvent((prev) =>
        prev
          ? { ...prev, title: result.title, startDate: result.startDate, endDate: result.endDate }
          : prev
      )
      setEditingDetails(false)
    }
  }

  const handleFinishEvent = async () => {
    const result = await finishEventAction(leagueId, eventId)
    if (result) {
      setEvent((prev) => (prev ? { ...prev, status: 'DONE' as const } : prev))
    }
  }

  if (loading) return <div className="p-8 text-gray-400">{t('liveView.loading')}</div>
  if (error) return <div className="p-8 text-red-600">{error}</div>
  if (!event) return null

  const canManage = isMaintainer(leagueId)
  const canUmpire = isUmpire(leagueId)

  const allGroupsDone = event.groups.length > 0 && event.groups.every((g) => g.status === 'DONE')

  const inProgressMatches = event.groups.flatMap((g) =>
    g.matches
      .filter((m) => m.status === 'IN_PROGRESS')
      .map((m) => {
        const p1 = g.players.find((p) => p.groupPlayerId === m.groupPlayer1Id)
        const p2 = g.players.find((p) => p.groupPlayerId === m.groupPlayer2Id)
        return {
          matchId: m.matchId,
          tableNumber: m.tableNumber,
          p1Name: p1?.user ? `${p1.user.firstName} ${p1.user.lastName}` : `#${p1?.userId ?? '?'}`,
          p2Name: p2?.user ? `${p2.user.firstName} ${p2.user.lastName}` : `#${p2?.userId ?? '?'}`,
          division: g.division,
          groupNo: g.groupNo,
        }
      })
  )

  return (
    <div className="max-w-7xl mx-auto py-6 px-4">
      {/* Event header */}
      <div className="flex items-start justify-between mb-6">
        <div style={{ flex: 1, minWidth: 0, marginRight: 16 }}>
          <Link
            to={`/leagues/${leagueId}`}
            className="text-sm text-blue-600 hover:underline block mb-1"
          >
            {t('liveView.backToLeague')}
          </Link>
          {editingDetails ? (
            <div>
              <input
                type="text"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                aria-label={t('liveView.eventTitle')}
                style={{
                  fontSize: 22,
                  fontWeight: 700,
                  color: '#111827',
                  border: '1px solid #d1d5db',
                  borderRadius: 6,
                  padding: '4px 10px',
                  width: '100%',
                  maxWidth: 480,
                  marginBottom: 8,
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                <input
                  type="date"
                  value={editStartDate}
                  onChange={(e) => setEditStartDate(e.target.value)}
                  aria-label={t('liveView.startDate')}
                  style={{
                    fontSize: 13,
                    border: '1px solid #d1d5db',
                    borderRadius: 6,
                    padding: '4px 8px',
                    color: '#374151',
                  }}
                />
                <span style={{ color: '#9ca3af', fontSize: 13 }}>—</span>
                <input
                  type="date"
                  value={editEndDate}
                  onChange={(e) => setEditEndDate(e.target.value)}
                  aria-label={t('liveView.endDate')}
                  style={{
                    fontSize: 13,
                    border: '1px solid #d1d5db',
                    borderRadius: 6,
                    padding: '4px 8px',
                    color: '#374151',
                  }}
                />
                <button
                  onClick={handleSaveDetails}
                  disabled={savingDetails || !editTitle.trim()}
                  style={{
                    padding: '4px 14px',
                    borderRadius: 6,
                    backgroundColor: savingDetails ? '#93c5fd' : '#2563eb',
                    color: '#fff',
                    fontWeight: 600,
                    fontSize: 13,
                    border: 'none',
                    cursor: savingDetails ? 'not-allowed' : 'pointer',
                  }}
                >
                  {t('liveView.saveDetails')}
                </button>
                <button
                  onClick={() => setEditingDetails(false)}
                  disabled={savingDetails}
                  style={{
                    padding: '4px 14px',
                    borderRadius: 6,
                    backgroundColor: '#f3f4f6',
                    color: '#374151',
                    fontWeight: 500,
                    fontSize: 13,
                    border: '1px solid #d1d5db',
                    cursor: 'pointer',
                  }}
                >
                  {t('liveView.cancelEdit')}
                </button>
              </div>
              {detailsError && <p className="text-red-600 text-xs mt-1">{detailsError}</p>}
            </div>
          ) : (
            <div>
              <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                {event.title}
                {event.status === 'IN_PROGRESS' && (
                  <span
                    title={connected ? t('liveView.liveConnected') : t('liveView.liveReconnecting')}
                    style={{
                      display: 'inline-block',
                      width: 10,
                      height: 10,
                      borderRadius: '50%',
                      backgroundColor: connected ? '#22c55e' : '#ef4444',
                      flexShrink: 0,
                      transition: 'background-color 0.3s',
                    }}
                    aria-label={connected ? t('liveView.liveConnected') : t('liveView.liveReconnecting')}
                  />
                )}
                {canManage && (
                  <button
                    onClick={handleEditDetails}
                    title={t('liveView.editDetails')}
                    style={{
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      padding: '2px 4px',
                      color: '#9ca3af',
                      display: 'inline-flex',
                      alignItems: 'center',
                    }}
                  >
                    <svg width="15" height="15" viewBox="0 0 15 15" fill="none" aria-hidden="true">
                      <path
                        d="M10.5 1.5l3 3-9 9H1.5v-3l9-9z"
                        stroke="currentColor"
                        strokeWidth="1.4"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  </button>
                )}
              </h1>
              <p className="text-sm text-gray-500 mt-1">
                {formatDate(event.startDate)} — {formatDate(event.endDate)}
              </p>
            </div>
          )}
          {finishEventError && <p className="text-red-600 text-xs mt-1">{finishEventError}</p>}
        </div>
        <div className="flex items-center gap-3">
          {(canManage || canUmpire) && (
            <Button variant="secondary" onClick={() => setCollapseSignal((s) => s + 1)}>
              {t('liveView.collapseAll')}
            </Button>
          )}
          {canManage && event.status === 'IN_PROGRESS' && (
            <Link
              to={`/leagues/${leagueId}/events/${eventId}/setup`}
              className="inline-flex items-center px-3 py-1.5 rounded text-sm font-medium bg-white border border-gray-300 text-gray-700 hover:bg-gray-50"
            >
              {t('liveView.managePlayers', 'Manage players')}
            </Link>
          )}
          {canManage && event.status === 'IN_PROGRESS' && allGroupsDone && (
            <Button variant="primary" onClick={handleFinishEvent} loading={finishingEvent}>
              {t('liveView.finishEvent')}
            </Button>
          )}
          <Badge variant={event.status} />
        </div>
      </div>

      {/* In-progress matches panel */}
      {inProgressMatches.length > 0 && (
        <div style={{ marginBottom: 24, border: '1px solid var(--border)', background: '#fff' }}>
          <div
            style={{
              padding: '6px 14px',
              background: '#f8fafc',
              borderBottom: '1px solid var(--border)',
              display: 'flex',
              alignItems: 'center',
              gap: 7,
            }}
          >
            <span
              style={{
                fontSize: 10,
                fontWeight: 700,
                letterSpacing: '0.09em',
                textTransform: 'uppercase',
                color: '#94a3b8',
              }}
            >
              {t('liveView.inProgress')}
            </span>
            <span
              style={{
                fontSize: 11,
                fontWeight: 700,
                color: '#fff',
                background: '#f59e0b',
                borderRadius: 4,
                padding: '1px 8px',
              }}
            >
              {inProgressMatches.length}
            </span>
          </div>
          <div className="overflow-x-auto">
            <table style={{ minWidth: 360, width: '100%', borderCollapse: 'collapse' }}>
              <tbody>
                {inProgressMatches.map((m, i) => (
                  <tr key={m.matchId} style={{ borderTop: i === 0 ? 'none' : '1px solid #f1f5f9' }}>
                    <td style={{ padding: '10px 14px', width: 72 }}>
                      {m.tableNumber != null ? (
                        <span
                          style={{
                            fontWeight: 700,
                            color: '#fff',
                            background: '#f59e0b',
                            borderRadius: 4,
                            padding: '2px 8px',
                            fontSize: 12,
                          }}
                        >
                          T{m.tableNumber}
                        </span>
                      ) : (
                        <span style={{ color: '#94a3b8', fontSize: 12 }}>—</span>
                      )}
                    </td>
                    <td
                      style={{
                        padding: '10px 6px',
                        fontWeight: 500,
                        fontSize: 13,
                        color: 'var(--navy)',
                      }}
                    >
                      {m.p1Name}
                    </td>
                    <td
                      style={{
                        padding: '10px 4px',
                        fontSize: 10,
                        fontWeight: 700,
                        color: '#e2e8f0',
                        textAlign: 'center',
                        userSelect: 'none',
                      }}
                    >
                      vs
                    </td>
                    <td
                      style={{
                        padding: '10px 6px',
                        fontWeight: 500,
                        fontSize: 13,
                        color: 'var(--navy)',
                      }}
                    >
                      {m.p2Name}
                    </td>
                    <td
                      style={{
                        padding: '10px 14px',
                        textAlign: 'right',
                        fontSize: 11,
                        color: '#94a3b8',
                      }}
                    >
                      {groupTitle(m.division, m.groupNo)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Groups grid */}
      <div>
        {(() => {
          const sorted = [...event.groups].sort(
            (a, b) => groupSortKey(a.division, a.groupNo) - groupSortKey(b.division, b.groupNo)
          )
          // Collect divisions in order of first appearance
          const divisionOrder: string[] = []
          const byDivision: Record<string, GroupDetail[]> = {}
          for (const g of sorted) {
            if (!byDivision[g.division]) {
              divisionOrder.push(g.division)
              byDivision[g.division] = []
            }
            byDivision[g.division].push(g)
          }
          return divisionOrder.map((div) => {
            const divGroups = byDivision[div]
            const isCollapsed = collapsedDivisions.has(div)
            const divLabel = div === 'S' ? t('groupCard.superleague') : t('groupCard.divisionLabel', { div })
            return (
              <div key={div} style={{ marginBottom: 16 }}>
                {/* Division header — visually distinct from group cards */}
                <button
                  onClick={() => toggleDivision(div)}
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    padding: '9px 14px 9px 0',
                    backgroundColor: 'transparent',
                    border: 'none',
                    borderBottom: isCollapsed ? '2px solid #f59e0b' : '2px solid #f59e0b',
                    cursor: 'pointer',
                    marginBottom: isCollapsed ? 0 : 8,
                    userSelect: 'none',
                  }}
                >
                  {/* Amber accent block */}
                  <span style={{
                    display: 'inline-block',
                    width: 4,
                    height: 28,
                    backgroundColor: '#f59e0b',
                    borderRadius: 2,
                    flexShrink: 0,
                    marginLeft: 0,
                  }} />
                  {/* Chevron SVG — clearly distinct from GroupCard's ▾ */}
                  <svg
                    width="13"
                    height="13"
                    viewBox="0 0 13 13"
                    fill="none"
                    style={{
                      transform: isCollapsed ? 'rotate(-90deg)' : 'rotate(0deg)',
                      transition: 'transform 0.2s',
                      flexShrink: 0,
                    }}
                  >
                    <path
                      d="M2 4.5l4.5 4.5 4.5-4.5"
                      stroke="#f59e0b"
                      strokeWidth="1.8"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  {/* Division label — uppercase tracked, amber */}
                  <span style={{
                    color: '#f59e0b',
                    fontWeight: 700,
                    fontSize: 11,
                    letterSpacing: '0.1em',
                    textTransform: 'uppercase',
                    flex: 1,
                    textAlign: 'left',
                  }}>
                    {divLabel}
                  </span>
                  {/* Group count pill */}
                  <span style={{
                    backgroundColor: 'rgba(245,158,11,0.12)',
                    color: '#d97706',
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '2px 9px',
                    borderRadius: 10,
                    border: '1px solid rgba(245,158,11,0.28)',
                    letterSpacing: '0.01em',
                  }}>
                    {t('groupCard.groups', { count: divGroups.length })}
                  </span>
                </button>
                {/* Groups within division — indented with faint amber connector */}
                {!isCollapsed && (
                  <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 8,
                    paddingBottom: 4,
                    paddingLeft: 12,
                    borderLeft: '2px solid rgba(245,158,11,0.2)',
                    marginLeft: 2,
                  }}>
                    {divGroups.map((group: GroupDetail) => (
                      <GroupCard
                        key={group.groupId}
                        division={group.division}
                        groupNo={group.groupNo}
                        status={group.status}
                        collapsible={canManage || canUmpire}
                        defaultCollapsed={canManage || canUmpire}
                        collapseSignal={collapseSignal}
                        groupViewUrl={`/leagues/${leagueId}/events/${eventId}/groups/${group.groupId}`}
                      >
                        {/* Standings */}
                        <div className="mb-4">
                          <GroupStandings
                            players={group.players}
                            matches={group.matches}
                            onSetPlayerStatus={
                              canManage && group.status !== 'DONE'
                                ? (gpId, currentStatus) =>
                                    handleSetPlayerStatus(group.groupId, gpId, currentStatus)
                                : undefined
                            }
                            onScoreClick={
                              canUmpire && group.status !== 'DONE'
                                ? (m) => {
                                    if (m.status === 'DRAFT' && numberOfTables > 0) {
                                      setTableModal({ match: m, groupId: group.groupId })
                                      return
                                    }
                                    const p1 = group.players.find((p) => p.groupPlayerId === m.groupPlayer1Id)
                                    const p2 = group.players.find((p) => p.groupPlayerId === m.groupPlayer2Id)
                                    const p1Name = p1?.user
                                      ? `${p1.user.firstName} ${p1.user.lastName}`
                                      : `#${p1?.userId}`
                                    const p2Name = p2?.user
                                      ? `${p2.user.firstName} ${p2.user.lastName}`
                                      : `#${p2?.userId}`
                                    setScoreModal({
                                      match: m,
                                      groupId: group.groupId,
                                      player1Name: p1Name,
                                      player2Name: p2Name,
                                    })
                                  }
                                : undefined
                            }
                          />
                        </div>

                        {/* Match grid */}
                        {canManage || canUmpire ? (
                          <>
                            <div className="mb-4">
                              <p className="text-xs uppercase text-gray-400 font-medium mb-2">
                                {t('liveView.matchResults')}
                              </p>
                              <MatchGrid
                                players={group.players}
                                matches={group.matches}
                                onScoreClick={
                                  canUmpire && group.status !== 'DONE'
                                    ? (m) => {
                                        if (m.status === 'DRAFT' && numberOfTables > 0) {
                                          setTableModal({ match: m, groupId: group.groupId })
                                          return
                                        }
                                        const p1 = group.players.find(
                                          (p) => p.groupPlayerId === m.groupPlayer1Id
                                        )
                                        const p2 = group.players.find(
                                          (p) => p.groupPlayerId === m.groupPlayer2Id
                                        )
                                        const p1Name = p1?.user
                                          ? `${p1.user.firstName} ${p1.user.lastName}`
                                          : `#${p1?.userId}`
                                        const p2Name = p2?.user
                                          ? `${p2.user.firstName} ${p2.user.lastName}`
                                          : `#${p2?.userId}`
                                        setScoreModal({
                                          match: m,
                                          groupId: group.groupId,
                                          player1Name: p1Name,
                                          player2Name: p2Name,
                                        })
                                      }
                                    : undefined
                                }
                              />
                            </div>

                            {/* Actions: show only if canManage or canUmpire */}
                            <div className="flex flex-wrap gap-2 mt-3">
                              {canManage && group.status === 'DONE' && event.status !== 'DONE' && (
                                <Button
                                  variant="secondary"
                                  onClick={() => handleReopenGroup(group.groupId)}
                                  loading={reopening}
                                >
                                  {t('liveView.reopenGroup')}
                                </Button>
                              )}
                              {canUmpire &&
                                group.status !== 'DONE' &&
                                (() => {
                                  const activePlayers = group.players.filter(
                                    (p) => !p.isNonCalculated && p.playerStatus !== 'dns'
                                  )
                                  const allScored =
                                    group.players.length > 0 &&
                                    (activePlayers.length <= 1 ||
                                      (group.matches.length > 0 && group.matches.every((m) => m.status === 'DONE')))
                                  return (
                                    <Button
                                      variant="primary"
                                      onClick={() => handleFinishGroup(group.groupId)}
                                      loading={finishing}
                                      disabled={!allScored}
                                      title={!allScored ? t('liveView.enterAllScoresFirst') : undefined}
                                    >
                                      {t('liveView.finishGroup')}
                                    </Button>
                                  )
                                })()}
                            </div>
                          </>
                        ) : (
                          ''
                        )}
                      </GroupCard>
                    ))}
                  </div>
                )}
              </div>
            )
          })
        })()}
      </div>

      {/* Score entry modal */}
      <Modal
        open={!!scoreModal}
        onClose={() => setScoreModal(null)}
        title={t('liveView.enterScoreTitle')}
      >
        {scoreModal && (
          <ScoreEntryForm
            match={scoreModal.match}
            gamesToWin={gamesToWin}
            player1Name={scoreModal.player1Name}
            player2Name={scoreModal.player2Name}
            onSubmit={handleScoreSubmit}
            onClear={handleClearScore}
            onClose={() => setScoreModal(null)}
            loading={scoreSaving || scoreResetting}
          />
        )}
      </Modal>

      {/* Table assign modal */}
      <Modal
        open={!!tableModal}
        onClose={() => setTableModal(null)}
        title={t('liveView.assignTableTitle')}
      >
        {tableModal && (
          <TableAssignModal
            match={tableModal.match}
            numberOfTables={numberOfTables}
            tablesInUse={tablesInUse}
            onSubmit={handleTableAssignSubmit}
            onClose={() => setTableModal(null)}
            loading={tableSaving}
          />
        )}
      </Modal>

      {/* Placement override modal */}
      <Modal
        open={!!placementModal}
        onClose={() => setPlacementModal(null)}
        title={t('liveView.manualPlacementTitle')}
      >
        {placementModal && (
          <div>
            <p className="text-sm text-gray-600 mb-4">{t('liveView.manualPlacementDescription')}</p>
            <PlacementOverride
              players={placementModal.players}
              onConfirm={handleConfirmPlacement}
              onClose={() => setPlacementModal(null)}
              loading={placing}
            />
          </div>
        )}
      </Modal>

      {/* DNS confirmation dialog */}
      {dnsConfirm && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            backgroundColor: 'rgba(0, 0, 0, 0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 50,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              borderRadius: '8px',
              boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
              padding: '24px',
              maxWidth: '400px',
              width: '90%',
            }}
          >
            <h2 style={{ fontSize: '18px', fontWeight: 700, marginBottom: '12px', color: '#1f2937' }}>
              {t('liveView.dnsConfirmTitle')}
            </h2>
            <p style={{ fontSize: '14px', color: '#6b7280', marginBottom: '20px' }}>
              {t('liveView.dnsConfirmBody', { name: dnsConfirm.playerName, count: dnsConfirm.scoredMatchCount })}
            </p>
            <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
              <button
                onClick={() => setDnsConfirm(null)}
                style={{
                  padding: '8px 16px',
                  borderRadius: '6px',
                  border: '1px solid #d1d5db',
                  backgroundColor: '#f9fafb',
                  color: '#374151',
                  fontWeight: 500,
                  cursor: 'pointer',
                  fontSize: '14px',
                }}
              >
                {t('scoreEntry.cancel')}
              </button>
              <button
                onClick={async () => {
                  await executeDnsToggle(dnsConfirm.groupId, dnsConfirm.groupPlayerId, 'active')
                  setDnsConfirm(null)
                }}
                style={{
                  padding: '8px 16px',
                  borderRadius: '6px',
                  backgroundColor: '#dc2626',
                  color: '#fff',
                  fontWeight: 500,
                  cursor: 'pointer',
                  fontSize: '14px',
                }}
              >
                {t('liveView.confirmDns')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
