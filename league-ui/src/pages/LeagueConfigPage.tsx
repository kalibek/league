import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useLeague, useUpdateConfig, useLeagueRoles, useAssignRole } from '../hooks/useLeagues'
import { usePlayers } from '../hooks/usePlayers'
import { LeagueConfigForm } from '../components/LeagueConfigForm/LeagueConfigForm'
import { Button } from '../components/Button/Button'
import { useAuth } from '../hooks/useAuth'
import type { LeagueConfig } from '../types'

export function LeagueConfigPage() {
  const { id } = useParams<{ id: string }>()
  const leagueId = Number(id)
  const { league, loading } = useLeague(leagueId)
  const { update, loading: saving, error } = useUpdateConfig()
  const navigate = useNavigate()
  const { isAdmin, isMaintainer } = useAuth()

  const canManageRoles = isAdmin || isMaintainer(leagueId)
  const { roles, refresh: refreshRoles } = useLeagueRoles(leagueId)
  const { assign, remove, loading: roleSaving, error: roleError } = useAssignRole()
  const { players } = usePlayers({ limit: 200 })

  const [selectedUserId, setSelectedUserId] = useState<number | ''>('')

  const handleSubmit = async (config: LeagueConfig) => {
    const result = await update(leagueId, config)
    if (result) navigate(`/leagues/${leagueId}`)
  }

  const handleAssignMaintainer = async () => {
    if (!selectedUserId) return
    const ok = await assign(leagueId, { userId: Number(selectedUserId), roleName: 'maintainer' })
    if (ok) {
      setSelectedUserId('')
      refreshRoles()
    }
  }

  const handleRemoveMaintainer = async (userId: number) => {
    const ok = await remove(leagueId, userId, 'maintainer')
    if (ok) refreshRoles()
  }

  if (loading) return <div className="p-8 text-gray-400">Loading...</div>
  if (!league) return <div className="p-8 text-red-600">League not found</div>

  const maintainers = roles.filter((r) => r.roleName === 'maintainer')

  return (
    <div className="max-w-xl mx-auto py-8 px-4">
      <Link to={`/leagues/${leagueId}`} className="text-sm text-blue-600 hover:underline mb-4 block">
        &larr; Back to League
      </Link>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Configure League</h1>
      <h2 className="text-base text-gray-600 mb-4">{league.title}</h2>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <LeagueConfigForm
        initial={league.configuration}
        onSubmit={handleSubmit}
        loading={saving}
        showDraftWarning
      />

      {canManageRoles && (
        <div className="mt-10 border-t border-gray-200 pt-8">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Maintainers</h3>

          {roleError && <p className="text-red-600 text-sm mb-3">{roleError}</p>}

          <ul className="divide-y divide-gray-100 mb-4">
            {maintainers.length === 0 && (
              <li className="py-2 text-sm text-gray-400">No maintainers assigned.</li>
            )}
            {maintainers.map((m) => (
              <li key={m.userId} className="py-2 flex items-center justify-between">
                <span className="text-sm text-gray-800">
                  {m.firstName} {m.lastName}
                  <span className="text-gray-400 ml-2 text-xs">{m.email}</span>
                </span>
                <button
                  onClick={() => handleRemoveMaintainer(m.userId)}
                  disabled={roleSaving}
                  className="text-xs text-red-500 hover:text-red-700 disabled:opacity-50"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>

          <div className="flex gap-2 items-center">
            <select
              value={selectedUserId}
              onChange={(e) => setSelectedUserId(e.target.value === '' ? '' : Number(e.target.value))}
              className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Select a player...</option>
              {players.map((p) => (
                <option key={p.userId} value={p.userId}>
                  {p.firstName} {p.lastName} ({p.email})
                </option>
              ))}
            </select>
            <Button
              variant="primary"
              onClick={handleAssignMaintainer}
              disabled={!selectedUserId || roleSaving}
              loading={roleSaving}
            >
              Add Maintainer
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
