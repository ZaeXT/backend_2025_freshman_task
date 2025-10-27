<template>
  <div class="mx-auto max-w-4xl space-y-8 px-6 py-10">
    <header class="space-y-2">
      <h1 class="text-3xl font-semibold text-slate-900 dark:text-white">{{ t('models.title') }}</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('models.subtitle') }}</p>
    </header>

    <section class="grid gap-4 md:grid-cols-2">
      <div
        v-for="model in availableModels"
        :key="model.id"
        class="rounded-2xl border border-emerald-300/40 bg-emerald-100/40 p-6 shadow-sm dark:border-emerald-400/30 dark:bg-emerald-950/40"
      >
        <h2 class="text-lg font-semibold text-emerald-800 dark:text-emerald-200">
          {{ model.name }}
        </h2>
        <p class="mt-2 text-xs text-emerald-700/80 dark:text-emerald-200/70">
          {{ model.description || t('models.availableHint') }}
        </p>
        <span
          class="mt-4 inline-flex rounded-full bg-emerald-500/20 px-3 py-1 text-xs text-emerald-700 dark:text-emerald-200"
        >
          {{ t('models.accessible') }}
        </span>
      </div>
    </section>

    <section v-if="lockedModels.length" class="space-y-3">
      <h2 class="text-sm font-semibold text-slate-600 dark:text-slate-300">
        {{ t('models.lockedTitle') }}
      </h2>
      <div class="grid gap-4 md:grid-cols-2">
        <div
          v-for="model in lockedModels"
          :key="model.id"
          class="rounded-2xl border border-slate-300/60 bg-white/70 p-6 opacity-60 dark:border-slate-700/50 dark:bg-slate-900/50"
        >
          <h2 class="text-lg font-semibold text-slate-700 dark:text-slate-200">{{ model.name }}</h2>
          <p class="mt-2 text-xs text-slate-500 dark:text-slate-400">
            {{ model.description || t('models.lockedHint') }}
          </p>
          <span
            class="mt-4 inline-flex rounded-full bg-slate-500/10 px-3 py-1 text-xs text-slate-500 dark:text-slate-300"
          >
            {{ t('models.higherTierRequired') }}
          </span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

import { useModelStore } from '@/stores'

const modelStore = useModelStore()
const { availableModels, lockedModels } = storeToRefs(modelStore)
const { t } = useI18n()

onMounted(() => {
  modelStore.loadModels()
})
</script>
