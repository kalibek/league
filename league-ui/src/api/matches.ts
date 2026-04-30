import client from './client'
import type { Match } from '../types'

export const updateMatchScore = (
  groupId: number,
  matchId: number,
  data: { score1: number; score2: number }
) =>
  client.put<Match>(`/secured/groups/${groupId}/matches/${matchId}`, data)

export const setMatchTableNumber = (groupId: number, matchId: number, tableNumber: number) =>
  client.put(`/secured/groups/${groupId}/matches/${matchId}/table`, { tableNumber })

export const resetMatchScore = (groupId: number, matchId: number) =>
  client.delete(`/secured/groups/${groupId}/matches/${matchId}/score`)

export const getTablesInUse = (eventId: number) =>
  client.get<{ tablesInUse: number[] }>(`/public/events/${eventId}/tables-in-use`)
