import client from './client'
import type { LeagueEvent, EventDetail, LeagueConfig } from '../types'

export const listEvents = (leagueId: number) =>
  client.get<LeagueEvent[]>(`/public/leagues/${leagueId}/events`)

export const getEvent = (leagueId: number, eventId: number) =>
  client.get<EventDetail>(`/public/leagues/${leagueId}/events/${eventId}`)

export const createDraftEvent = (leagueId: number, data: { title: string; startDate: string; endDate: string }) =>
  client.post<LeagueEvent>(`/secured/leagues/${leagueId}/events`, data)

export const updateEventConfig = (leagueId: number, eventId: number, config: Partial<LeagueConfig>) =>
  client.put<LeagueEvent>(`/secured/leagues/${leagueId}/events/${eventId}/config`, config)

export const startEvent = (leagueId: number, eventId: number) =>
  client.post<LeagueEvent>(`/secured/leagues/${leagueId}/events/${eventId}/start`, {})

export const finishEvent = (leagueId: number, eventId: number) =>
  client.post<LeagueEvent>(`/secured/leagues/${leagueId}/events/${eventId}/finish`, {})

export const createNextEvent = (leagueId: number, eventId: number) =>
  client.post<LeagueEvent>(`/secured/leagues/${leagueId}/events/${eventId}/next`, {})
