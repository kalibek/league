import { useParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useLeague } from '../hooks/useLeagues'
import { useEvents, useCreateDraftEvent, useStartEvent, useCreateNextEvent } from '../hooks/useEvents'
import { formatDate } from '../hooks/utils'
import { Badge } from '../components/Badge/Badge'
import { Button } from '../components/Button/Button'
import { useAuth } from '../hooks/useAuth'
import { useState } from 'react'

const inputStyle: React.CSSProperties = {
  border: '1.5px solid var(--border)',
  borderRadius: 8,
  padding: '9px 12px',
  fontSize: 14,
  color: 'var(--dark)',
  backgroundColor: '#fff',
  outline: 'none',
  width: '100%',
}

export function LeaguePage() {
  const { t } = useTranslation()
  const { id } = useParams<{ id: string }>()
  const leagueId = Number(id)
  const { league, loading: leagueLoading } = useLeague(leagueId)
  const { events, loading: eventsLoading, refresh: refreshEvents } = useEvents(leagueId)
  const { isMaintainer } = useAuth()
  const { create: createDraft, loading: creating } = useCreateDraftEvent()
  const { start, loading: starting } = useStartEvent()
  const { createNext, loading: creatingNext, error: nextError } = useCreateNextEvent()

  const [showCreateForm, setShowCreateForm] = useState(false)
  const [form, setForm] = useState({ title: '', startDate: '', endDate: '' })

  const canManage = isMaintainer(leagueId)

  const handleCreateDraft = async (e: React.FormEvent) => {
    e.preventDefault()
    const ev = await createDraft(leagueId, form)
    if (ev) {
      setShowCreateForm(false)
      refreshEvents()
    }
  }

  const handleStartEvent = async (eventId: number) => {
    await start(leagueId, eventId)
    refreshEvents()
  }

  const handleCreateNext = async (eventId: number) => {
    const ev = await createNext(leagueId, eventId)
    if (ev) refreshEvents()
  }

  if (leagueLoading) return (
    <div style={{ padding: '48px 16px', textAlign: 'center', color: '#94a3b8', fontSize: 14 }}>
      {t('league.loading')}
    </div>
  )
  if (!league) return (
    <div style={{ padding: '48px 16px', textAlign: 'center', color: '#dc2626', fontSize: 14 }}>
      {t('league.notFound')}
    </div>
  )

  return (
    <div className="max-w-4xl mx-auto py-10 px-4">
      <Link
        to="/leagues"
        style={{ fontSize: 13, color: '#64748b', textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: 4, marginBottom: 24 }}
        className="hover:text-[#0B3C5D] transition-colors"
      >
        {t('league.backToLeagues')}
      </Link>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 28, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.5px', marginBottom: 6 }}>
            {league.title}
          </h1>
          {league.description && (
            <p style={{ fontSize: 14, color: '#64748b' }}>{league.description}</p>
          )}
        </div>
        {canManage && (
          <Link
            to={`/leagues/${leagueId}/config`}
            style={{ fontSize: 13, color: 'var(--orange)', fontWeight: 600, textDecoration: 'none' }}
            className="hover:opacity-80 transition-opacity"
          >
            {t('league.configure')}
          </Link>
        )}
      </div>

      {/* League config pills */}
      <div
        style={{
          display: 'flex',
          gap: 8,
          marginBottom: 28,
          flexWrap: 'wrap',
        }}
      >
        {[
          { label: t('league.advances'), value: league.configuration.numberOfAdvances },
          { label: t('league.recedes'), value: league.configuration.numberOfRecedes },
          { label: t('league.gamesToWin'), value: league.configuration.gamesToWin },
          { label: t('league.groupSize'), value: league.configuration.groupSize },
        ].map(({ label, value }) => (
          <div
            key={label}
            style={{
              backgroundColor: '#fff',
              border: '1px solid var(--border)',
              borderRadius: 8,
              padding: '8px 14px',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              minWidth: 90,
            }}
          >
            <span style={{ fontSize: 20, fontWeight: 800, color: 'var(--navy)' }}>{value}</span>
            <span style={{ fontSize: 10, color: '#94a3b8', fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase', marginTop: 2 }}>
              {label}
            </span>
          </div>
        ))}
      </div>

      <div className="flex items-center justify-between mb-4">
        <h2 style={{ fontSize: 18, fontWeight: 700, color: 'var(--navy)' }}>{t('league.events')}</h2>
        {canManage && (
          <Button variant="primary" onClick={() => setShowCreateForm(true)}>
            {t('league.createEvent')}
          </Button>
        )}
      </div>

      {showCreateForm && (
        <form
          onSubmit={handleCreateDraft}
          style={{
            marginBottom: 16,
            padding: '18px 20px',
            borderRadius: 12,
            border: '1.5px solid var(--border)',
            backgroundColor: '#fff',
            display: 'flex',
            flexDirection: 'column',
            gap: 12,
            boxShadow: '0 2px 8px rgba(11,60,93,0.06)',
          }}
        >
          <input
            placeholder={t('league.eventTitlePlaceholder')}
            style={inputStyle}
            className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
            value={form.title}
            onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
            required
          />
          <div style={{ display: 'flex', gap: 10 }}>
            <input
              type="date"
              style={{ ...inputStyle, flex: 1 }}
              className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
              value={form.startDate}
              onChange={(e) => setForm((f) => ({ ...f, startDate: e.target.value }))}
              required
            />
            <input
              type="date"
              style={{ ...inputStyle, flex: 1 }}
              className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
              value={form.endDate}
              onChange={(e) => setForm((f) => ({ ...f, endDate: e.target.value }))}
              required
            />
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button type="button" variant="secondary" onClick={() => setShowCreateForm(false)}>{t('league.cancel')}</Button>
            <Button type="submit" loading={creating}>{t('league.create')}</Button>
          </div>
        </form>
      )}

      {eventsLoading && <p style={{ color: '#94a3b8', fontSize: 13 }}>{t('league.loadingEvents')}</p>}
      {nextError && (
        <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '10px 14px', fontSize: 13, marginBottom: 12 }}>
          {nextError}
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {events.map((ev) => (
          <div
            key={ev.eventId}
            style={{
              borderRadius: 12,
              border: '1px solid var(--border)',
              backgroundColor: '#fff',
              padding: '14px 18px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              boxShadow: '0 1px 4px rgba(11,60,93,0.04)',
            }}
          >
            <div>
              <Link
                 to={`/leagues/${leagueId}/events/${ev.eventId}`}
                 style={{ fontSize: 13, color: 'var(--orange)', fontWeight: 600, textDecoration: 'none' }}
                 className="hover:opacity-80"
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                  <span style={{ fontWeight: 600, color: 'var(--navy)', fontSize: 15 }}>{ev.title}</span>
                  <Badge variant={ev.status} />
                </div>
              </Link>
              <p style={{ fontSize: 12, color: '#94a3b8' }}>
                {formatDate(ev.startDate)} — {formatDate(ev.endDate)}
              </p>
            </div>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <Link
                to={`/leagues/${leagueId}/events/${ev.eventId}`}
                style={{ fontSize: 13, color: 'var(--orange)', fontWeight: 600, textDecoration: 'none' }}
                className="hover:opacity-80"
              >
                {t('league.view')}
              </Link>
              {canManage && ev.status === 'DRAFT' && (
                <Link
                  to={`/leagues/${leagueId}/events/${ev.eventId}/setup`}
                  style={{ fontSize: 13, color: 'var(--navy)', fontWeight: 600, textDecoration: 'none' }}
                  className="hover:opacity-80"
                >
                  {t('league.setup')}
                </Link>
              )}
              {canManage && ev.status === 'DRAFT' && (
                <Button variant="primary" onClick={() => handleStartEvent(ev.eventId)} loading={starting}>
                  {t('league.start')}
                </Button>
              )}
              {canManage && ev.status === 'DONE' && (
                <Button variant="primary" onClick={() => handleCreateNext(ev.eventId)} loading={creatingNext}>
                  {t('league.createNext')}
                </Button>
              )}
            </div>
          </div>
        ))}
        {!eventsLoading && events.length === 0 && (
          <p style={{ color: '#94a3b8', fontSize: 13, textAlign: 'center', padding: '32px 0' }}>
            {t('league.noEventsYet')}
          </p>
        )}
      </div>
    </div>
  )
}
