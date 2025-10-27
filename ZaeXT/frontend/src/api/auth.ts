import type { LoginPayload, LoginResponse, RegisterPayload } from '@/types'

import { http } from './http'

export const register = (payload: RegisterPayload) => http.post<void, void>('/register', payload)

export const login = (payload: LoginPayload) =>
  http.post<LoginResponse, LoginResponse>('/login', payload)
