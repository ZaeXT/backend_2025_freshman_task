<template>
  <aside
    class="hidden w-72 flex-shrink-0 flex-col border-l border-white/10 bg-slate-950/60 px-4 py-6 text-sm text-slate-200 backdrop-blur dark:border-slate-800 lg:flex"
  >
    <div v-if="!conversation" class="mt-24 text-center text-xs text-slate-500">
      {{ t('chat.noConversationSelected') }}
    </div>
    <div v-else class="space-y-6">
      <section>
        <h2 class="text-xs font-semibold uppercase tracking-widest text-slate-400">
          {{ t('chat.metadata') }}
        </h2>
        <dl class="mt-3 space-y-2 text-xs text-slate-400">
          <div class="flex items-center justify-between">
            <dt>{{ t('chat.createdAt') }}</dt>
            <dd>{{ formatDateTime(conversation.created_at) }}</dd>
          </div>
          <div class="flex items-center justify-between">
            <dt>{{ t('chat.updatedAt') }}</dt>
            <dd>{{ formatDateTime(conversation.updated_at) }}</dd>
          </div>
          <div class="flex items-center justify-between">
            <dt>{{ t('chat.model') }}</dt>
            <dd>{{ conversation.model_id || t('chat.defaultModel') }}</dd>
          </div>
        </dl>
      </section>

      <section>
        <div class="flex items-center justify-between">
          <h2 class="text-xs font-semibold uppercase tracking-widest text-slate-400">
            {{ t('chat.category') }}
          </h2>
          <RouterLink
            class="text-xs text-brand transition hover:underline"
            :to="{ name: 'categories' }"
          >
            {{ t('chat.manageCategories') }}
          </RouterLink>
        </div>
        <select
          v-model="selectedCategory"
          class="mt-2 w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-xs text-slate-100 focus:border-brand focus:outline-none"
          @change="updateCategory"
        >
          <option value="">{{ t('chat.noCategory') }}</option>
          <option v-for="option in categoryOptions" :key="option.id" :value="option.id">
            {{ option.label }}
          </option>
        </select>
        <button
          class="mt-3 w-full rounded-full border border-brand/50 px-4 py-1.5 text-xs font-medium text-brand transition hover:bg-brand/10"
          type="button"
          @click="autoClassify"
        >
          {{ t('chat.autoClassify') }}
        </button>
      </section>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import { useCategoryStore, useConversationStore } from '@/stores'
import type { CategoryTreeNode } from '@/types'
import { formatDateTime } from '@/utils/time'

const conversationStore = useConversationStore()
const categoryStore = useCategoryStore()

const { t } = useI18n()

const { tree } = storeToRefs(categoryStore)
const conversation = computed(() => conversationStore.currentConversation)

const selectedCategory = ref('')

const categoryOptions = computed(() => {
  const result: { id: string; label: string }[] = []
  const traverse = (nodes: CategoryTreeNode[], depth = 0) => {
    nodes.forEach((node) => {
      result.push({ id: node.id.toString(), label: `${'— '.repeat(depth)}${node.name}` })
      if (node.children?.length) traverse(node.children, depth + 1)
    })
  }
  traverse(tree.value, 0)
  return result
})

watch(
  conversation,
  (value) => {
    selectedCategory.value = value?.category_id != null ? String(value.category_id) : ''
  },
  { immediate: true },
)

const updateCategory = async () => {
  if (!conversation.value) return
  const nextCategory = selectedCategory.value ? Number(selectedCategory.value) : null
  await conversationStore.mutateCategory(conversation.value.id, {
    category_id: nextCategory,
  })
  await categoryStore.loadCategories()
}

const autoClassify = async () => {
  if (!conversation.value) return
  const result = await conversationStore.requestAutoClassify(conversation.value.id)
  if (typeof result.category_id === 'number') {
    selectedCategory.value = String(result.category_id)
    await categoryStore.loadCategories()
  }
}
</script>
