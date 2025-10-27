<template>
  <div
    class="flex min-h-screen flex-col items-center justify-center gap-4 bg-surface-light px-6 text-center dark:bg-surface-dark"
  >
    <div class="space-y-2">
      <h1 class="text-4xl font-semibold text-slate-900 dark:text-slate-100">{{ title }}</h1>
      <p class="max-w-xl text-sm text-slate-500 dark:text-slate-400">
        <slot>
          {{ description }}
        </slot>
      </p>
    </div>
    <RouterLink
      class="rounded-full bg-brand px-6 py-2 text-sm font-medium text-white shadow-elevated transition hover:bg-brand-dark"
      :to="{ name: 'chat' }"
    >
      {{ t('common.backToChat') }}
    </RouterLink>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

withDefaults(
  defineProps<{
    title?: string
    description?: string
  }>(),
  {
    title: 'Oops',
    description: 'Something went wrong.',
  },
)

const { t } = useI18n()
</script>
