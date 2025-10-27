<template>
  <div class="mx-auto max-w-4xl space-y-6 px-6 py-10">
    <header class="space-y-1">
      <button
        class="inline-flex items-center gap-2 rounded-full border border-slate-300/60 px-3 py-1 text-xs text-slate-500 transition hover:border-brand hover:text-brand dark:border-slate-700/60 dark:text-slate-300"
        type="button"
        @click="goBack"
      >
        ← {{ t('common.backToChat') }}
      </button>
      <h1 class="text-3xl font-semibold text-slate-900 dark:text-white">
        {{ t('recycleBin.title') }}
      </h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('recycleBin.subtitle') }}</p>
    </header>

    <section
      class="rounded-2xl border border-slate-200/80 bg-white/70 shadow-sm dark:border-slate-700/60 dark:bg-slate-900/60"
    >
      <AppLoading v-if="store.loading">{{ t('common.loading') }}</AppLoading>
      <AppEmptyState v-else-if="!store.items.length">
        <template #icon>🗃️</template>
        <template #title>{{ t('recycleBin.emptyTitle') }}</template>
        <template #description>{{ t('recycleBin.emptyBody') }}</template>
      </AppEmptyState>
      <div v-else class="divide-y divide-slate-200/70 dark:divide-slate-700/60">
        <article
          v-for="item in store.items"
          :key="item.id"
          class="flex flex-wrap items-center justify-between gap-3 px-6 py-4"
        >
          <div class="min-w-0 flex-1">
            <h2 class="truncate text-sm font-semibold text-slate-800 dark:text-slate-100">
              {{ item.title }}
            </h2>
            <p class="mt-1 text-xs text-slate-400">
              {{ t('recycleBin.deletedAt', { time: formatDateTime(item.deleted_at) }) }}
            </p>
            <p v-if="item.last_message_preview" class="mt-1 line-clamp-2 text-xs text-slate-400">
              {{ item.last_message_preview }}
            </p>
          </div>
          <div class="flex items-center gap-2 text-xs">
            <button
              class="rounded-full border border-emerald-400/60 px-3 py-1.5 text-emerald-500 transition hover:bg-emerald-500/10 disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="store.loading"
              @click="restore(item.id)"
            >
              {{ t('recycleBin.restore') }}
            </button>
            <button
              class="rounded-full border border-rose-400/60 px-3 py-1.5 text-rose-500 transition hover:bg-rose-500/10 disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="store.loading"
              @click="remove(item.id)"
            >
              {{ t('recycleBin.deleteForever') }}
            </button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import AppEmptyState from '@/components/AppEmptyState.vue'
import AppLoading from '@/components/AppLoading.vue'
import { useRecycleBinStore } from '@/stores'
import { formatDateTime } from '@/utils/time'

const store = useRecycleBinStore()
const { t } = useI18n()
const router = useRouter()

onMounted(() => {
  store.loadItems()
})

const restore = async (id: number) => {
  await store.restoreItem(id)
}

const remove = async (id: number) => {
  if (!window.confirm(t('recycleBin.confirmDelete'))) return
  await store.deletePermanently(id)
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.replace({ name: 'chat' }).catch(() => undefined)
}
</script>
