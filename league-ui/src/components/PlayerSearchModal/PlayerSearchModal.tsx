import { useState, useEffect } from 'react'
import { Modal } from '../Modal/Modal'
import { usePlayers } from '../../hooks/usePlayers'
import type { User } from '../../types'

interface PlayerSearchModalProps {
  open: boolean
  onClose: () => void
  onAdd: (userId: number, playerName: string) => Promise<boolean>
  assignedUserIds: Set<number>
  title: string
  loading?: boolean
}

export function PlayerSearchModal({
  open,
  onClose,
  onAdd,
  assignedUserIds,
  title,
  loading = false,
}: PlayerSearchModalProps) {
  const [query, setQuery] = useState('')
  const [successMessage, setSuccessMessage] = useState('')
  const [loadingPlayerId, setLoadingPlayerId] = useState<number | null>(null)

  const { players } = usePlayers({ q: query, limit: 20, sort: 'rating' })

  useEffect(() => {
    if (!open) {
      setQuery('')
      setSuccessMessage('')
    }
  }, [open])

  const filteredPlayers = players.filter((p) => !assignedUserIds.has(p.userId))

  const handleAddPlayer = async (player: User) => {
    setLoadingPlayerId(player.userId)
    try {
      const fullName = `${player.firstName} ${player.lastName}`
      const success = await onAdd(player.userId, fullName)
      if (success) {
        setSuccessMessage(`${fullName} has been added`)
      }
    } finally {
      setLoadingPlayerId(null)
    }
  }

  const handleQueryChange = (newQuery: string) => {
    setQuery(newQuery)
    setSuccessMessage('')
  }

  return (
    <Modal open={open} onClose={onClose} title={title}>
      <div className="flex flex-col gap-4">
        <input
          autoFocus
          type="text"
          placeholder="Type to search players"
          value={query}
          onChange={(e) => handleQueryChange(e.target.value)}
          className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#FF7A00] focus:border-[#FF7A00]"
        />

        {successMessage && (
          <p className="text-sm text-green-600">{successMessage}</p>
        )}

        <div className="flex flex-col gap-2 max-h-72 overflow-y-auto">
          {query === '' && filteredPlayers.length === 0 && (
            <p className="text-sm text-gray-400">Type to search players</p>
          )}

          {query !== '' && filteredPlayers.length === 0 && (
            <p className="text-sm text-gray-400">No players found</p>
          )}

          {filteredPlayers.map((player) => (
            <div
              key={player.userId}
              className="flex items-center justify-between rounded border border-gray-200 bg-gray-50 px-3 py-2"
            >
              <div className="flex-1">
                <span className="text-sm text-gray-800">
                  {player.firstName} {player.lastName}
                </span>
                <span className="ml-2 text-xs text-gray-500">
                  {Math.round(player.currentRating)}
                </span>
              </div>
              <button
                onClick={() => handleAddPlayer(player)}
                disabled={loadingPlayerId === player.userId || loading}
                className="ml-2 rounded bg-blue-600 px-3 py-1 text-xs font-semibold text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loadingPlayerId === player.userId ? 'Adding...' : 'Add'}
              </button>
            </div>
          ))}
        </div>
      </div>
    </Modal>
  )
}
