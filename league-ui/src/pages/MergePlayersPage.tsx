import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getDuplicatePlayers, mergePlayers, getPlayer } from '../api/players'
import type { DuplicateGroup, MergeResult } from '../api/players'
import type { User } from '../types'

function UserRow({ user, isTarget }: { user: User; isTarget: boolean }) {
  const { t } = useTranslation()
  return (
    <tr style={{ backgroundColor: isTarget ? 'rgba(34,197,94,0.08)' : undefined }}>
      <td className="px-3 py-2 text-sm text-gray-700">{user.userId}</td>
      <td className="px-3 py-2 text-sm font-medium text-gray-900">
        {user.lastName} {user.firstName}
        {isTarget && (
          <span className="ml-2 text-xs font-semibold text-green-700 bg-green-100 px-1.5 py-0.5 rounded">
            {t('mergePlayers.target')}
          </span>
        )}
      </td>
      <td className="px-3 py-2 text-sm text-gray-500">{user.email}</td>
      <td className="px-3 py-2 text-sm text-gray-700">{Math.round(user.currentRating)}</td>
    </tr>
  )
}

function DuplicateCard({
  group,
  onMerged,
}: {
  group: DuplicateGroup
  onMerged: () => void
}) {
  const { t } = useTranslation()
  // Earliest-created user is the target (index 0, since FindAllActive returns ORDER BY created ASC)
  const target = group.users[0]
  const [selected, setSelected] = useState<Set<number>>(
    new Set(group.users.slice(1).map((u) => u.userId))
  )
  const [merging, setMerging] = useState(false)
  const [result, setResult] = useState<MergeResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const toggle = (id: number) =>
    setSelected((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })

  const handleMerge = async () => {
    if (selected.size === 0) return
    setMerging(true)
    setError(null)
    try {
      const res = await mergePlayers(target.userId, Array.from(selected))
      setResult(res.data)
      onMerged()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(t('mergePlayers.mergeError', { error: msg }))
    } finally {
      setMerging(false)
    }
  }

  if (result) {
    return (
      <div className="border border-green-200 rounded-lg p-4 bg-green-50">
        <p className="text-sm text-green-800 font-medium">
          {result.recalcFromEvent
            ? t('mergePlayers.mergeSuccess', { event: result.recalcFromEvent })
            : t('mergePlayers.mergeDone')}
        </p>
        {result.conflictGroupIds.length > 0 && (
          <p className="text-xs text-green-700 mt-1">
            {t('mergePlayers.conflictGroups', { ids: result.conflictGroupIds.join(', ') })}
          </p>
        )}
      </div>
    )
  }

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      <div className="bg-gray-50 px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wide">
        {group.normalizedName}
      </div>
      <table className="w-full">
        <thead>
          <tr className="border-b border-gray-100">
            <th className="px-3 py-2 text-left text-xs text-gray-500 w-12">{t('mergePlayers.id')}</th>
            <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.name')}</th>
            <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.email')}</th>
            <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.rating')}</th>
            <th className="px-3 py-2 text-left text-xs text-gray-500 w-10"></th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-50">
          <tr style={{ backgroundColor: 'rgba(34,197,94,0.08)' }}>
            <td className="px-3 py-2 text-sm text-gray-700">{target.userId}</td>
            <td className="px-3 py-2 text-sm font-medium text-gray-900">
              {target.lastName} {target.firstName}
              <span className="ml-2 text-xs font-semibold text-green-700 bg-green-100 px-1.5 py-0.5 rounded">
                {t('mergePlayers.target')}
              </span>
            </td>
            <td className="px-3 py-2 text-sm text-gray-500">{target.email}</td>
            <td className="px-3 py-2 text-sm text-gray-700">{Math.round(target.currentRating)}</td>
            <td />
          </tr>
          {group.users.slice(1).map((u) => (
            <tr key={u.userId}>
              <td className="px-3 py-2 text-sm text-gray-700">{u.userId}</td>
              <td className="px-3 py-2 text-sm text-gray-900">
                {u.lastName} {u.firstName}
              </td>
              <td className="px-3 py-2 text-sm text-gray-500">{u.email}</td>
              <td className="px-3 py-2 text-sm text-gray-700">{Math.round(u.currentRating)}</td>
              <td className="px-3 py-2 text-center">
                <input
                  type="checkbox"
                  checked={selected.has(u.userId)}
                  onChange={() => toggle(u.userId)}
                  className="rounded border-gray-300"
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {error && <p className="px-4 py-2 text-sm text-red-600">{error}</p>}
      <div className="px-4 py-3 bg-gray-50 flex justify-end">
        <button
          onClick={handleMerge}
          disabled={merging || selected.size === 0}
          className="px-4 py-1.5 text-sm font-medium text-white rounded"
          style={{ backgroundColor: selected.size === 0 ? '#9ca3af' : 'var(--orange)', cursor: selected.size === 0 ? 'not-allowed' : 'pointer' }}
        >
          {merging ? t('mergePlayers.merging') : t('mergePlayers.mergeSelected')}
        </button>
      </div>
    </div>
  )
}

function DetectedTab() {
  const { t } = useTranslation()
  const [groups, setGroups] = useState<DuplicateGroup[]>([])
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    try {
      const res = await getDuplicatePlayers()
      setGroups(res.data)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  if (loading) return <p className="text-sm text-gray-500 py-6">{t('mergePlayers.loading')}</p>
  if (groups.length === 0) return <p className="text-sm text-gray-500 py-6">{t('mergePlayers.noDuplicates')}</p>

  return (
    <div className="flex flex-col gap-4">
      {groups.map((g) => (
        <DuplicateCard key={g.normalizedName} group={g} onMerged={load} />
      ))}
    </div>
  )
}

function ManualTab() {
  const { t } = useTranslation()
  const [targetId, setTargetId] = useState('')
  const [sourceIds, setSourceIds] = useState('')
  const [previewUsers, setPreviewUsers] = useState<User[]>([])
  const [loadingPreview, setLoadingPreview] = useState(false)
  const [merging, setMerging] = useState(false)
  const [result, setResult] = useState<MergeResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const parseIds = () => {
    const tid = parseInt(targetId, 10)
    const sids = sourceIds
      .split(',')
      .map((s) => parseInt(s.trim(), 10))
      .filter((n) => !isNaN(n))
    return { tid, sids }
  }

  const handlePreview = async () => {
    const { tid, sids } = parseIds()
    if (isNaN(tid) || sids.length === 0) return
    setLoadingPreview(true)
    setError(null)
    try {
      const ids = [tid, ...sids]
      const results = await Promise.allSettled(ids.map((id) => getPlayer(id)))
      setPreviewUsers(
        results
          .filter((r): r is PromiseFulfilledResult<Awaited<ReturnType<typeof getPlayer>>> => r.status === 'fulfilled')
          .map((r) => r.value.data)
      )
    } finally {
      setLoadingPreview(false)
    }
  }

  const handleMerge = async () => {
    const { tid, sids } = parseIds()
    if (isNaN(tid) || sids.length === 0) return
    setMerging(true)
    setError(null)
    try {
      const res = await mergePlayers(tid, sids)
      setResult(res.data)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(t('mergePlayers.mergeError', { error: msg }))
    } finally {
      setMerging(false)
    }
  }

  const { tid } = parseIds()

  return (
    <div className="flex flex-col gap-4 max-w-lg">
      <div className="flex flex-col gap-1">
        <label className="text-sm font-medium text-gray-700">{t('mergePlayers.targetIdLabel')}</label>
        <input
          type="number"
          value={targetId}
          onChange={(e) => setTargetId(e.target.value)}
          className="border border-gray-300 rounded px-3 py-2 text-sm"
          placeholder="e.g. 42"
        />
      </div>
      <div className="flex flex-col gap-1">
        <label className="text-sm font-medium text-gray-700">{t('mergePlayers.sourceIdsLabel')}</label>
        <input
          type="text"
          value={sourceIds}
          onChange={(e) => setSourceIds(e.target.value)}
          className="border border-gray-300 rounded px-3 py-2 text-sm"
          placeholder="e.g. 17, 83"
        />
      </div>
      <div className="flex gap-2">
        <button
          onClick={handlePreview}
          disabled={loadingPreview}
          className="px-3 py-1.5 text-sm border border-gray-300 rounded text-gray-700 hover:bg-gray-50"
        >
          {loadingPreview ? '…' : t('mergePlayers.loadPreview')}
        </button>
        <button
          onClick={handleMerge}
          disabled={merging}
          className="px-4 py-1.5 text-sm font-medium text-white rounded"
          style={{ backgroundColor: 'var(--orange)' }}
        >
          {merging ? t('mergePlayers.merging') : t('mergePlayers.mergeSelected')}
        </button>
      </div>

      {previewUsers.length > 0 && (
        <div className="border border-gray-200 rounded-lg overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-100">
                <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.id')}</th>
                <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.name')}</th>
                <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.email')}</th>
                <th className="px-3 py-2 text-left text-xs text-gray-500">{t('mergePlayers.rating')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {previewUsers.map((u) => (
                <UserRow key={u.userId} user={u} isTarget={u.userId === tid} />
              ))}
            </tbody>
          </table>
        </div>
      )}

      {error && <p className="text-sm text-red-600">{error}</p>}
      {result && (
        <div className="border border-green-200 rounded-lg p-4 bg-green-50">
          <p className="text-sm text-green-800 font-medium">
            {result.recalcFromEvent
              ? t('mergePlayers.mergeSuccess', { event: result.recalcFromEvent })
              : t('mergePlayers.mergeDone')}
          </p>
          {result.conflictGroupIds.length > 0 && (
            <p className="text-xs text-green-700 mt-1">
              {t('mergePlayers.conflictGroups', { ids: result.conflictGroupIds.join(', ') })}
            </p>
          )}
        </div>
      )}
    </div>
  )
}

export function MergePlayersPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<'detected' | 'manual'>('detected')

  const tabStyle = (active: boolean): React.CSSProperties => ({
    padding: '8px 16px',
    fontSize: 14,
    fontWeight: active ? 600 : 400,
    color: active ? 'var(--orange)' : '#6b7280',
    borderBottom: active ? '2px solid var(--orange)' : '2px solid transparent',
    background: 'none',
    border: 'none',
    borderBottom: active ? '2px solid var(--orange)' : '2px solid transparent',
    cursor: 'pointer',
  })

  return (
    <div className="max-w-4xl mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">{t('mergePlayers.title')}</h1>
      <div className="flex border-b border-gray-200 mb-6">
        <button style={tabStyle(tab === 'detected')} onClick={() => setTab('detected')}>
          {t('mergePlayers.tabDetected')}
        </button>
        <button style={tabStyle(tab === 'manual')} onClick={() => setTab('manual')}>
          {t('mergePlayers.tabManual')}
        </button>
      </div>
      {tab === 'detected' ? <DetectedTab /> : <ManualTab />}
    </div>
  )
}
