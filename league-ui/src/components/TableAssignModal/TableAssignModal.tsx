import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Match } from '../../types'
import { Select } from '../Select/Select'
import { Button } from '../Button/Button'

interface TableAssignModalProps {
  match: Match
  numberOfTables: number
  tablesInUse: number[]
  onSubmit: (tableNumber: number) => void
  onClose: () => void
  loading?: boolean
}

export function TableAssignModal({
  numberOfTables,
  tablesInUse,
  onSubmit,
  onClose,
  loading = false,
}: TableAssignModalProps) {
  const { t } = useTranslation()

  const availableTables = Array.from({ length: numberOfTables }, (_, i) => i + 1).filter(
    (n) => !tablesInUse.includes(n)
  )

  const [selected, setSelected] = useState<number>(availableTables[0] ?? 0)

  if (availableTables.length === 0) {
    return (
      <div className="flex flex-col gap-4">
        <p className="text-sm text-gray-600">{t('liveView.allTablesOccupied')}</p>
        <Button variant="secondary" onClick={onClose}>
          {t('scoreEntry.cancel')}
        </Button>
      </div>
    )
  }

  const options = availableTables.map((n) => ({ value: n, label: `${t('liveView.table')} ${n}` }))

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (selected) onSubmit(selected)
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <Select
        label={t('liveView.table')}
        options={options}
        value={selected}
        onChange={(e) => setSelected(Number(e.target.value))}
      />
      <div className="flex gap-2 justify-end">
        <Button type="button" variant="secondary" onClick={onClose} disabled={loading}>
          {t('scoreEntry.cancel')}
        </Button>
        <Button type="submit" variant="primary" loading={loading}>
          {t('liveView.assign')}
        </Button>
      </div>
    </form>
  )
}
