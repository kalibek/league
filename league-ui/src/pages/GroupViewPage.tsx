import { useCallback, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useGroup } from '../hooks/useGroups'
import { useEventWebSocket } from '../hooks/useWebSocket'
import { GroupStandings } from '../components/GroupStandings/GroupStandings'
import { MatchGrid } from '../components/MatchGrid/MatchGrid'
import { Badge } from '../components/Badge/Badge'
import type { GroupDetail, WSMessage } from '../types'
import { groupTitle } from '../utils/group'

export function GroupViewPage() {
  const { t } = useTranslation()
  const { id, eid, gid } = useParams<{ id: string; eid: string; gid: string }>()
  const leagueId = Number(id)
  const eventId = Number(eid)
  const groupId = Number(gid)

  const { group: fetchedGroup, loading, error } = useGroup(eventId, groupId)
  const [wsGroup, setWsGroup] = useState<GroupDetail | null>(null)

  const group = wsGroup ?? fetchedGroup ?? null

  // WebSocket handler — update local state on live messages.
  const handleWSMessage = useCallback(
    (msg: WSMessage) => {
      if (msg.groupId !== groupId) return

      setWsGroup((prev) => {
        const base = prev ?? fetchedGroup
        if (!base) return prev

        if (msg.type === 'match_updated') {
          return {
            ...base,
            matches: base.matches.map((m) => {
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
        }

        if (msg.type === 'group_finished') {
          return { ...base, status: 'DONE' as const }
        }

        return prev
      })
    },
    [groupId, fetchedGroup]
  )

  const { connected } = useEventWebSocket(eventId, handleWSMessage, {
    enabled: group?.status === 'IN_PROGRESS',
  })

  if (loading) return <div className="p-8 text-gray-400">{t('liveView.loading')}</div>
  if (error) return <div className="p-8 text-red-600">{error}</div>
  if (!group) return null

  const title = groupTitle(group.division, group.groupNo)
  const divLabel = group.division === 'S' ? 'Superleague' : `Division ${group.division}`

  return (
    <div className="max-w-4xl mx-auto py-6 px-4">
      {/* Back link */}
      <Link
        to={`/leagues/${leagueId}/events/${eventId}`}
        className="text-sm text-blue-600 hover:underline block mb-4"
      >
        ← {t('liveView.backToEvent', 'Back to event')}
      </Link>

      {/* Group header */}
      <div className="mb-6">
        <div
          className="flex items-center gap-3 px-4 py-3 rounded-t-lg"
          style={{ backgroundColor: 'var(--navy)' }}
        >
          {/* Amber accent block */}
          <span
            style={{
              display: 'inline-block',
              width: 4,
              height: 28,
              backgroundColor: '#f59e0b',
              borderRadius: 2,
              flexShrink: 0,
            }}
          />
          {/* Division label */}
          <span
            style={{
              color: '#f59e0b',
              fontWeight: 700,
              fontSize: 11,
              letterSpacing: '0.1em',
              textTransform: 'uppercase',
            }}
          >
            {divLabel}
          </span>
          {/* Title */}
          <h1 style={{ color: '#fff', fontWeight: 600, fontSize: 18, flex: 1, marginLeft: 8 }}>
            {title}
          </h1>
          {/* Live indicator */}
          {group.status === 'IN_PROGRESS' && (
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
          {/* Status badge */}
          <Badge variant={group.status} />
        </div>

        {/* Content area */}
        <div className="rounded-b-lg overflow-hidden shadow-sm" style={{ border: '1px solid var(--border)', borderTop: 'none' }}>
          <div className="p-4 bg-white">
            {/* Standings */}
            <div className="mb-6">
              <h2 className="text-xs uppercase text-gray-400 font-medium mb-3">
                {t('liveView.standings', 'Standings')}
              </h2>
              <GroupStandings players={group.players} matches={group.matches} />
            </div>

            {/* Match grid */}
            <div>
              <h2 className="text-xs uppercase text-gray-400 font-medium mb-3">
                {t('liveView.matchResults', 'Match results')}
              </h2>
              <MatchGrid players={group.players} matches={group.matches} />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
