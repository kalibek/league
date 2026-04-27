import { useState } from 'react'
import type { GroupPlayer } from '../../types'
import { Button } from '../Button/Button'
import { useTranslation } from 'react-i18next'

interface PlacementOverrideProps {
  players: GroupPlayer[]
  onConfirm: (orderedPlayerIds: number[]) => void
  onClose: () => void
  loading?: boolean
}

export function PlacementOverride({ players, onConfirm, onClose, loading = false }: PlacementOverrideProps) {
  const { t } = useTranslation()
  const [ordered, setOrdered] = useState<GroupPlayer[]>([...players])
  const [draggingIndex, setDraggingIndex] = useState<number | null>(null)

  const playerName = (p: GroupPlayer) =>
    p.user ? `${p.user.firstName} ${p.user.lastName}` : `Player #${p.userId}`

  const handleDragStart = (e: React.DragEvent, index: number) => {
    setDraggingIndex(index)
    e.dataTransfer.effectAllowed = 'move'
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggingIndex === null || draggingIndex === index) return

    const newOrder = [...ordered]
    const [removed] = newOrder.splice(draggingIndex, 1)
    newOrder.splice(index, 0, removed)
    setOrdered(newOrder)
    setDraggingIndex(index)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDraggingIndex(null)
  }

  const moveUp = (index: number) => {
    if (index === 0) return
    const newOrder = [...ordered]
    ;[newOrder[index - 1], newOrder[index]] = [newOrder[index], newOrder[index - 1]]
    setOrdered(newOrder)
  }

  const moveDown = (index: number) => {
    if (index === ordered.length - 1) return
    const newOrder = [...ordered]
    ;[newOrder[index], newOrder[index + 1]] = [newOrder[index + 1], newOrder[index]]
    setOrdered(newOrder)
  }

  const handleConfirm = () => {
    onConfirm(ordered.map((p) => p.groupPlayerId))
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm text-gray-600">
        {t('placementOverride.dragInstruction')}
      </p>
      <ul className="divide-y divide-gray-100 rounded-md border border-gray-200">
        {ordered.map((p, index) => (
          <li
            key={p.groupPlayerId}
            draggable
            onDragStart={(e) => handleDragStart(e, index)}
            onDragOver={(e) => handleDragOver(e, index)}
            onDrop={handleDrop}
            className={`flex items-center justify-between px-4 py-3 cursor-grab active:cursor-grabbing select-none transition-colors ${
              draggingIndex === index ? 'bg-blue-50' : 'hover:bg-gray-50'
            }`}
          >
            <div className="flex items-center gap-3">
              <span className="text-gray-400 text-xs w-4">{index + 1}</span>
              <span className="font-medium text-gray-800">{playerName(p)}</span>
              <span className="text-xs text-gray-400">
                {p.points} pts / {p.tiebreakPoints} TB
              </span>
            </div>
            <div className="flex gap-1">
              <button
                onClick={() => moveUp(index)}
                disabled={index === 0}
                className="p-1 text-gray-400 hover:text-gray-700 disabled:opacity-20"
                aria-label={t('placementOverride.moveUp')}
              >
                ↑
              </button>
              <button
                onClick={() => moveDown(index)}
                disabled={index === ordered.length - 1}
                className="p-1 text-gray-400 hover:text-gray-700 disabled:opacity-20"
                aria-label={t('placementOverride.moveDown')}
              >
                ↓
              </button>
            </div>
          </li>
        ))}
      </ul>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={loading}>
          {t('placementOverride.cancel')}
        </Button>
        <Button variant="primary" onClick={handleConfirm} loading={loading}>
          {t('placementOverride.confirmPlacement')}
        </Button>
      </div>
    </div>
  )
}
