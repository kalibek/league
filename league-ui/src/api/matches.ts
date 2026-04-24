import client from './client'
import type { Match } from '../types'

export const updateMatchScore = (
  groupId: number,
  matchId: number,
  data: { score1: number; score2: number }
) =>
  client.put<Match>(`/secured/groups/${groupId}/matches/${matchId}`, data)
