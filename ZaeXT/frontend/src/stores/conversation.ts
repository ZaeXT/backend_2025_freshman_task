import { defineStore } from 'pinia'
import { computed, reactive, ref } from 'vue'

import {
  createConversation,
  deleteConversation,
  fetchConversationMessages,
  fetchConversations,
  triggerAutoClassify,
  updateConversationCategory,
  updateConversationTitle,
} from '@/api'
import { streamConversation } from '@/api/sse'
import type {
  ConversationMessage,
  ConversationSummary,
  CreateConversationPayload,
  SendMessagePayload,
  UpdateConversationCategoryPayload,
  UpdateConversationTitlePayload,
} from '@/types'
import type { StreamController } from '@/types'
import { emitToast } from '@/utils/notifications'

interface ChatMessage extends ConversationMessage {
  status?: 'pending' | 'streaming' | 'done' | 'error'
}

interface StreamingState {
  controller: StreamController | null
  pendingAssistantId: number | null
  isStreaming: boolean
  enableThinking: boolean
  modelId?: string
}

const createAssistantPlaceholder = (conversationId: number): ChatMessage => ({
  id: Date.now() * -1,
  conversation_id: conversationId,
  role: 'assistant',
  content: '',
  created_at: new Date().toISOString(),
  status: 'streaming',
})

const createUserMessage = (conversationId: number, content: string): ChatMessage => ({
  id: Date.now(),
  conversation_id: conversationId,
  role: 'user',
  content,
  created_at: new Date().toISOString(),
  status: 'done',
})

export const useConversationStore = defineStore('conversations', () => {
  const conversations = ref<ConversationSummary[]>([])
  const currentId = ref<number | null>(null)
  const messageMap = ref<Record<number, ChatMessage[]>>({})
  const listLoading = ref(false)
  const messagesLoading = ref(false)
  const searchQuery = ref('')
  const streamingState = reactive<StreamingState>({
    controller: null,
    pendingAssistantId: null,
    isStreaming: false,
    enableThinking: false,
  })

  const filteredConversations = computed(() => {
    if (!searchQuery.value) return conversations.value
    const keyword = searchQuery.value.toLowerCase()
    return conversations.value.filter((item) =>
      [item.title, item.last_message_preview ?? '', item.model_id ?? '']
        .join(' ')
        .toLowerCase()
        .includes(keyword),
    )
  })

  const currentConversation = computed(
    () => conversations.value.find((item) => item.id === currentId.value) ?? null,
  )

  const currentMessages = computed(() => {
    if (!currentId.value) return []
    return messageMap.value[currentId.value] ?? []
  })

  const setMessages = (conversationId: number, messages: ChatMessage[]) => {
    messageMap.value = {
      ...messageMap.value,
      [conversationId]: messages,
    }
  }

  const appendMessage = (conversationId: number, message: ChatMessage) => {
    const existing = messageMap.value[conversationId] ?? []
    setMessages(conversationId, [...existing, message])
  }

  const updateMessage = (
    conversationId: number,
    messageId: number,
    updater: (msg: ChatMessage) => void,
  ) => {
    const existing = messageMap.value[conversationId]
    if (!existing) return
    setMessages(
      conversationId,
      existing.map((msg) => {
        if (msg.id === messageId) {
          const clone = { ...msg }
          updater(clone)
          return clone
        }
        return msg
      }),
    )
  }

  const upsertConversation = (conversation: ConversationSummary) => {
    const index = conversations.value.findIndex((item) => item.id === conversation.id)
    if (index >= 0) {
      conversations.value.splice(index, 1, conversation)
    } else {
      conversations.value = [conversation, ...conversations.value]
    }
  }

  const loadConversations = async () => {
    if (listLoading.value) return
    try {
      listLoading.value = true
      const result = await fetchConversations()
      conversations.value = result.sort(
        (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
      )
      if (!currentId.value && conversations.value.length) {
        currentId.value = conversations.value[0].id
      }

      if (currentId.value && !messageMap.value[currentId.value]) {
        await loadMessages(currentId.value)
      }
    } catch (error) {
      console.error('[conversations] failed to fetch list', error)
    } finally {
      listLoading.value = false
    }
  }

  const loadMessages = async (conversationId: number) => {
    if (messagesLoading.value) return
    try {
      messagesLoading.value = true
      const result = await fetchConversationMessages(conversationId)
      const messages = result.map<ChatMessage>((msg) => ({ ...msg, status: 'done' }))
      setMessages(conversationId, messages)
    } catch (error) {
      console.error('[conversations] failed to fetch messages', error)
    } finally {
      messagesLoading.value = false
    }
  }

  const selectConversation = async (conversationId: number) => {
    if (conversationId === currentId.value) return
    currentId.value = conversationId
    if (!messageMap.value[conversationId]) {
      await loadMessages(conversationId)
    }
  }

  const createNewConversation = async (payload?: CreateConversationPayload) => {
    const conversation = await createConversation(payload)
    upsertConversation(conversation)
    currentId.value = conversation.id
    setMessages(conversation.id, [])
    emitToast({ type: 'success', message: '已创建新的对话。' })
    return conversation
  }

  const ensureConversation = async () => {
    if (currentId.value) return currentId.value
    const conversation = await createNewConversation()
    return conversation.id
  }

  const updateConversationMeta = (conversation: ConversationSummary) => {
    upsertConversation(conversation)
  }

  const mutateTitle = async (conversationId: number, payload: UpdateConversationTitlePayload) => {
    const updated = await updateConversationTitle(conversationId, payload)
    updateConversationMeta(updated)
    emitToast({ type: 'success', message: '标题已更新。' })
  }

  const mutateCategory = async (
    conversationId: number,
    payload: UpdateConversationCategoryPayload,
  ) => {
    const updated = await updateConversationCategory(conversationId, payload)
    updateConversationMeta(updated)
    emitToast({ type: 'success', message: '分类已更新。' })
  }

  const requestAutoClassify = async (conversationId: number) => {
    const result = await triggerAutoClassify(conversationId)
    emitToast({ type: 'success', message: '已触发自动分类。' })
    await loadConversations()
    return result
  }

  const removeConversation = async (conversationId: number) => {
    await deleteConversation(conversationId)
    conversations.value = conversations.value.filter((item) => item.id !== conversationId)
    const { [conversationId]: _removed, ...rest } = messageMap.value
    messageMap.value = rest
    emitToast({ type: 'success', message: '对话已移动到回收站。' })
    if (currentId.value === conversationId) {
      currentId.value = conversations.value[0]?.id ?? null
    }
  }

  const abortStreaming = () => {
    if (streamingState.controller) {
      streamingState.controller.abort()
    }
    if (streamingState.pendingAssistantId && currentId.value) {
      updateMessage(currentId.value, streamingState.pendingAssistantId, (msg) => {
        msg.status = 'done'
      })
    }
    streamingState.controller = null
    streamingState.pendingAssistantId = null
    streamingState.isStreaming = false
  }

  const sendMessage = async (payload: SendMessagePayload) => {
    if (streamingState.isStreaming) {
      emitToast({ type: 'warning', message: '请等待当前回复完成或手动停止。' })
      return
    }

    const trimmed = payload.message.trim()
    if (trimmed.length === 0) {
      emitToast({ type: 'warning', message: '请输入内容后再发送。' })
      return
    }

    if (trimmed.length > 4000) {
      emitToast({ type: 'warning', message: '消息过长，请精简后重试。' })
      return
    }

    const conversationId = await ensureConversation()
    currentId.value = conversationId

    const userMessage = createUserMessage(conversationId, trimmed)
    appendMessage(conversationId, userMessage)

    const assistantMessage = createAssistantPlaceholder(conversationId)
    appendMessage(conversationId, assistantMessage)

    streamingState.pendingAssistantId = assistantMessage.id
    streamingState.isStreaming = true
    streamingState.enableThinking = Boolean(payload.enable_thinking)
    streamingState.modelId = payload.model_id

    try {
      const controller = await streamConversation({
        conversationId,
        payload: { ...payload, message: trimmed },
        onDelta(chunk) {
          if (!streamingState.pendingAssistantId) return
          updateMessage(conversationId, streamingState.pendingAssistantId, (msg) => {
            msg.content += chunk
          })
        },
        onDone() {
          if (!streamingState.pendingAssistantId) return
          updateMessage(conversationId, streamingState.pendingAssistantId, (msg) => {
            msg.status = 'done'
          })
          streamingState.controller = null
          streamingState.pendingAssistantId = null
          streamingState.isStreaming = false
          loadConversations()
        },
        onError(error) {
          console.error('[conversations] streaming error', error)
          if (streamingState.pendingAssistantId) {
            updateMessage(conversationId, streamingState.pendingAssistantId, (msg) => {
              msg.status = 'error'
            })
          }
          emitToast({ type: 'error', message: '生成失败，请重试。' })
          streamingState.controller = null
          streamingState.pendingAssistantId = null
          streamingState.isStreaming = false
        },
      })

      streamingState.controller = controller
    } catch (error) {
      console.error('[conversations] failed to send message', error)
      streamingState.controller = null
      streamingState.pendingAssistantId = null
      streamingState.isStreaming = false
      emitToast({ type: 'error', message: '消息发送失败。' })
    }
  }

  return {
    conversations,
    currentId,
    currentConversation,
    currentMessages,
    filteredConversations,
    listLoading,
    messagesLoading,
    searchQuery,
    streamingState,
    loadConversations,
    loadMessages,
    selectConversation,
    createNewConversation,
    mutateTitle,
    mutateCategory,
    requestAutoClassify,
    removeConversation,
    sendMessage,
    abortStreaming,
    ensureConversation,
  }
})
