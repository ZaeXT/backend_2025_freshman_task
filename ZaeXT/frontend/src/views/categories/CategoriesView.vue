<template>
  <div class="mx-auto max-w-4xl space-y-6 px-6 py-10">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div class="space-y-2">
        <button
          class="inline-flex items-center gap-2 rounded-full border border-slate-300/60 px-3 py-1 text-xs text-slate-500 transition hover:border-brand hover:text-brand dark:border-slate-700/60 dark:text-slate-300"
          type="button"
          @click="goBack"
        >
          ← {{ t('common.backToChat') }}
        </button>
        <h1 class="text-3xl font-semibold text-slate-900 dark:text-white">
          {{ t('categories.title') }}
        </h1>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
          {{ t('categories.subtitle') }}
        </p>
      </div>
      <button
        class="rounded-full bg-brand px-4 py-1.5 text-sm font-medium text-white shadow transition hover:bg-brand-dark"
        type="button"
        @click="createRootCategory"
      >
        {{ t('categories.createRoot') }}
      </button>
    </header>

    <section
      v-if="categoryStore.loading"
      class="rounded-2xl border border-slate-200/70 bg-white/70 p-6 dark:border-slate-700/60 dark:bg-slate-900/60"
    >
      <AppLoading>{{ t('common.loading') }}</AppLoading>
    </section>

    <section
      v-else
      class="rounded-2xl border border-slate-200/80 bg-white/70 p-6 shadow-sm dark:border-slate-700/60 dark:bg-slate-900/60"
    >
      <AppEmptyState v-if="!categoryStore.tree.length">
        <template #title>{{ t('categories.emptyTitle') }}</template>
        <template #description>{{ t('categories.emptyBody') }}</template>
      </AppEmptyState>
      <ul v-else class="space-y-2">
        <CategoryNode v-for="node in categoryStore.tree" :key="node.id" :node="node" />
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import AppEmptyState from '@/components/AppEmptyState.vue'
import AppLoading from '@/components/AppLoading.vue'
import CategoryNode from '@/features/categories/CategoryNode.vue'
import { useCategoryStore } from '@/stores'

const categoryStore = useCategoryStore()
const { t } = useI18n()
const router = useRouter()

const createRootCategory = async () => {
  const name = window.prompt(t('categories.promptName'))
  if (!name) return
  await categoryStore.addCategory({ name })
}

onMounted(() => {
  categoryStore.loadCategories()
})

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.replace({ name: 'chat' }).catch(() => undefined)
}
</script>
