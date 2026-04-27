import { useState, useEffect, useCallback } from 'react'
import { listLeagues, getLeague, createLeague, updateConfig, listRoles, assignRole, removeRole } from '../api/leagues'
import type { League, LeagueConfig } from '../types'
import { extractErrorMessage } from './utils'

export interface LeagueRoleEntry {
  userId: number
  leagueId: number
  roleName: string
  firstName: string
  lastName: string
  email: string
}

export function useLeagues() {
  const [leagues, setLeagues] = useState<League[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    listLeagues()
      .then((res) => { if (!cancelled) { setLeagues(res.data ?? []); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { leagues, loading, error, refresh }
}

export function useLeague(id: number) {
  const [league, setLeague] = useState<League | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    getLeague(id)
      .then((res) => { if (!cancelled) { setLeague(res.data); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [id, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { league, loading, error, refresh }
}

export function useCreateLeague() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const create = async (data: { title: string; description: string; configuration: LeagueConfig }) => {
    setLoading(true)
    setError(null)
    try {
      const res = await createLeague(data)
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

export function useUpdateConfig() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const update = async (leagueId: number, config: LeagueConfig) => {
    setLoading(true)
    setError(null)
    try {
      const res = await updateConfig(leagueId, config)
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

export function useLeagueRoles(leagueId: number) {
  const [roles, setRoles] = useState<LeagueRoleEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    listRoles(leagueId)
      .then((res) => { if (!cancelled) { setRoles(res.data ?? []); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [leagueId, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { roles, loading, error, refresh }
}

export function useAssignRole() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const assign = async (leagueId: number, data: { userId: number; roleName: string }) => {
    setLoading(true)
    setError(null)
    try {
      await assignRole(leagueId, data)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  const remove = async (leagueId: number, userId: number, roleName: string) => {
    setLoading(true)
    setError(null)
    try {
      await removeRole(leagueId, userId, roleName)
      return true
    } catch (e) {
      setError(extractErrorMessage(e))
      return false
    } finally {
      setLoading(false)
    }
  }

  return { assign, remove, loading, error }
}
