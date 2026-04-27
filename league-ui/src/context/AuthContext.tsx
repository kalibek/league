import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import type { ReactNode } from 'react'
import { getMe, logout as logoutApi } from '../api/auth'
import type { User, UserRole } from '../types'

interface AuthContextValue {
  user: User | null
  roles: Record<number, string[]>
  isAdmin: boolean
  isMaintainer: (leagueId: number) => boolean
  isUmpire: (leagueId: number) => boolean
  logout: () => void
  refresh: () => void
  loading: boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [roles, setRoles] = useState<Record<number, string[]>>({})
  const [isAdmin, setIsAdmin] = useState(false)
  const [loading, setLoading] = useState(true)
  const [tick, setTick] = useState(0)

  const refresh = useCallback(() => setTick((t) => t + 1), [])

  useEffect(() => {
    getMe()
      .then((res) => {
        const { roles: rawRoles, ...userFields } = res.data
        setUser(userFields as User)
        setIsAdmin(!!(userFields as User).isAdmin)
        const grouped: Record<number, string[]> = {}
        ;(rawRoles ?? []).forEach((r: UserRole) => {
          if (!grouped[r.leagueId]) grouped[r.leagueId] = []
          grouped[r.leagueId].push(r.roleName)
        })
        setRoles(grouped)
        setLoading(false)
      })
      .catch(() => {
        setUser(null)
        setRoles({})
        setIsAdmin(false)
        setLoading(false)
      })
  }, [tick])

  const isMaintainer = useCallback(
    (leagueId: number) => roles[leagueId]?.includes('maintainer') ?? false,
    [roles]
  )

  const isUmpire = useCallback(
    (leagueId: number) =>
      roles[leagueId]?.includes('umpire') || roles[leagueId]?.includes('maintainer') ? true : false,
    [roles]
  )

  const logout = useCallback(() => {
    logoutApi().finally(() => {
      setUser(null)
      setRoles({})
      window.location.href = '/login'
    })
  }, [])

  const value = useMemo(
    () => ({ user, roles, isAdmin, isMaintainer, isUmpire, logout, refresh, loading }),
    [user, roles, isAdmin, isMaintainer, isUmpire, logout, refresh, loading]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuthContext must be used inside AuthProvider')
  return ctx
}
