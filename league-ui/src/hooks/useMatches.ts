import { useState, useEffect, useCallback } from 'react'
import { updateMatchScore, setMatchTableNumber, getTablesInUse, resetMatchScore } from '../api/matches'
import type { Match } from '../types'
import { extractErrorMessage } from './utils'

export function useUpdateMatchScore() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const update = async (
    groupId: number,
    matchId: number,
    data: { score1: number; score2: number,
      withdraw1: boolean; withdraw2: boolean,
    }
  ): Promise<Match | null> => {
    setLoading(true)
    setError(null)
    try {
      const res = await updateMatchScore(groupId, matchId, data)
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

export function useSetTableNumber() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const assign = async (groupId: number, matchId: number, tableNumber: number): Promise<boolean> => {
    setLoading(true)
    setError(null)
    try {
      await setMatchTableNumber(groupId, matchId, tableNumber)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { assign, loading, error }
}

export function useResetMatchScore() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const reset = async (groupId: number, matchId: number): Promise<boolean> => {
    setLoading(true)
    setError(null)
    try {
      await resetMatchScore(groupId, matchId)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { reset, loading, error }
}

export function useTablesInUse(eventId: number) {
  const [tablesInUse, setTablesInUse] = useState<number[]>([])
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    getTablesInUse(eventId)
      .then((res) => { if (!cancelled) setTablesInUse(res.data.tablesInUse ?? []) })
      .catch(() => { if (!cancelled) setTablesInUse([]) })
    return () => { cancelled = true }
  }, [eventId, tick])

  const refresh = useCallback(() => setTick((t) => t + 1), [])

  return { tablesInUse, refresh }
}
