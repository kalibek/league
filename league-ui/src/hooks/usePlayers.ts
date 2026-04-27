import { useState, useEffect, useCallback, useRef } from 'react'
import { listPlayers, getPlayer, createPlayer, listPlayerEvents } from '../api/players'
import type { User, PlayerEventSummary } from '../types'
import type { PlayerDetail } from '../api/players'
import { extractErrorMessage } from './utils'
import { useDebounce } from './useDebounce'

export function usePlayers(params?: { q?: string; sort?: string; limit?: number; offset?: number }) {
  const [players, setPlayers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  const debouncedQ = useDebounce(params?.q ?? '', 300)

  useEffect(() => {
    let cancelled = false
    listPlayers({ ...params, q: debouncedQ || undefined })
      .then((res) => { if (!cancelled) { setPlayers(res.data ?? []); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedQ, params?.sort, params?.limit, params?.offset, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { players, loading, error, refresh }
}

export function usePlayer(id: number) {
  const [player, setPlayer] = useState<PlayerDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    getPlayer(id)
      .then((res) => { if (!cancelled) { setPlayer(res.data); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [id])

  return { player, loading, error }
}

export function useCreatePlayer() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const create = async (data: { firstName: string; lastName: string; email: string }) => {
    setLoading(true)
    setError(null)
    try {
      const res = await createPlayer(data)
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

const EVENTS_PAGE_SIZE = 5

export function usePlayerEvents(userId: number) {
  const [events, setEvents] = useState<PlayerEventSummary[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const loaded = useRef(0)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    loaded.current = 0
    listPlayerEvents(userId, EVENTS_PAGE_SIZE, 0)
      .then((res) => {
        if (!cancelled) {
          const page = res.data
          setEvents(page.events)
          setTotal(page.total)
          loaded.current = page.events.length
          setLoading(false)
        }
      })
      .catch(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [userId, tick])

  const loadMore = useCallback(() => {
    setLoading(true)
    listPlayerEvents(userId, EVENTS_PAGE_SIZE, loaded.current)
      .then((res) => {
        const page = res.data
        setEvents((prev) => [...prev, ...page.events])
        setTotal(page.total)
        loaded.current += page.events.length
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [userId])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return {
    events,
    total,
    loadMore,
    refresh,
    loading,
    hasMore: events.length < total,
  }
}
