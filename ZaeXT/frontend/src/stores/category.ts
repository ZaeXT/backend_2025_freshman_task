import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { createCategory, deleteCategory, fetchCategories, updateCategory } from '@/api'
import type { CategoryTreeNode, CreateCategoryPayload, UpdateCategoryPayload } from '@/types'
import { emitToast } from '@/utils/notifications'

const flattenTree = (nodes: CategoryTreeNode[]): CategoryTreeNode[] => {
  const result: CategoryTreeNode[] = []
  const stack = [...nodes]
  while (stack.length) {
    const node = stack.shift()!
    result.push(node)
    if (node.children?.length) {
      stack.unshift(...node.children)
    }
  }
  return result
}

export const useCategoryStore = defineStore('categories', () => {
  const tree = ref<CategoryTreeNode[]>([])
  const loading = ref(false)
  const fetched = ref(false)

  const flatList = computed(() => flattenTree(tree.value))

  const loadCategories = async () => {
    if (loading.value) return
    try {
      loading.value = true
      tree.value = await fetchCategories()
      fetched.value = true
    } catch (error) {
      console.error('[categories] failed to fetch categories', error)
    } finally {
      loading.value = false
    }
  }

  const addCategory = async (payload: CreateCategoryPayload) => {
    const created = await createCategory(payload)
    emitToast({ type: 'success', message: '分类创建成功。' })
    await loadCategories()
    return created
  }

  const patchCategory = async (id: number, payload: UpdateCategoryPayload) => {
    await updateCategory(id, payload)
    emitToast({ type: 'success', message: '分类已更新。' })
    await loadCategories()
  }

  const removeCategory = async (id: number) => {
    await deleteCategory(id)
    emitToast({ type: 'success', message: '分类已删除。' })
    await loadCategories()
  }

  return {
    tree,
    flatList,
    loading,
    fetched,
    loadCategories,
    addCategory,
    patchCategory,
    removeCategory,
  }
})
