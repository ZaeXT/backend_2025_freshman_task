export type MessageRole = 'system' | 'user' | 'assistant'

export interface ConversationSummary {
  id: number
  title: string
  model_id: string | null
  category_id: number | null
  is_temporary: boolean
  created_at: string
  updated_at: string
  last_message_at: string
  last_message_preview?: string
  auto_titled?: boolean
}

export interface ConversationMessage {
  id: number
  conversation_id: number
  role: MessageRole
  content: string
  created_at: string
  metadata?: Record<string, unknown>
}

export interface ConversationDetail extends ConversationSummary {
  messages: ConversationMessage[]
}

export interface CreateConversationPayload {
  is_temporary?: boolean
  category_id?: number
}

export interface UpdateConversationTitlePayload {
  title: string
}

export interface UpdateConversationCategoryPayload {
  category_id: number | null
}

export interface AutoClassifyResponse {
  category_id: number | null
  confidence: number
}

export interface SendMessagePayload {
  message: string
  model_id?: string
  enable_thinking?: boolean
}

export interface StreamingMessage {
  conversation_id: number
  draft_message_id: number
  content: string
}
