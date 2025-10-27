import type {
  OpenAIStreamChunk,
  SendMessagePayload,
  StreamCallbacks,
  StreamController,
} from '@/types'
import { emitToast } from '@/utils/notifications'
import { getAuthHeader } from '@/utils/token'

interface SSEOptions extends StreamCallbacks {
  conversationId: number
  payload: SendMessagePayload
  retry?: boolean
}

const DONE = '[DONE]'
const HEARTBEAT_INTERVAL = 15_000

class SSEConnectionController implements StreamController {
  private abortController: AbortController

  constructor(abortController: AbortController) {
    this.abortController = abortController
  }

  abort() {
    this.abortController.abort()
  }
}

export const streamConversation = async ({
  conversationId,
  payload,
  onDelta,
  onDone,
  onError,
}: SSEOptions): Promise<StreamController> => {
  const abortController = new AbortController()
  const controller = new SSEConnectionController(abortController)

  const url = `${import.meta.env.VITE_API_BASE}/conversations/${conversationId}/messages`

  try {
    const headers = new Headers({
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
    })
    const authHeader = getAuthHeader()
    if (authHeader.Authorization) {
      headers.set('Authorization', authHeader.Authorization)
    }

    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(payload),
      signal: abortController.signal,
    })

    if (!response.ok || !response.body) {
      const error = new Error(`Failed to open stream: ${response.status} ${response.statusText}`)
      onError?.(error)
      throw error
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let buffer = ''
    let lastHeartbeat = Date.now()

    const processChunk = (chunk: string) => {
      const lines = chunk.split('\n')
      for (const line of lines) {
        if (!line.trim()) {
          continue
        }

        if (line.startsWith(':')) {
          // heartbeat line
          lastHeartbeat = Date.now()
          continue
        }

        if (!line.startsWith('data:')) {
          continue
        }

        const data = line.replace(/^data:\s*/, '')
        if (data === DONE) {
          onDone()
          return
        }

        try {
          const parsed = JSON.parse(data) as OpenAIStreamChunk
          const delta = parsed.choices?.[0]?.delta?.content
          if (delta) {
            onDelta(delta, parsed)
          }
          lastHeartbeat = Date.now()
        } catch (error) {
          console.error('Failed to parse SSE chunk', data, error)
        }
      }
    }

    const pump = async () => {
      while (!abortController.signal.aborted) {
        const { value, done } = await reader.read()
        if (done) {
          onDone()
          break
        }
        buffer += decoder.decode(value, { stream: true })
        const parts = buffer.split('\n\n')
        buffer = parts.pop() ?? ''
        parts.forEach(processChunk)

        if (Date.now() - lastHeartbeat > HEARTBEAT_INTERVAL) {
          const heartbeatError = new Error('SSE heartbeat timeout')
          emitToast({ type: 'warning', message: '连接中断，正在尝试恢复...' })
          onError?.(heartbeatError)
          break
        }
      }
    }

    pump().catch((error) => {
      if (error.name === 'AbortError') {
        return
      }
      console.error('[SSE] stream error', error)
      onError?.(error)
    })
  } catch (error) {
    if ((error as Error).name !== 'AbortError') {
      emitToast({ type: 'error', message: '消息发送失败，请稍后重试。' })
      onError?.(error as Error)
    }
  }

  return controller
}
