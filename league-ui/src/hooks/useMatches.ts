import { useState } from 'react'
import { updateMatchScore } from '../api/matches'
import type { Match } from '../types'
import { extractErrorMessage } from './utils'

export function useUpdateMatchScore() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const update = async (
    groupId: number,
    matchId: number,
    data: { score1: number; score2: number }
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
