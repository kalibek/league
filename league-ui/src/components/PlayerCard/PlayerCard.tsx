import type { User } from '../../types'

interface PlayerCardProps {
  player: User
}

export function PlayerCard({ player }: PlayerCardProps) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-gray-200 bg-white px-4 py-3 shadow-sm">
      <div>
        <p className="font-medium text-gray-900">
          {player.firstName} {player.lastName}
        </p>
        <p className="text-xs text-gray-500">{player.email}</p>
      </div>
      <div className="text-right">
        <p className="text-lg font-semibold text-blue-700">{Math.round(player.currentRating)}</p>
        <p className="text-xs text-gray-400">RD ±{Math.round(player.deviation)}</p>
      </div>
    </div>
  )
}
