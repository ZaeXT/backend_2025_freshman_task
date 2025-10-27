import axios, { AxiosError, type AxiosInstance } from 'axios'

import { router } from '@/router'
import type { ApiErrorPayload, ApiResp } from '@/types'
import { appEventBus } from '@/utils/event-bus'
import { emitToast } from '@/utils/notifications'
import { clearToken, getToken } from '@/utils/token'

const DEFAULT_ERROR_MESSAGE = 'Request failed, please try again later.'

export class ApiError extends Error {
  code: number
  payload?: unknown

  constructor(message: string, code: number, payload?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.payload = payload
  }
}

const createHttpClient = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE,
    timeout: 60000,
    withCredentials: false,
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
  })

  instance.interceptors.request.use((config) => {
    const token = getToken()
    if (token) {
      config.headers = config.headers ?? {}
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  instance.interceptors.response.use(
    (response) => {
      const payload = response.data as ApiResp<unknown>
      if (payload && typeof payload === 'object' && 'code' in payload) {
        if (typeof payload.code === 'number' && payload.code !== 200) {
          const error = new ApiError(
            payload.msg ?? DEFAULT_ERROR_MESSAGE,
            payload.code,
            payload.data,
          )
          emitToast({ type: 'error', message: error.message })
          return Promise.reject(error)
        }
        return (payload as ApiResp<unknown>).data
      }
      return response.data
    },
    async (error: AxiosError<ApiErrorPayload>) => {
      let message = DEFAULT_ERROR_MESSAGE
      const status = error.response?.status

      if (error.response?.data?.msg) {
        message = error.response.data.msg
      } else if (error.message) {
        message = error.message
      }

      const apiError = new ApiError(message, status ?? -1, error.response?.data)

      if (status === 401) {
        clearToken()
        appEventBus.emit('auth:unauthorized')
        emitToast({ type: 'error', message: '登录已失效，请重新登录。' })
        if (router.currentRoute.value.name !== 'login') {
          await router.replace({
            name: 'login',
            query: { redirect: router.currentRoute.value.fullPath },
          })
        }
      } else if (status === 403) {
        emitToast({ type: 'error', message: '您没有执行该操作的权限。' })
      } else if (status === 404) {
        emitToast({ type: 'error', message: '请求的资源不存在。' })
        if (router.currentRoute.value.name !== 'not-found') {
          await router.replace({ name: 'not-found' })
        }
      } else if (status && status >= 500) {
        emitToast({ type: 'error', message: '服务器开小差了，请稍后重试。' })
        if (router.currentRoute.value.name !== 'error') {
          await router
            .replace({
              name: 'error',
              params: { code: status },
            })
            .catch(() => {
              // fallback to chat if error route not registered yet
              router.replace({ name: 'chat' }).catch(() => undefined)
            })
        }
      } else {
        emitToast({ type: 'error', message })
      }

      return Promise.reject(apiError)
    },
  )

  return instance
}

export const http = createHttpClient()

export const get = http.get.bind(http)
export const post = http.post.bind(http)
export const put = http.put.bind(http)
export const del = http.delete.bind(http)
