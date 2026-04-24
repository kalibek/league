import client from './client'
import type { PlayerProfileDetail, Country, City, Blade, Rubber } from '../types'

export const getMyProfile = () =>
  client.get<PlayerProfileDetail>('/secured/profile')

export const upsertMyProfile = (data: {
  firstName?: string
  lastName?: string
  countryId?: number | null
  cityId?: number | null
  birthdate?: string | null
  grip?: string | null
  gender?: string | null
  bladeId?: number | null
  fhRubberId?: number | null
  bhRubberId?: number | null
}) => client.put<PlayerProfileDetail>('/secured/profile', data)

export const listCountries = () =>
  client.get<Country[]>('/public/countries')

export const listCities = (countryId: number) =>
  client.get<City[]>(`/public/countries/${countryId}/cities`)

export const addCity = (countryId: number, name: string) =>
  client.post<City>(`/secured/countries/${countryId}/cities`, { name })

export const listBlades = () =>
  client.get<Blade[]>('/public/blades')

export const addBlade = (name: string) =>
  client.post<Blade>('/secured/blades', { name })

export const listRubbers = () =>
  client.get<Rubber[]>('/public/rubbers')

export const addRubber = (name: string) =>
  client.post<Rubber>('/secured/rubbers', { name })
