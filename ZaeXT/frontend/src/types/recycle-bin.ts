import type { ConversationSummary } from './conversation'

export interface RecycleBinItem extends ConversationSummary {
  deleted_at: string
}

export type RestoreConversationParams = { id: number }
