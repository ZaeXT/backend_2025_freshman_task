import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchModels } from '@/api'
import type { ModelInfo } from '@/types'

import { useAuthStore } from './auth'

const tierOrder = ['free', 'basic', 'pro', 'enterprise']

const getTierRank = (tier: string | undefined) => {
  if (!tier) return 0
  const index = tierOrder.indexOf(tier)
  return index === -1 ? tierOrder.length : index
}

export const useModelStore = defineStore('models', () => {
  const models = ref<ModelInfo[]>([])
  const loading = ref(false)
  const fetched = ref(false)

  const authStore = useAuthStore()

  const availableModels = computed(() => {
    if (!authStore.profile) return models.value
    const userRank = getTierRank(authStore.profile.tier)
    return models.value.filter((model) => getTierRank(model.tier) <= userRank)
  })

  const lockedModels = computed(() => {
    if (!authStore.profile) return []
    const userRank = getTierRank(authStore.profile.tier)
    return models.value.filter((model) => getTierRank(model.tier) > userRank)
  })

  const loadModels = async () => {
    if (loading.value || (fetched.value && models.value.length > 0)) return
    try {
      loading.value = true
      models.value = await fetchModels()
      fetched.value = true
    } catch (error) {
      console.error('[models] failed to fetch models', error)
    } finally {
      loading.value = false
    }
  }

  return {
    models,
    loading,
    availableModels,
    lockedModels,
    loadModels,
  }
})
