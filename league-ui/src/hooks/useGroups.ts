import { useState, useEffect, useCallback } from 'react'
import { listGroups, getGroup, createGroup, seedPlayer, removeGroupPlayer, finishGroup, reopenGroup, addPlayer, setManualPlace, setPlayerStatus, addPlayerToActiveGroup } from '../api/groups'
import type { Group, GroupDetail, PlayerStatus } from '../types'
import { extractErrorMessage } from './utils'

export function useGroups(eventId: number) {
  const [groups, setGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    listGroups(eventId)
      .then((res) => { if (!cancelled) { setGroups(res.data ?? []); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [eventId, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { groups, loading, error, refresh }
}

export function useCreateGroup() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const create = async (
    eventId: number,
    data: { division: string; groupNo: number; scheduled: string }
  ) => {
    setLoading(true)
    setError(null)
    try {
      const res = await createGroup(eventId, data)
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

export function useSeedPlayer() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const seed = async (eventId: number, groupId: number, userId: number) => {
    setLoading(true)
    setError(null)
    try {
      await seedPlayer(eventId, groupId, userId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { seed, loading, error }
}

export function useRemoveGroupPlayer() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const remove = async (eventId: number, groupId: number, groupPlayerId: number) => {
    setLoading(true)
    setError(null)
    try {
      await removeGroupPlayer(eventId, groupId, groupPlayerId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { remove, loading, error }
}

export function useGroup(eventId: number, groupId: number) {
  const [group, setGroup] = useState<GroupDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    getGroup(eventId, groupId)
      .then((res) => { if (!cancelled) { setGroup(res.data); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [eventId, groupId, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { group, loading, error, refresh }
}

export function useFinishGroup() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const finish = async (eventId: number, groupId: number) => {
    setLoading(true)
    setError(null)
    try {
      const res = await finishGroup(eventId, groupId)
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

export function useReopenGroup() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const reopen = async (eventId: number, groupId: number) => {
    setLoading(true)
    setError(null)
    try {
      await reopenGroup(eventId, groupId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { reopen, loading, error }
}

export function useAddPlayer() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const add = async (eventId: number, groupId: number, userId: number) => {
    setLoading(true)
    setError(null)
    try {
      await addPlayer(eventId, groupId, userId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { add, loading, error }
}

export function useSetManualPlace() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const setPlace = async (eventId: number, groupId: number, orderedPlayerIds: number[]) => {
    setLoading(true)
    setError(null)
    try {
      await setManualPlace(eventId, groupId, orderedPlayerIds)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { setPlace, loading, error }
}

export function useSetPlayerStatus() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const setStatus = async (
    eventId: number,
    groupId: number,
    groupPlayerId: number,
    status: PlayerStatus
  ) => {
    setLoading(true)
    setError(null)
    try {
      await setPlayerStatus(eventId, groupId, groupPlayerId, status)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { setStatus, loading, error }
}

export function useAddPlayerToActiveGroup() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const addActive = async (eventId: number, groupId: number, userId: number) => {
    setLoading(true)
    setError(null)
    try {
      await addPlayerToActiveGroup(eventId, groupId, userId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { addActive, loading, error }
}
