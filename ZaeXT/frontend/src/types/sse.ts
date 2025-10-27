import type { MessageRole } from './conversation'

export interface OpenAIStreamChoiceDelta {
  role?: MessageRole
  content?: string
  function_call?: {
    name?: string
    arguments?: string
  }
}

export interface OpenAIStreamChoice {
  index?: number
  delta: OpenAIStreamChoiceDelta
  finish_reason?: string | null
}

export interface OpenAIStreamChunk {
  id?: string
  object?: string
  created?: number
  model?: string
  choices: OpenAIStreamChoice[]
}

export type SSEMessageHandler = (fragment: string, chunk: OpenAIStreamChunk) => void

export interface StreamCallbacks {
  onDelta: SSEMessageHandler
  onDone: () => void
  onError?: (error: Error) => void
}

export interface StreamController {
  abort: () => void
}
