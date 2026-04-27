import { useState, useEffect, useCallback } from 'react'
import {
  listEvents,
  getEvent,
  createDraftEvent,
  updateEventConfig,
  startEvent,
  finishEvent,
  createNextEvent,
} from '../api/events'
import type { LeagueEvent, EventDetail, LeagueConfig } from '../types'
import { extractErrorMessage } from './utils'

export function useEvents(leagueId: number) {
  const [events, setEvents] = useState<LeagueEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    listEvents(leagueId)
      .then((res) => { if (!cancelled) { setEvents(res.data ?? []); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [leagueId, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { events, loading, error, refresh }
}

export function useEvent(leagueId: number, eventId: number) {
  const [event, setEvent] = useState<EventDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    getEvent(leagueId, eventId)
      .then((res) => { if (!cancelled) { setEvent(res.data); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [leagueId, eventId, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { event, setEvent, loading, error, refresh }
}

export function useCreateDraftEvent() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const create = async (
    leagueId: number,
    data: { title: string; startDate: string; endDate: string }
  ) => {
    setLoading(true)
    setError(null)
    try {
      const res = await createDraftEvent(leagueId, data)
      return res.data
    } catch (e) {
      setError(extractErrorMessage(e))
      return null
    } finally {
      setLoading(false)
    }
  }

  return { create, loading, error }
}

export function useUpdateEventConfig() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const update = async (leagueId: number, eventId: number, config: Partial<LeagueConfig>) => {
    setLoading(true)
    setError(null)
    try {
      const res = await updateEventConfig(leagueId, eventId, config)
      return res.data
    } catch (e) {
      setError(extractErrorMessage(e))
      return null
    } finally {
      setLoading(false)
    }
  }

  return { update, loading, error }
}

export function useStartEvent() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const start = async (leagueId: number, eventId: number) => {
    setLoading(true)
    setError(null)
    try {
      const res = await startEvent(leagueId, eventId)
      return res.data
    } catch (e) {
      setError(extractErrorMessage(e))
      return null
    } finally {
      setLoading(false)
    }
  }

  return { start, loading, error }
}

export function useFinishEvent() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const finish = async (leagueId: number, eventId: number) => {
    setLoading(true)
    setError(null)
    try {
      const res = await finishEvent(leagueId, eventId)
      return res.data
    } catch (e) {
      setError(extractErrorMessage(e))
      return null
    } finally {
      setLoading(false)
    }
  }

  return { finish, loading, error }
}

export function useCreateNextEvent() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const createNext = async (leagueId: number, eventId: number) => {
    setLoading(true)
    setError(null)
    try {
      const res = await createNextEvent(leagueId, eventId)
      return res.data
    } catch (e) {
      setError(extractErrorMessage(e))
      return null
    } finally {
      setLoading(false)
    }
  }

  return { createNext, loading, error }
}
