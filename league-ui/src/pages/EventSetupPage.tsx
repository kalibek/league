import { useState, useMemo } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useEvent } from '../hooks/useEvents'
import { useGroups, useCreateGroup, useSeedPlayer, useRemoveGroupPlayer } from '../hooks/useGroups'
import { usePlayers } from '../hooks/usePlayers'
import { Badge } from '../components/Badge/Badge'
import { Button } from '../components/Button/Button'
import type { Group, GroupPlayer } from '../types'

const DIVISIONS = ['Superleague', 'A', 'B', 'C', 'D', 'E']

interface GroupWithPlayers extends Group {
  players: GroupPlayer[]
}

export function EventSetupPage() {
  const { id, eid } = useParams<{ id: string; eid: string }>()
  const leagueId = Number(id)
  const eventId = Number(eid)

  const { event, loading: eventLoading } = useEvent(leagueId, eventId)
  const { groups, loading: groupsLoading, refresh: refreshGroups } = useGroups(eventId)
  const { players, loading: playersLoading } = usePlayers({ limit: 500 })
  const { create: createGroup, loading: creating, error: createError } = useCreateGroup()
  const { seed, loading: seeding, error: seedError } = useSeedPlayer()
  const { remove, loading: removing } = useRemoveGroupPlayer()

  // groupPlayers is fetched per group via the group detail endpoint.
  // We maintain a map of groupId → GroupPlayer[] in local state.
  const [groupPlayers, setGroupPlayers] = useState<Record<number, GroupPlayer[]>>({})

  const [showAddGroup, setShowAddGroup] = useState(false)
  const [groupForm, setGroupForm] = useState({ division: 'A', groupNo: 1, scheduled: '' })

  // Per-group: which player is selected in the dropdown
  const [selectedPlayer, setSelectedPlayer] = useState<Record<number, number>>({})

  // All userIds already assigned to any group in this event
  const assignedUserIds = useMemo(
    () => new Set(Object.values(groupPlayers).flat().map((gp) => gp.userId)),
    [groupPlayers]
  )

  const availablePlayers = useMemo(
    () => players.filter((p) => !assignedUserIds.has(p.userId)),
    [players, assignedUserIds]
  )

  // Load players for a group via the group detail endpoint
  const loadGroupPlayers = async (groupId: number) => {
    const { getGroup } = await import('../api/groups')
    try {
      const res = await getGroup(eventId, groupId)
      setGroupPlayers((prev) => ({ ...prev, [groupId]: res.data.players ?? [] }))
    } catch {
      setGroupPlayers((prev) => ({ ...prev, [groupId]: [] }))
    }
  }

  // Load all groups' players on mount / refresh
  const refreshAll = async () => {
    refreshGroups()
    for (const g of groups) {
      await loadGroupPlayers(g.groupId)
    }
  }

  // When groups change, load missing player lists
  useMemo(() => {
    groups.forEach((g) => {
      if (groupPlayers[g.groupId] === undefined) {
        loadGroupPlayers(g.groupId)
      }
    })
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groups])

  const handleCreateGroup = async (e: React.FormEvent) => {
    e.preventDefault()
    const scheduled = groupForm.scheduled
      ? new Date(groupForm.scheduled).toISOString()
      : new Date().toISOString()
    const grp = await createGroup(eventId, {
      division: groupForm.division,
      groupNo: groupForm.groupNo,
      scheduled,
    })
    if (grp) {
      setShowAddGroup(false)
      setGroupPlayers((prev) => ({ ...prev, [grp.groupId]: [] }))
      refreshGroups()
    }
  }

  const handleSeedPlayer = async (groupId: number) => {
    const userId = selectedPlayer[groupId]
    if (!userId) return
    const ok = await seed(eventId, groupId, userId)
    if (ok) {
      setSelectedPlayer((prev) => ({ ...prev, [groupId]: 0 }))
      await loadGroupPlayers(groupId)
      setGroupPlayers((prev) => ({ ...prev })) // trigger re-render for availablePlayers
    }
  }

  const handleRemovePlayer = async (groupId: number, groupPlayerId: number) => {
    const ok = await remove(eventId, groupId, groupPlayerId)
    if (ok) await loadGroupPlayers(groupId)
  }

  const groupsWithPlayers: GroupWithPlayers[] = groups.map((g) => ({
    ...g,
    players: groupPlayers[g.groupId] ?? [],
  }))

  if (eventLoading) return <div className="p-8 text-gray-400">Loading...</div>
  if (!event) return <div className="p-8 text-red-600">Event not found</div>

  const isDraft = event.status === 'DRAFT'

  return (
    <div className="max-w-5xl mx-auto py-8 px-4">
      <Link
        to={`/leagues/${leagueId}`}
        className="text-sm text-blue-600 hover:underline mb-4 block"
      >
        &larr; Back to League
      </Link>

      {/* Event header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-2xl font-bold text-gray-900">{event.title}</h1>
            <Badge variant={event.status} />
          </div>
          <p className="text-sm text-gray-400">
            {event.startDate} — {event.endDate}
          </p>
        </div>
        {isDraft && (
          <Link
            to={`/leagues/${leagueId}`}
            className="text-sm text-gray-500 hover:text-gray-700"
          >
            Start event from league page
          </Link>
        )}
      </div>

      {!isDraft && (
        <div className="mb-4 rounded-md bg-yellow-50 border border-yellow-200 px-4 py-3 text-sm text-yellow-800">
          Event is not in DRAFT status — group and player setup is locked.
        </div>
      )}

      {/* Groups */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-800">
          Groups {!groupsLoading && `(${groups.length})`}
        </h2>
        {isDraft && (
          <Button variant="primary" onClick={() => setShowAddGroup(true)}>
            + Add Group
          </Button>
        )}
      </div>

      {/* Add group form */}
      {showAddGroup && (
        <form
          onSubmit={handleCreateGroup}
          className="mb-4 p-4 rounded-lg border border-gray-200 bg-gray-50 flex flex-col gap-3"
        >
          <div className="flex gap-3 flex-wrap">
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Division</label>
              <select
                className="rounded border border-gray-300 px-3 py-2 text-sm"
                value={groupForm.division}
                onChange={(e) => setGroupForm((f) => ({ ...f, division: e.target.value }))}
              >
                {DIVISIONS.map((d) => (
                  <option key={d} value={d}>{d}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Group #</label>
              <input
                type="number"
                min={1}
                className="rounded border border-gray-300 px-3 py-2 text-sm w-24"
                value={groupForm.groupNo}
                onChange={(e) => setGroupForm((f) => ({ ...f, groupNo: Number(e.target.value) }))}
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">
                Scheduled date/time
              </label>
              <input
                type="datetime-local"
                className="rounded border border-gray-300 px-3 py-2 text-sm"
                value={groupForm.scheduled}
                onChange={(e) => setGroupForm((f) => ({ ...f, scheduled: e.target.value }))}
              />
            </div>
          </div>
          {createError && <p className="text-xs text-red-600">{createError}</p>}
          <div className="flex gap-2">
            <Button type="button" variant="secondary" onClick={() => setShowAddGroup(false)}>
              Cancel
            </Button>
            <Button type="submit" loading={creating}>
              Create Group
            </Button>
          </div>
        </form>
      )}

      {groupsLoading && <p className="text-gray-400 text-sm">Loading groups...</p>}

      {/* Group cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {groupsWithPlayers.map((grp) => {
          const playerMap = Object.fromEntries(players.map((p) => [p.userId, p]))
          const groupAvailable = availablePlayers

          return (
            <div key={grp.groupId} className="rounded-lg border border-gray-200 bg-white p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-gray-800">
                  {grp.division === 'Superleague' ? 'Superleague' : `${grp.division}${grp.groupNo}`}
                </h3>
                <span className="text-xs text-gray-400">
                  {grp.players.length} player{grp.players.length !== 1 ? 's' : ''}
                </span>
              </div>

              {/* Player list */}
              <ul className="space-y-1 mb-3 min-h-[2rem]">
                {grp.players.map((gp) => {
                  const user = playerMap[gp.userId]
                  return (
                    <li
                      key={gp.groupPlayerId}
                      className="flex items-center justify-between text-sm"
                    >
                      <span className="text-gray-700">
                        {user ? `${user.firstName} ${user.lastName}` : `Player #${gp.userId}`}
                        {user && (
                          <span className="ml-1 text-xs text-gray-400">
                            {Math.round(user.currentRating)}
                          </span>
                        )}
                      </span>
                      {isDraft && (
                        <button
                          onClick={() => handleRemovePlayer(grp.groupId, gp.groupPlayerId)}
                          disabled={removing}
                          className="text-gray-300 hover:text-red-500 text-xs px-1"
                          title="Remove player"
                        >
                          ×
                        </button>
                      )}
                    </li>
                  )
                })}
                {grp.players.length === 0 && (
                  <li className="text-xs text-gray-400 italic">No players yet</li>
                )}
              </ul>

              {/* Add player */}
              {isDraft && (
                <div className="flex gap-2 mt-2">
                  <select
                    className="flex-1 rounded border border-gray-300 px-2 py-1 text-xs"
                    value={selectedPlayer[grp.groupId] ?? ''}
                    onChange={(e) =>
                      setSelectedPlayer((prev) => ({
                        ...prev,
                        [grp.groupId]: Number(e.target.value),
                      }))
                    }
                    disabled={playersLoading || groupAvailable.length === 0}
                  >
                    <option value="">
                      {groupAvailable.length === 0 ? 'No players available' : '— Add player —'}
                    </option>
                    {groupAvailable.map((p) => (
                      <option key={p.userId} value={p.userId}>
                        {p.firstName} {p.lastName} ({Math.round(p.currentRating)})
                      </option>
                    ))}
                  </select>
                  <Button
                    variant="secondary"
                    onClick={() => handleSeedPlayer(grp.groupId)}
                    disabled={!selectedPlayer[grp.groupId] || seeding}
                    loading={seeding}
                  >
                    Add
                  </Button>
                </div>
              )}
              {seedError && <p className="text-xs text-red-600 mt-1">{seedError}</p>}
            </div>
          )
        })}
      </div>

      {!groupsLoading && groups.length === 0 && (
        <p className="text-gray-400 text-sm mt-2">
          No groups yet.{isDraft ? ' Add groups above.' : ''}
        </p>
      )}

      <div className="mt-4 text-xs text-gray-400">
        {assignedUserIds.size} of {players.length} players assigned
      </div>

      <button
        onClick={refreshAll}
        className="mt-2 text-xs text-blue-500 hover:underline"
      >
        Refresh
      </button>
    </div>
  )
}
