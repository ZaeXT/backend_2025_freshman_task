<template>
  <div class="mx-auto max-w-3xl space-y-6 px-6 py-10">
    <header class="space-y-1">
      <button
        class="inline-flex items-center gap-2 rounded-full border border-slate-300/60 px-3 py-1 text-xs text-slate-500 transition hover:border-brand hover:text-brand dark:border-slate-700/60 dark:text-slate-300"
        type="button"
        @click="goBack"
      >
        ← {{ t('common.backToChat') }}
      </button>
      <h1 class="text-3xl font-semibold text-slate-900 dark:text-white">
        {{ t('profile.title') }}
      </h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('profile.subtitle') }}</p>
    </header>

    <section
      class="rounded-2xl border border-slate-200/80 bg-white/80 p-6 shadow-sm dark:border-slate-700/60 dark:bg-slate-900/60"
    >
      <div v-if="!profile" class="py-12 text-center text-sm text-slate-400">
        {{ t('profile.loadingProfile') }}
      </div>
      <template v-else>
        <dl class="grid gap-4 text-sm text-slate-600 dark:text-slate-300 sm:grid-cols-2">
          <div>
            <dt class="text-xs uppercase tracking-widest text-slate-400">
              {{ t('profile.username') }}
            </dt>
            <dd class="mt-1 font-medium text-slate-900 dark:text-white">{{ profile.username }}</dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-widest text-slate-400">
              {{ t('profile.tier') }}
            </dt>
            <dd
              class="mt-1 inline-flex rounded-full bg-brand/10 px-3 py-1 text-xs font-medium text-brand"
            >
              {{ profile.tier }}
            </dd>
          </div>
          <div>
            <dt class="text-xs uppercase tracking-widest text-slate-400">
              {{ t('profile.createdAt') }}
            </dt>
            <dd class="mt-1">{{ formatDateTime(profile.created_at) }}</dd>
          </div>
        </dl>

        <div class="mt-8 space-y-3">
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">
            {{ t('profile.memoryLabel') }}
          </label>
          <textarea
            v-model="memoryDraft"
            class="h-40 w-full rounded-2xl border border-slate-300 bg-white/80 px-4 py-3 text-sm text-slate-700 shadow-inner focus:border-brand focus:outline-none dark:border-slate-700 dark:bg-slate-900/80 dark:text-slate-100"
            :placeholder="t('profile.memoryPlaceholder')"
            :disabled="saving"
          />
          <div class="flex justify-end gap-2 text-xs">
            <button
              class="rounded-full border border-slate-300 px-4 py-1.5 text-slate-500 transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="!hasChanges || saving"
              @click="resetDraft"
            >
              {{ t('profile.reset') }}
            </button>
            <button
              class="rounded-full bg-brand px-4 py-1.5 text-white shadow transition hover:bg-brand-dark disabled:cursor-not-allowed disabled:bg-brand/60"
              type="button"
              :disabled="!hasChanges || saving"
              @click="saveMemory"
            >
              <span v-if="saving">{{ t('profile.saving') }}</span>
              <span v-else>{{ t('profile.save') }}</span>
            </button>
          </div>
        </div>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores'
import { formatDateTime } from '@/utils/time'

const authStore = useAuthStore()
const router = useRouter()
const { profile } = storeToRefs(authStore)
const { t } = useI18n()

const memoryDraft = ref('')
const saving = ref(false)

onMounted(() => {
  authStore.fetchProfileSafely()
})

const resetDraft = () => {
  memoryDraft.value = profile.value?.memory_info ?? ''
}

const hasChanges = computed(() => memoryDraft.value !== (profile.value?.memory_info ?? ''))

watch(
  profile,
  () => {
    resetDraft()
  },
  { immediate: true },
)

const saveMemory = async () => {
  if (saving.value || !hasChanges.value) return
  saving.value = true
  try {
    await authStore.updateMemoryInfo({ memory_info: memoryDraft.value })
  } finally {
    saving.value = false
  }
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.replace({ name: 'chat' }).catch(() => undefined)
}
</script>
