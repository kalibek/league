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
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    listLeagues()
      .then((res) => setLeagues(res.data ?? []))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  return { leagues, loading, error, refresh: load }
}

export function useLeague(id: number) {
  const [league, setLeague] = useState<League | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    getLeague(id)
      .then((res) => setLeague(res.data))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  return { league, loading, error, refresh: load }
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
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    listRoles(leagueId)
      .then((res) => setRoles(res.data ?? []))
      .catch((e) => setError(extractErrorMessage(e)))
      .finally(() => setLoading(false))
  }, [leagueId])

  useEffect(() => {
    load()
  }, [load])

  return { roles, loading, error, refresh: load }
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
