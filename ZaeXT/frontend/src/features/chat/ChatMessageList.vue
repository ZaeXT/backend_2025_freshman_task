<template>
  <section class="flex-1 min-h-0 overflow-hidden">
    <div class="flex h-full min-h-0 w-full flex-col">
      <div
        v-if="messages.length"
        ref="listEl"
        class="flex-1 min-h-0 overflow-y-auto bg-white/60 dark:bg-slate-900/50"
      >
        <ul class="divide-y divide-slate-200 dark:divide-slate-700">
          <li v-for="item in messages" :key="item.id">
            <article
              class="flex w-full flex-col gap-1 px-4 py-3 text-sm"
              :class="item.role === 'user' ? 'items-end text-right' : 'items-start text-left'"
            >
              <header
                class="flex w-full items-center text-[11px] font-medium text-slate-400 dark:text-slate-500"
                :class="item.role === 'user' ? 'justify-end gap-2' : 'justify-between'"
              >
                <span class="tracking-wide">
                  {{ item.role === 'user' ? t('chat.userLabel') : t('chat.assistantLabel') }}
                </span>
                <time class="font-mono text-[10px] text-slate-400 dark:text-slate-500">
                  {{ formatDateTime(item.created_at, 'HH:mm') }}
                </time>
              </header>
              <div
                class="leading-relaxed text-slate-900 dark:text-slate-100"
                :class="item.role === 'user' ? 'max-w-[70%] text-right' : 'max-w-[70%] text-left'"
              >
                <template v-if="item.role === 'assistant'">
                  <AppMarkdownRenderer :content="item.content" />
                  <div v-if="item.status === 'streaming'" class="mt-1 text-[11px] text-slate-400">
                    {{ t('chat.typing') }}
                  </div>
                  <div v-if="item.status === 'error'" class="mt-1 text-[11px] text-rose-400">
                    {{ t('chat.messageFailed') }}
                  </div>
                </template>
                <template v-else>
                  <p class="whitespace-pre-wrap text-slate-700 dark:text-slate-200">
                    {{ item.content }}
                  </p>
                </template>
              </div>
            </article>
          </li>
        </ul>
      </div>
      <AppEmptyState v-else>
        <template #icon>💬</template>
        <template #title>{{ t('chat.emptyConversationTitle') }}</template>
        <template #description>{{ t('chat.emptyConversationDescription') }}</template>
      </AppEmptyState>
    </div>
  </section>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppEmptyState from '@/components/AppEmptyState.vue'
import AppMarkdownRenderer from '@/components/AppMarkdownRenderer.vue'
import { useConversationStore } from '@/stores'
import { formatDateTime } from '@/utils/time'

const conversationStore = useConversationStore()
const { currentMessages: messages } = storeToRefs(conversationStore)
const { t } = useI18n()

const listEl = ref<HTMLDivElement | null>(null)

const scrollToBottom = async () => {
  await nextTick()
  const container = listEl.value
  if (!container) return
  container.scrollTop = container.scrollHeight
}

onMounted(() => {
  scrollToBottom()
})

watch(
  () => messages.value.length,
  () => {
    scrollToBottom()
  },
)

watch(
  () => messages.value.at(-1)?.content,
  () => {
    scrollToBottom()
  },
)
</script>
