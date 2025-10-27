import { createI18n } from 'vue-i18n'

import enUS from './locales/en-US.json'
import zhCN from './locales/zh-CN.json'

export type AppMessageSchema = typeof zhCN

const STORAGE_KEY = 'ai-chat-lang'

const locale =
  (typeof window !== 'undefined' && window.localStorage.getItem(STORAGE_KEY)) || 'zh-CN'

export const i18n = createI18n<[AppMessageSchema], 'zh-CN' | 'en-US'>({
  legacy: false,
  globalInjection: true,
  locale: locale === 'en-US' ? 'en-US' : 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

export const persistLocale = (value: 'zh-CN' | 'en-US') => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(STORAGE_KEY, value)
  const localeRef = i18n.global.locale as unknown as { value: 'zh-CN' | 'en-US' }
  localeRef.value = value
}
