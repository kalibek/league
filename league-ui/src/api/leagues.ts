import client from './client'
import type { League, LeagueConfig } from '../types'

export const listLeagues = () => client.get<League[]>('/public/leagues')

export const getLeague = (id: number) => client.get<League>(`/public/leagues/${id}`)

export const createLeague = (data: { title: string; description: string; configuration: LeagueConfig }) =>
  client.post<League>('/secured/leagues', data)

export const updateConfig = (leagueId: number, config: LeagueConfig) =>
  client.put<League>(`/secured/leagues/${leagueId}/config`, config)

export const listRoles = (leagueId: number) =>
  client.get(`/secured/leagues/${leagueId}/roles`)

export const assignRole = (leagueId: number, data: { userId: number; roleName: string }) =>
  client.post(`/secured/leagues/${leagueId}/roles`, data)

export const removeRole = (leagueId: number, userId: number, roleName: string) =>
  client.delete(`/secured/leagues/${leagueId}/roles`, { data: { userId, roleName } })
