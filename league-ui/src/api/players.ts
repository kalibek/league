import client from './client'
import type { User, RatingHistory, GroupDetail, PlayerEventsPage, PlayerProfileDetail } from '../types'

export interface PlayerDetail extends User {
  ratingHistory: RatingHistory[]
  groups: GroupDetail[]
  profile?: PlayerProfileDetail
}

export const listPlayers = (params?: { q?: string; sort?: string; limit?: number; offset?: number }) =>
  client.get<User[]>('/public/players', { params })

export const getPlayer = (id: number) => client.get<PlayerDetail>(`/public/players/${id}`)

export const createPlayer = (data: { firstName: string; lastName: string; email: string }) =>
  client.post<User>('/secured/players', data)

export const listPlayerEvents = (id: number, limit = 5, offset = 0) =>
  client.get<PlayerEventsPage>(`/public/players/${id}/events`, { params: { limit, offset } })

export const importCSV = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return client.post<{ imported: number; skipped: number; errors: string[] }>('/secured/players/import', form)
}
