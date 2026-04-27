import client from './client'
import type { Group, GroupDetail } from '../types'

export const listGroups = (eventId: number) =>
  client.get<Group[]>(`/public/events/${eventId}/groups`)

export const getGroup = (eventId: number, groupId: number) =>
  client.get<GroupDetail>(`/public/events/${eventId}/groups/${groupId}`)

export const createGroup = (
  eventId: number,
  data: { division: string; groupNo: number; scheduled: string }
) => client.post<Group>(`/secured/events/${eventId}/groups`, data)

export const seedPlayer = (eventId: number, groupId: number, userId: number) =>
  client.post(`/secured/events/${eventId}/groups/${groupId}/seed`, { userId })

export const removeGroupPlayer = (eventId: number, groupId: number, groupPlayerId: number) =>
  client.delete(`/secured/events/${eventId}/groups/${groupId}/players/${groupPlayerId}`)

export const finishGroup = (eventId: number, groupId: number) =>
  client.post<GroupDetail>(`/secured/events/${eventId}/groups/${groupId}/finish`, {})

export const reopenGroup = (eventId: number, groupId: number) =>
  client.post(`/secured/events/${eventId}/groups/${groupId}/reopen`, {})

export const addPlayer = (eventId: number, groupId: number, userId: number) =>
  client.post(`/secured/events/${eventId}/groups/${groupId}/players`, { userId })

export const setManualPlace = (
  eventId: number,
  groupId: number,
  orderedPlayerIds: number[]
) =>
  client.put(`/secured/events/${eventId}/groups/${groupId}/placement`, { orderedPlayerIds })
