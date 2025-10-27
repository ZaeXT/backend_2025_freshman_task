import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { deleteConversationPermanently, fetchRecycleBinItems, restoreConversation } from '@/api'
import type { RecycleBinItem } from '@/types'
import { emitToast } from '@/utils/notifications'

export const useRecycleBinStore = defineStore('recycle-bin', () => {
  const items = ref<RecycleBinItem[]>([])
  const loading = ref(false)

  const hasItems = computed(() => items.value.length > 0)

  const loadItems = async () => {
    if (loading.value) return
    try {
      loading.value = true
      items.value = await fetchRecycleBinItems()
    } catch (error) {
      console.error('[recycle-bin] failed to fetch items', error)
    } finally {
      loading.value = false
    }
  }

  const restoreItem = async (id: number) => {
    await restoreConversation(id)
    emitToast({ type: 'success', message: '对话已恢复。' })
    await loadItems()
  }

  const deletePermanently = async (id: number) => {
    await deleteConversationPermanently(id)
    emitToast({ type: 'success', message: '对话已永久删除。' })
    await loadItems()
  }

  return {
    items,
    loading,
    hasItems,
    loadItems,
    restoreItem,
    deletePermanently,
  }
})
