import { useState, useEffect, useCallback } from 'react'
import { listGroups, getGroup, createGroup, seedPlayer, removeGroupPlayer, finishGroup, reopenGroup, addPlayer, markNoShow, setManualPlace } from '../api/groups'
import type { Group, GroupDetail } from '../types'
import { extractErrorMessage } from './utils'

export function useGroups(eventId: number) {
  const [groups, setGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    listGroups(eventId)
      .then((res) => setGroups(res.data ?? []))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
  }, [eventId])

  useEffect(() => { load() }, [load])

  return { groups, loading, error, refresh: load }
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
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    getGroup(eventId, groupId)
      .then((res) => setGroup(res.data))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
  }, [eventId, groupId])

  useEffect(() => {
    load()
  }, [load])

  return { group, loading, error, refresh: load }
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

export function useMarkNoShow() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const noShow = async (eventId: number, groupId: number, groupPlayerId: number) => {
    setLoading(true)
    setError(null)
    try {
      await markNoShow(eventId, groupId, groupPlayerId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { noShow, loading, error }
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
