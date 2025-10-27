import { ReadableStream } from 'node:stream/web'

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { streamConversation } from '../sse'

const encoder = new TextEncoder()

const createStream = (chunks: string[]) =>
  new ReadableStream<Uint8Array>({
    start(controller) {
      chunks.forEach((chunk) => controller.enqueue(encoder.encode(chunk)))
      controller.close()
    },
  })

describe('streamConversation', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('parses streaming content and resolves on DONE', async () => {
    const chunks = [
      'data: {"choices":[{"delta":{"content":"你"}}]}\n\n',
      'data: {"choices":[{"delta":{"content":"好"}}]}\n\n',
      'data: [DONE]\n\n',
    ]

    const body = createStream(chunks)
    const response = new Response(body as unknown as BodyInit, {
      status: 200,
      headers: { 'Content-Type': 'text/event-stream' },
    })

    const fetchMock = vi.fn(async () => response)
    vi.stubGlobal('fetch', fetchMock)

    const deltas: string[] = []
    let resolveDone!: () => void
    const donePromise = new Promise<void>((resolve) => {
      resolveDone = resolve
    })

    const onDone = vi.fn(() => resolveDone())
    const onDelta = vi.fn((fragment: string) => {
      deltas.push(fragment)
    })

    await streamConversation({
      conversationId: 1,
      payload: { message: 'hi' },
      onDelta,
      onDone,
    })

    await donePromise

    expect(fetchMock).toHaveBeenCalled()
    expect(onDone).toHaveBeenCalled()
    expect(onDelta).toHaveBeenCalledTimes(2)
    expect(deltas.join('')).toBe('你好')
  })

  it('invokes onError when fetch fails', async () => {
    const error = new Error('network failure')
    const fetchMock = vi.fn(async () => {
      throw error
    })
    vi.stubGlobal('fetch', fetchMock)

    const onError = vi.fn()

    await streamConversation({
      conversationId: 2,
      payload: { message: 'test' },
      onDelta: vi.fn(),
      onDone: vi.fn(),
      onError,
    })

    expect(onError).toHaveBeenCalledWith(error)
  })
})
