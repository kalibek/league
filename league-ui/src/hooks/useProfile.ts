import { useState, useEffect, useCallback } from 'react'
import {
  getMyProfile,
  upsertMyProfile,
  listCountries,
  listCities,
  addCity,
  listBlades,
  addBlade,
  listRubbers,
  addRubber,
} from '../api/profile'
import type { PlayerProfileDetail, Country, City, Blade, Rubber } from '../types'
import { extractErrorMessage } from './utils'

export function useMyProfile() {
  const [profile, setProfile] = useState<PlayerProfileDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    getMyProfile()
      .then((res) => { if (!cancelled) { setProfile(res.data); setError(null); setLoading(false) } })
      .catch((e) => { if (!cancelled) { setError(extractErrorMessage(e)); setLoading(false) } })
    return () => { cancelled = true }
  }, [tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  return { profile, setProfile, loading, error, refresh }
}

export function useUpsertProfile() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const save = async (data: Parameters<typeof upsertMyProfile>[0]) => {
    setLoading(true)
    setError(null)
    try {
      const res = await upsertMyProfile(data)
      return res.data
    } catch (e) {
      setError(extractErrorMessage(e))
      return null
    } finally {
      setLoading(false)
    }
  }

  return { save, loading, error }
}

export function useCountries() {
  const [countries, setCountries] = useState<Country[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listCountries()
      .then((res) => setCountries(res.data))
      .finally(() => setLoading(false))
  }, [])

  return { countries, loading }
}

export function useCities(countryId: number | null) {
  const [cities, setCities] = useState<City[]>([])
  const [loading, setLoading] = useState(false)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    if (!countryId) {
      Promise.resolve().then(() => setCities([]))
      return
    }
    let cancelled = false
    listCities(countryId)
      .then((res) => { if (!cancelled) { setCities(res.data); setLoading(false) } })
      .catch(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [countryId, tick])

  const refresh = useCallback(() => { setLoading(true); setTick((t) => t + 1) }, [])

  const add = async (name: string): Promise<City | null> => {
    if (!countryId) return null
    try {
      const res = await addCity(countryId, name)
      const city = res.data
      setCities((prev) => [...prev, city].sort((a, b) => a.name.localeCompare(b.name)))
      return city
    } catch {
      return null
    }
  }

  return { cities, loading, add, refresh }
}

export function useBlades() {
  const [blades, setBlades] = useState<Blade[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listBlades()
      .then((res) => setBlades(res.data))
      .finally(() => setLoading(false))
  }, [])

  const add = async (name: string): Promise<Blade | null> => {
    try {
      const res = await addBlade(name)
      const blade = res.data
      setBlades((prev) => [...prev, blade].sort((a, b) => a.name.localeCompare(b.name)))
      return blade
    } catch {
      return null
    }
  }

  return { blades, loading, add }
}

export function useRubbers() {
  const [rubbers, setRubbers] = useState<Rubber[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listRubbers()
      .then((res) => setRubbers(res.data))
      .finally(() => setLoading(false))
  }, [])

  const add = async (name: string): Promise<Rubber | null> => {
    try {
      const res = await addRubber(name)
      const rubber = res.data
      setRubbers((prev) => [...prev, rubber].sort((a, b) => a.name.localeCompare(b.name)))
      return rubber
    } catch {
      return null
    }
  }

  return { rubbers, loading, add }
}
