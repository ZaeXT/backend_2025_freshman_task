import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { getProfile, login, register, updateMemory } from '@/api'
import type { LoginPayload, RegisterPayload, UpdateMemoryPayload, UserProfile } from '@/types'
import { appEventBus } from '@/utils/event-bus'
import { emitToast } from '@/utils/notifications'
import { clearToken, getToken, setToken } from '@/utils/token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const profile = ref<UserProfile | null>(null)
  const loading = ref(false)
  const initialized = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value))
  const tier = computed(() => profile.value?.tier ?? 'free')

  const handleUnauthorized = () => {
    token.value = null
    profile.value = null
    initialized.value = false
  }

  appEventBus.on('auth:unauthorized', handleUnauthorized)

  const setAuthToken = (value: string | null) => {
    token.value = value
    if (value) {
      setToken(value)
    } else {
      clearToken()
    }
  }

  const fetchProfileSafely = async () => {
    if (!token.value || initialized.value) return
    try {
      loading.value = true
      profile.value = await getProfile()
    } catch (error) {
      console.error('[auth] failed to fetch profile', error)
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  const performLogin = async (payload: LoginPayload) => {
    loading.value = true
    try {
      const response = await login(payload)
      setAuthToken(response.token)
      await fetchProfileSafely()
      emitToast({ type: 'success', message: '登录成功。' })
    } finally {
      loading.value = false
    }
  }

  const performRegister = async (payload: RegisterPayload) => {
    loading.value = true
    try {
      await register(payload)
      emitToast({ type: 'success', message: '注册成功，请登录。' })
    } finally {
      loading.value = false
    }
  }

  const performLogout = () => {
    setAuthToken(null)
    profile.value = null
    initialized.value = false
    emitToast({ type: 'info', message: '您已退出登录。' })
  }

  const updateMemoryInfo = async (payload: UpdateMemoryPayload) => {
    if (!token.value) return
    await updateMemory(payload)
    if (profile.value) {
      profile.value = { ...profile.value, memory_info: payload.memory_info }
    }
    emitToast({ type: 'success', message: '记忆已更新。' })
  }

  return {
    token,
    profile,
    tier,
    loading,
    isAuthenticated,
    initialized,
    setAuthToken,
    fetchProfileSafely,
    performLogin,
    performRegister,
    performLogout,
    updateMemoryInfo,
  }
})
