import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { LeagueConfig } from '../../types'
import { Button } from '../Button/Button'
import { Input } from '../Input/Input'

interface LeagueConfigFormProps {
  initial: LeagueConfig
  onSubmit: (config: LeagueConfig) => void
  loading?: boolean
  showDraftWarning?: boolean
  embedded?: boolean
}

export function LeagueConfigForm({
  initial,
  onSubmit,
  loading = false,
  showDraftWarning = false,
  embedded = false,
}: LeagueConfigFormProps) {
  const { t } = useTranslation()
  const [config, setConfig] = useState<LeagueConfig>(initial)

  const set = (key: keyof LeagueConfig) => (e: React.ChangeEvent<HTMLInputElement>) => {
    const next = { ...config, [key]: Number(e.target.value) }
    setConfig(next)
    if (embedded) onSubmit(next)
  }

  const fields = (
    <>
      {showDraftWarning && (
        <div className="rounded-md bg-yellow-50 border border-yellow-200 px-4 py-3 text-sm text-yellow-800">
          {t('leagueConfig.draftWarning')}
        </div>
      )}
      <div className="grid grid-cols-2 gap-4">
        <Input
          label={t('leagueConfig.advancesPerGroup')}
          type="number"
          min={0}
          value={config.numberOfAdvances}
          onChange={set('numberOfAdvances')}
          required
        />
        <Input
          label={t('leagueConfig.recedesPerGroup')}
          type="number"
          min={0}
          value={config.numberOfRecedes}
          onChange={set('numberOfRecedes')}
          required
        />
        <Input
          label={t('leagueConfig.gamesToWin')}
          type="number"
          min={1}
          value={config.gamesToWin}
          onChange={set('gamesToWin')}
          required
        />
        <Input
          label={t('leagueConfig.groupSize')}
          type="number"
          min={2}
          value={config.groupSize}
          onChange={set('groupSize')}
          required
        />
        <Input
          label={t('leagueConfig.numberOfTables')}
          type="number"
          min={0}
          value={config.numberOfTables}
          onChange={set('numberOfTables')}
          required
        />
      </div>
      {!embedded && (
        <Button type="submit" loading={loading}>
          {t('leagueConfig.saveConfiguration')}
        </Button>
      )}
    </>
  )

  if (embedded) {
    return <div className="flex flex-col gap-4">{fields}</div>
  }

  return (
    <form onSubmit={(e) => { e.preventDefault(); onSubmit(config) }} className="flex flex-col gap-4">
      {fields}
    </form>
  )
}
