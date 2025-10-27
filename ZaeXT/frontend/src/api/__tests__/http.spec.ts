import AxiosMockAdapter from 'axios-mock-adapter'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/router', () => {
  const replace = vi.fn(() => Promise.resolve())
  return {
    router: {
      currentRoute: { value: { name: 'chat', fullPath: '/' } },
      replace,
    },
  }
})

vi.mock('@/utils/token', () => ({
  getToken: vi.fn(() => 'mock-token'),
  clearToken: vi.fn(),
  setToken: vi.fn(),
  TOKEN_KEY: 'ai-chat-token',
}))

vi.mock('@/utils/notifications', () => ({
  emitToast: vi.fn(),
}))

import { router } from '@/router'
import * as notifications from '@/utils/notifications'
import * as tokenUtils from '@/utils/token'

import { ApiError, http } from '../http'

const mock = new AxiosMockAdapter(http)

describe('http client', () => {
  beforeEach(() => {
    mock.reset()
    vi.clearAllMocks()
  })

  it('attaches authorization header when token exists', async () => {
    mock.onGet('/protected').reply((config) => {
      expect(config.headers?.Authorization).toBe('Bearer mock-token')
      return [200, { code: 200, msg: 'ok', data: { success: true } }]
    })

    const result = await http.get<{ success: boolean }, { success: boolean }>('/protected')
    expect(result.success).toBe(true)
  })

  it('throws ApiError when backend returns business error code', async () => {
    mock.onGet('/business-error').reply(200, { code: 400, msg: '业务异常', data: null })

    await expect(http.get('/business-error')).rejects.toBeInstanceOf(ApiError)
    expect(notifications.emitToast).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'error', message: '业务异常' }),
    )
  })

  it('handles 401 by clearing token and redirecting to login', async () => {
    mock.onGet('/401').reply(401, { msg: 'unauthorized' })

    await expect(http.get('/401')).rejects.toBeInstanceOf(ApiError)
    expect(tokenUtils.clearToken).toHaveBeenCalled()
    expect(router.replace).toHaveBeenCalledWith({ name: 'login', query: { redirect: '/' } })
  })

  it('handles 403 with toast notification', async () => {
    mock.onGet('/403').reply(403, { msg: 'forbidden' })

    await expect(http.get('/403')).rejects.toBeInstanceOf(ApiError)
    expect(notifications.emitToast).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'error', message: '您没有执行该操作的权限。' }),
    )
  })

  it('redirects to not-found for 404 status', async () => {
    mock.onGet('/404').reply(404, { msg: 'not found' })

    await expect(http.get('/404')).rejects.toBeInstanceOf(ApiError)
    expect(router.replace).toHaveBeenCalledWith({ name: 'not-found' })
  })

  it('redirects to error route for 500 status', async () => {
    mock.onGet('/500').reply(500, { msg: 'server error' })

    await expect(http.get('/500')).rejects.toBeInstanceOf(ApiError)
    expect(router.replace).toHaveBeenCalledWith({
      name: 'error',
      params: { code: 500 },
    })
  })
})
