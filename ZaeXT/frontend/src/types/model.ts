import type { UserTier } from './user'

export interface ModelInfo {
  id: string
  name: string
  tier?: UserTier
  description?: string
  tags?: string[]
}
