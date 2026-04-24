import { useState, useEffect, useCallback, useRef } from 'react'
import { listPlayers, getPlayer, createPlayer, listPlayerEvents } from '../api/players'
import type { User, PlayerEventSummary } from '../types'
import type { PlayerDetail } from '../api/players'
import { extractErrorMessage } from './utils'
import { useDebounce } from './useDebounce'

export function usePlayers(params?: { q?: string; sort?: string; limit?: number; offset?: number }) {
  const [players, setPlayers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const debouncedQ = useDebounce(params?.q ?? '', 300)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    listPlayers({ ...params, q: debouncedQ || undefined })
      .then((res) => setPlayers(res.data ?? []))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedQ, params?.sort, params?.limit, params?.offset])

  useEffect(() => {
    load()
  }, [load])

  return { players, loading, error, refresh: load }
}

export function usePlayer(id: number) {
  const [player, setPlayer] = useState<PlayerDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    getPlayer(id)
      .then((res) => setPlayer(res.data))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
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
  const [loading, setLoading] = useState(false)
  const loaded = useRef(0)

  const fetch = useCallback(async (offset: number, append: boolean) => {
    setLoading(true)
    try {
      const res = await listPlayerEvents(userId, EVENTS_PAGE_SIZE, offset)
      const page = res.data
      setEvents((prev) => (append ? [...prev, ...page.events] : page.events))
      setTotal(page.total)
      loaded.current = offset + page.events.length
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }, [userId])

  useEffect(() => {
    loaded.current = 0
    fetch(0, false)
  }, [fetch])

  const loadMore = useCallback(() => fetch(loaded.current, true), [fetch])

  return {
    events,
    total,
    loadMore,
    loading,
    hasMore: events.length < total,
  }
}
