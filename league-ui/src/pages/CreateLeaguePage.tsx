import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useCreateLeague } from '../hooks/useLeagues'
import { Button } from '../components/Button/Button'
import { Input } from '../components/Input/Input'
import { LeagueConfigForm } from '../components/LeagueConfigForm/LeagueConfigForm'
import type { LeagueConfig } from '../types'

const defaultConfig: LeagueConfig = {
  numberOfAdvances: 2,
  numberOfRecedes: 2,
  gamesToWin: 3,
  groupSize: 6,
}

export function CreateLeaguePage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { create, loading, error } = useCreateLeague()
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [config, setConfig] = useState<LeagueConfig>(defaultConfig)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const league = await create({ title, description, configuration: config })
    if (league) navigate(`/leagues/${league.leagueId}`)
  }

  return (
    <div className="max-w-xl mx-auto py-10 px-4">
      <Link
        to="/leagues"
        style={{ fontSize: 13, color: '#64748b', textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: 4, marginBottom: 24 }}
        className="hover:text-[#0B3C5D] transition-colors"
      >
        {t('createLeague.backToLeagues')}
      </Link>
      <h1 style={{ fontSize: 26, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.5px', marginBottom: 24 }}>
        {t('createLeague.title')}
      </h1>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
        <Input
          label={t('createLeague.leagueName')}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
        <Input
          label={t('createLeague.description')}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <div>
          <h2 style={{ fontSize: 13, fontWeight: 700, color: '#64748b', letterSpacing: '0.05em', textTransform: 'uppercase', marginBottom: 12 }}>
            {t('createLeague.configuration')}
          </h2>
          <LeagueConfigForm
            initial={config}
            onSubmit={setConfig}
            loading={false}
            embedded
          />
        </div>
        {error && (
          <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '10px 14px', fontSize: 13 }}>
            {error}
          </div>
        )}
        <div style={{ display: 'flex', gap: 10 }}>
          <Button type="button" variant="secondary" onClick={() => navigate('/leagues')}>
            {t('createLeague.cancel')}
          </Button>
          <Button type="submit" loading={loading}>
            {t('createLeague.createLeague')}
          </Button>
        </div>
      </form>
    </div>
  )
}
