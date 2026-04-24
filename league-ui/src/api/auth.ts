import client from './client'
import type { User, UserRole } from '../types'

export interface MeResponse extends User {
  roles: UserRole[]
}

export const getMe = () => client.get<MeResponse>('/auth/me')

export const registerEmail = (data: {
  firstName: string
  lastName: string
  email: string
  password: string
}) => client.post<User>('/auth/register', data)

export const loginEmail = (data: { email: string; password: string }) =>
  client.post<User>('/auth/login/email', data)

export const logout = () => client.post('/auth/logout')
