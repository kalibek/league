import { useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { formatDate } from '../hooks/utils'
import { useEvent, useFinishEvent } from '../hooks/useEvents'
import { useUpdateMatchScore } from '../hooks/useMatches'
import { useFinishGroup, useReopenGroup, useSetManualPlace } from '../hooks/useGroups'
import { useEventWebSocket } from '../hooks/useWebSocket'
import { useAuth } from '../hooks/useAuth'
import { GroupCard } from '../components/GroupCard/GroupCard'
import { GroupStandings } from '../components/GroupStandings/GroupStandings'
import { MatchGrid } from '../components/MatchGrid/MatchGrid'
import { Modal } from '../components/Modal/Modal'
import { ScoreEntryForm } from '../components/ScoreEntryForm/ScoreEntryForm'
import { PlacementOverride } from '../components/PlacementOverride/PlacementOverride'
import { Button } from '../components/Button/Button'
import { Badge } from '../components/Badge/Badge'
import type { EventDetail, GroupDetail, Match, GroupPlayer, WSMessage } from '../types'

export function LiveViewPage() {
  const { t } = useTranslation()
  const { id, eid } = useParams<{ id: string; eid: string }>()
  const leagueId = Number(id)
  const eventId = Number(eid)

  const { event, setEvent, loading, error, refresh: refreshEvent } = useEvent(leagueId, eventId)
  const { isMaintainer, isUmpire } = useAuth()

  const { update: updateScore, loading: scoreSaving } = useUpdateMatchScore()
  const { finish: finishGroup, loading: finishing } = useFinishGroup()
  const { reopen: reopenGroup, loading: reopening } = useReopenGroup()
  const { setPlace, loading: placing } = useSetManualPlace()
  const {
    finish: finishEventAction,
    loading: finishingEvent,
    error: finishEventError,
  } = useFinishEvent()

  const [scoreModal, setScoreModal] = useState<{
    match: Match
    groupId: number
    player1Name: string
    player2Name: string
  } | null>(null)
  const [placementModal, setPlacementModal] = useState<{
    players: GroupPlayer[]
    groupId: number
    eventId: number
  } | null>(null)
  const [collapseSignal, setCollapseSignal] = useState(0)

  const gamesToWin = event?.groups?.[0]
    ? 3 // will be provided by event data in production
    : 3

  // WebSocket handler — update local state on live messages.
  const handleWSMessage = useCallback(
    (msg: WSMessage) => {
      if (!event) return

      setEvent((prev: EventDetail | null) => {
        if (!prev) return prev

        if (msg.type === 'match_updated') {
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

        return prev
      })
    },
    [event, eventId, setEvent]
  )

  const { connected } = useEventWebSocket(eventId, handleWSMessage, {
    enabled: event?.status === 'IN_PROGRESS',
  })

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

  const handleMarkNoShow = async (groupId: number, gpId: number) => {
    const groupMatches = event?.groups.find((g) => g.groupId === groupId)?.matches ?? []
    for (const match of groupMatches.filter(
      (m) => m.groupPlayer1Id === gpId || m.groupPlayer2Id === gpId
    )) {
      if (!match) continue
      let ok: Match | null = null
      let [ score1, score2, withdraw1, withdraw2 ] = [0, 0, false, false];
      if (match.groupPlayer1Id === gpId) {
          score1 =  0
          score2 = gamesToWin
          withdraw1 = true
          withdraw2 = false
      }
      if (match.groupPlayer2Id === gpId) {
        score1 = gamesToWin
        score2 = 0
        withdraw1 = false
        withdraw2 = true
      }
      ok = await updateScore(groupId, match.matchId, {
        score1,
        score2,
        withdraw1,
        withdraw2,
      })
      if (ok) {
        // Optimistically update local state.
        setEvent((prev) => {
          if (!prev) return prev
          return {
            ...prev,
            groups: prev.groups.map((g) => {
              if (g.groupId !== groupId) return g
              return {
                ...g,
                matches: g.matches.map((m) =>
                  m.matchId === match.matchId
                    ? { ...m, score1, score2, withdraw1, withdraw2, status: 'DONE' as const }
                    : m
                ),
              }
            }),
          }
        })
      }
    }
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

  return (
    <div className="max-w-7xl mx-auto py-6 px-4">
      {/* Event header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <Link
            to={`/leagues/${leagueId}`}
            className="text-sm text-blue-600 hover:underline block mb-1"
          >
            {t('liveView.backToLeague')}
          </Link>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            {event.title}
            {event.status === 'IN_PROGRESS' && (
              <span
                title={connected ? 'Live: connected' : 'Live: reconnecting…'}
                style={{
                  display: 'inline-block',
                  width: 10,
                  height: 10,
                  borderRadius: '50%',
                  backgroundColor: connected ? '#22c55e' : '#ef4444',
                  flexShrink: 0,
                  transition: 'background-color 0.3s',
                }}
                aria-label={connected ? 'WebSocket connected' : 'WebSocket disconnected'}
              />
            )}
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            {formatDate(event.startDate)} — {formatDate(event.endDate)}
          </p>
          {finishEventError && <p className="text-red-600 text-xs mt-1">{finishEventError}</p>}
        </div>
        <div className="flex items-center gap-3">
          {(canManage || canUmpire) && (
            <Button variant="secondary" onClick={() => setCollapseSignal((s) => s + 1)}>
              {t('liveView.collapseAll')}
            </Button>
          )}
          {canManage && event.status === 'IN_PROGRESS' && allGroupsDone && (
            <Button variant="primary" onClick={handleFinishEvent} loading={finishingEvent}>
              {t('liveView.finishEvent')}
            </Button>
          )}
          <Badge variant={event.status} />
        </div>
      </div>

      {/* Groups grid */}
      <div>
        {event.groups.map((group: GroupDetail) => (
          <GroupCard
            key={group.groupId}
            division={group.division}
            groupNo={group.groupNo}
            status={group.status}
            collapsible={canManage || canUmpire}
            defaultCollapsed={canManage || canUmpire}
            collapseSignal={collapseSignal}
          >
            {/* Standings */}
            <div className="mb-4">
              <GroupStandings
                players={group.players}
                matches={group.matches}
                onNoShow={
                  canManage && group.status !== 'DONE'
                    ? (gpId) => handleMarkNoShow(group.groupId, gpId)
                    : undefined
                }
                onScoreClick={
                  canUmpire && group.status !== 'DONE'
                    ? (m) => {
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
                      const allScored =
                        group.matches.length > 0 && group.matches.every((m) => m.status === 'DONE')
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
            onClose={() => setScoreModal(null)}
            loading={scoreSaving}
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
    </div>
  )
}
