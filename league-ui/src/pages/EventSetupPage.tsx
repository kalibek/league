import { useState, useMemo, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useEvent, useUpdateEventDetails } from '../hooks/useEvents'
import { useGroups, useCreateGroup, useSeedPlayer, useRemoveGroupPlayer, useAddPlayerToActiveGroup } from '../hooks/useGroups'
import { Badge } from '../components/Badge/Badge'
import { Button } from '../components/Button/Button'
import { PlayerSearchModal } from '../components/PlayerSearchModal/PlayerSearchModal'
import type { Group, GroupPlayer } from '../types'

const DIVISIONS = ['Superleague', 'A', 'B', 'C', 'D', 'E']

interface GroupWithPlayers extends Group {
  players: GroupPlayer[]
}

export function EventSetupPage() {
  const { t } = useTranslation()
  const { id, eid } = useParams<{ id: string; eid: string }>()
  const leagueId = Number(id)
  const eventId = Number(eid)

  const { event, setEvent, loading: eventLoading } = useEvent(leagueId, eventId)
  const { groups, loading: groupsLoading, refresh: refreshGroups } = useGroups(eventId)
  const { create: createGroup, loading: creating, error: createError } = useCreateGroup()
  const { seed, loading: seeding, error: seedError } = useSeedPlayer()
  const { remove, loading: removing } = useRemoveGroupPlayer()
  const { addActive, loading: addingActive, error: addActiveError } = useAddPlayerToActiveGroup()
  const { update: updateDetails, loading: updatingDetails, error: detailsError } = useUpdateEventDetails()

  // groupPlayers is fetched per group via the group detail endpoint.
  // We maintain a map of groupId → GroupPlayer[] in local state.
  const [groupPlayers, setGroupPlayers] = useState<Record<number, GroupPlayer[]>>({})

  // Inline editing state for event details
  const [isEditingDetails, setIsEditingDetails] = useState(false)
  const [detailsForm, setDetailsForm] = useState({ title: '', startDate: '', endDate: '' })

  // Helper to format ISO datetime → YYYY-MM-DD
  const toDateInput = (iso: string) => iso.slice(0, 10)

  // Sync form when event changes
  useEffect(() => {
    if (event) {
      setDetailsForm({
        title: event.title,
        startDate: toDateInput(event.startDate),
        endDate: toDateInput(event.endDate),
      })
    }
  }, [event])

  const [showAddGroup, setShowAddGroup] = useState(false)
  const [groupForm, setGroupForm] = useState({ division: 'A', groupNo: 1, scheduled: '' })

  // Which group has search modal open (null if none)
  const [searchModalGroup, setSearchModalGroup] = useState<number | null>(null)

  // All userIds already assigned to any group in this event
  const assignedUserIds = useMemo(
    () => new Set(Object.values(groupPlayers).flat().map((gp) => gp.userId)),
    [groupPlayers]
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

  const handleRemovePlayer = async (groupId: number, groupPlayerId: number) => {
    const ok = await remove(eventId, groupId, groupPlayerId)
    if (ok) await loadGroupPlayers(groupId)
  }

  const handleAddPlayerFromModal = async (groupId: number, userId: number): Promise<boolean> => {
    const isDraft = event?.status === 'DRAFT'
    const isInProgress = event?.status === 'IN_PROGRESS'

    let ok = false
    if (isDraft) {
      ok = await seed(eventId, groupId, userId)
    } else if (isInProgress) {
      ok = await addActive(eventId, groupId, userId)
    }

    if (ok) {
      await loadGroupPlayers(groupId)
    }
    return ok
  }

  const groupsWithPlayers: GroupWithPlayers[] = groups.map((g) => ({
    ...g,
    players: groupPlayers[g.groupId] ?? [],
  }))

  const handleSaveDetails = async () => {
    const updated = await updateDetails(leagueId, eventId, {
      title: detailsForm.title,
      startDate: detailsForm.startDate,
      endDate: detailsForm.endDate,
    })
    if (updated) {
      setEvent(updated)
      setIsEditingDetails(false)
    }
  }

  const handleCancelEdit = () => {
    if (event) {
      setDetailsForm({
        title: event.title,
        startDate: toDateInput(event.startDate),
        endDate: toDateInput(event.endDate),
      })
    }
    setIsEditingDetails(false)
  }

  if (eventLoading) return <div className="p-8 text-gray-400">{t('eventSetup.loading')}</div>
  if (!event) return <div className="p-8 text-red-600">{t('eventSetup.notFound')}</div>

  const isDraft = event.status === 'DRAFT'
  const isInProgress = event.status === 'IN_PROGRESS'

  return (
    <div className="max-w-5xl mx-auto py-8 px-4">
      <Link
        to={`/leagues/${leagueId}`}
        className="text-sm text-blue-600 hover:underline mb-4 block"
      >
        {t('eventSetup.backToLeague')}
      </Link>

      {/* Event header */}
      <div className="flex items-start justify-between mb-6">
        <div className="flex-1">
          {!isEditingDetails ? (
            <>
              <div className="flex items-center gap-3 mb-1">
                <h1 className="text-2xl font-bold text-gray-900">{event.title}</h1>
                <Badge variant={event.status} />
                {isInProgress && (
                  <button
                    onClick={() => setIsEditingDetails(true)}
                    className="text-gray-400 hover:text-gray-600 p-1"
                    title="Edit event details"
                  >
                    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536M9 13l6.536-6.536a2 2 0 012.828 0l.172.172a2 2 0 010 2.828L12 16H9v-3z" />
                    </svg>
                  </button>
                )}
              </div>
              <p className="text-sm text-gray-400">
                {event.startDate} — {event.endDate}
              </p>
            </>
          ) : (
            <div className="space-y-3 mb-4">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Title</label>
                <input
                  type="text"
                  className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                  value={detailsForm.title}
                  onChange={(e) => setDetailsForm((f) => ({ ...f, title: e.target.value }))}
                />
              </div>
              <div className="flex gap-3">
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-600 mb-1">Start Date</label>
                  <input
                    type="date"
                    className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                    value={detailsForm.startDate}
                    onChange={(e) => setDetailsForm((f) => ({ ...f, startDate: e.target.value }))}
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-600 mb-1">End Date</label>
                  <input
                    type="date"
                    className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                    value={detailsForm.endDate}
                    onChange={(e) => setDetailsForm((f) => ({ ...f, endDate: e.target.value }))}
                  />
                </div>
              </div>
              {detailsError && <p className="text-xs text-red-600">{detailsError}</p>}
              <div className="flex gap-2 pt-2">
                <Button type="button" variant="secondary" onClick={handleCancelEdit}>
                  {t('eventSetup.cancel')}
                </Button>
                <Button type="button" onClick={handleSaveDetails} loading={updatingDetails}>
                  {t('eventSetup.save', 'Save')}
                </Button>
              </div>
            </div>
          )}
        </div>
        {isDraft && !isEditingDetails && (
          <Link
            to={`/leagues/${leagueId}`}
            className="text-sm text-gray-500 hover:text-gray-700"
          >
            {t('eventSetup.startFromLeague')}
          </Link>
        )}
      </div>

      {!isDraft && !isInProgress && (
        <div className="mb-4 rounded-md bg-yellow-50 border border-yellow-200 px-4 py-3 text-sm text-yellow-800">
          {t('eventSetup.notDraftWarning')}
        </div>
      )}
      {isInProgress && (
        <div className="mb-4 rounded-md bg-blue-50 border border-blue-200 px-4 py-3 text-sm text-blue-800">
          {t('eventSetup.inProgressNote', 'Event is in progress. You can add players to groups below.')}
        </div>
      )}

      {/* Groups */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-800">
          {t('eventSetup.groups')} {!groupsLoading && `(${groups.length})`}
        </h2>
        {isDraft && (
          <Button variant="primary" onClick={() => setShowAddGroup(true)}>
            {t('eventSetup.addGroup')}
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
              <label className="block text-xs font-medium text-gray-600 mb-1">{t('eventSetup.division')}</label>
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
              <label className="block text-xs font-medium text-gray-600 mb-1">{t('eventSetup.groupNumber')}</label>
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
                {t('eventSetup.scheduledDateTime')}
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
              {t('eventSetup.cancel')}
            </Button>
            <Button type="submit" loading={creating}>
              {t('eventSetup.createGroup')}
            </Button>
          </div>
        </form>
      )}

      {groupsLoading && <p className="text-gray-400 text-sm">{t('eventSetup.loadingGroups')}</p>}

      {/* Group cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {groupsWithPlayers.map((grp) => (
          <div key={grp.groupId} className="rounded-lg border border-gray-200 bg-white p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-semibold text-gray-800">
                {grp.division === 'Superleague' ? 'Superleague' : `${grp.division}${grp.groupNo}`}
              </h3>
              <span className="text-xs text-gray-400">
                {t('eventSetup.players', { count: grp.players.length })}
              </span>
            </div>

            {/* Player list */}
            <ul className="space-y-1 mb-3 min-h-[2rem]">
              {grp.players.map((gp) => {
                const playerName = gp.user
                  ? `${gp.user.firstName} ${gp.user.lastName}`
                  : t('eventSetup.player', { id: gp.userId })
                const rating = gp.user ? Math.round(gp.user.currentRating) : null

                return (
                  <li
                    key={gp.groupPlayerId}
                    className="flex items-center justify-between text-sm"
                  >
                    <span className="text-gray-700">
                      {playerName}
                      {rating && (
                        <span className="ml-1 text-xs text-gray-400">
                          {rating}
                        </span>
                      )}
                    </span>
                    {isDraft && (
                      <button
                        onClick={() => handleRemovePlayer(grp.groupId, gp.groupPlayerId)}
                        disabled={removing}
                        className="text-gray-300 hover:text-red-500 text-xs px-1"
                        title={t('groupStandings.noShow', { name: playerName })}
                      >
                        ×
                      </button>
                    )}
                  </li>
                )
              })}
              {grp.players.length === 0 && (
                <li className="text-xs text-gray-400 italic">{t('eventSetup.noPlayersYet')}</li>
              )}
            </ul>

            {/* Add player button */}
            {(isDraft || isInProgress) && (
              <Button
                variant="secondary"
                onClick={() => setSearchModalGroup(grp.groupId)}
                className="w-full"
              >
                {t('eventSetup.add')} Player
              </Button>
            )}

            {seedError && <p className="text-xs text-red-600 mt-1">{seedError}</p>}
            {addActiveError && <p className="text-xs text-red-600 mt-1">{addActiveError}</p>}
          </div>
        ))}
      </div>

      {!groupsLoading && groups.length === 0 && (
        <p className="text-gray-400 text-sm mt-2">
          {isDraft ? t('eventSetup.noGroupsAddAbove') : t('eventSetup.noGroupsYet')}
        </p>
      )}

      <div className="mt-4 text-xs text-gray-400">
        {t('eventSetup.playersAssigned', { assigned: assignedUserIds.size })}
      </div>

      <button
        onClick={refreshAll}
        className="mt-2 text-xs text-blue-500 hover:underline"
      >
        {t('eventSetup.refresh')}
      </button>

      {searchModalGroup !== null && (() => {
        const selectedGroup = groupsWithPlayers.find((g) => g.groupId === searchModalGroup)
        const groupTitle = selectedGroup
          ? selectedGroup.division === 'Superleague'
            ? 'Superleague'
            : `${selectedGroup.division}${selectedGroup.groupNo}`
          : 'Group'

        return (
          <PlayerSearchModal
            open={true}
            onClose={() => setSearchModalGroup(null)}
            onAdd={(userId) => handleAddPlayerFromModal(searchModalGroup, userId)}
            assignedUserIds={assignedUserIds}
            title={`Add player to ${groupTitle}`}
            loading={seeding || addingActive}
          />
        )
      })()}
    </div>
  )
}
