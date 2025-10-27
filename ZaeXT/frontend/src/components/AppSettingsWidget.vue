<template>
  <div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2">
    <button
      class="rounded-full border border-white/20 bg-slate-900/80 px-3 py-1.5 text-xs text-slate-100 shadow hover:bg-slate-800/80"
      type="button"
      @click="toggleTheme"
    >
      {{ themeLabel }}
    </button>
    <button
      class="rounded-full border border-white/20 bg-slate-900/80 px-3 py-1.5 text-xs text-slate-100 shadow hover:bg-slate-800/80"
      type="button"
      @click="switchLocale"
    >
      {{ localeLabel }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { persistLocale } from '@/i18n'

const { locale, t } = useI18n()

const getInitialTheme = (): 'light' | 'dark' => {
  if (typeof window === 'undefined') return 'dark'
  return (window.localStorage.getItem('ai-chat-theme') as 'light' | 'dark') ?? 'dark'
}

const theme = ref<'light' | 'dark'>(getInitialTheme())

const applyTheme = () => {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  if (theme.value === 'dark') {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }
  window.localStorage.setItem('ai-chat-theme', theme.value)
}

onMounted(() => {
  applyTheme()
})

const toggleTheme = () => {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  applyTheme()
}

const switchLocale = () => {
  const next = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  persistLocale(next)
}

const themeLabel = computed(() =>
  theme.value === 'dark' ? t('settings.switchLight') : t('settings.switchDark'),
)

const localeLabel = computed(() =>
  locale.value === 'zh-CN' ? t('settings.switchEnglish') : t('settings.switchChinese'),
)
</script>
