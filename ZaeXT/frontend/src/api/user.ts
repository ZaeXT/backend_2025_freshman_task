import type { UpdateMemoryPayload, UserProfile } from '@/types'

import { http } from './http'

export const getProfile = () => http.get<UserProfile, UserProfile>('/profile')

export const updateMemory = (payload: UpdateMemoryPayload) =>
  http.put<void, void>('/profile/memory', payload)
