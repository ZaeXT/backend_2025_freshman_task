export type UserTier = 'free' | 'pro' | 'enterprise' | 'internal' | string

export interface UserProfile {
  id: number
  username: string
  tier: UserTier
  memory_info: string
  created_at: string
}

export interface UpdateMemoryPayload {
  memory_info: string
}
