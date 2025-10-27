import type {
  AutoClassifyResponse,
  ConversationMessage,
  ConversationSummary,
  CreateConversationPayload,
  UpdateConversationCategoryPayload,
  UpdateConversationTitlePayload,
} from '@/types'

import { http } from './http'

export const createConversation = (payload?: CreateConversationPayload) =>
  http.post<ConversationSummary, ConversationSummary>('/conversations', payload ?? {})

export const fetchConversations = () =>
  http.get<ConversationSummary[], ConversationSummary[]>('/conversations')

export const fetchConversationMessages = (conversationId: number) =>
  http.get<ConversationMessage[], ConversationMessage[]>(
    `/conversations/${conversationId}/messages`,
  )

export const updateConversationTitle = (
  conversationId: number,
  payload: UpdateConversationTitlePayload,
) =>
  http.put<ConversationSummary, ConversationSummary>(
    `/conversations/${conversationId}/title`,
    payload,
  )

export const updateConversationCategory = (
  conversationId: number,
  payload: UpdateConversationCategoryPayload,
) =>
  http.put<ConversationSummary, ConversationSummary>(
    `/conversations/${conversationId}/category`,
    payload,
  )

export const triggerAutoClassify = (conversationId: number) =>
  http.post<AutoClassifyResponse, AutoClassifyResponse>(
    `/conversations/${conversationId}/auto-classify`,
  )

export const deleteConversation = (conversationId: number) =>
  http.delete<void, void>(`/conversations/${conversationId}`)
