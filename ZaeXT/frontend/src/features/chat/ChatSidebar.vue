<template>
  <aside
    class="flex h-full flex-col border-r border-white/10 bg-slate-950/60 text-slate-200 backdrop-blur dark:border-slate-800"
  >
    <div class="flex items-center gap-2 px-4 py-4">
      <button
        type="button"
        class="flex flex-1 items-center justify-center gap-2 rounded-full bg-brand px-3 py-2 text-sm font-medium text-white shadow-elevated transition hover:bg-brand-dark"
        @click="handleCreateConversation"
      >
        <span class="text-lg leading-none">＋</span>
        {{ t('chat.newConversation') }}
      </button>
      <RouterLink
        class="hidden rounded-full border border-slate-700/70 px-3 py-2 text-xs text-slate-300 transition hover:border-rose-400 hover:text-rose-200 lg:inline-flex"
        :to="{ name: 'recycle-bin' }"
      >
        {{ t('chat.recycleBinShortcut') }}
      </RouterLink>
    </div>
    <RouterLink
      class="mx-4 flex items-center justify-center gap-2 rounded-xl border border-slate-700/70 bg-slate-900/60 px-3 py-2 text-xs text-slate-300 transition hover:border-rose-400 hover:text-rose-200 lg:hidden"
      :to="{ name: 'recycle-bin' }"
    >
      <span>🗃️</span>
      <span>{{ t('chat.recycleBinMobile') }}</span>
    </RouterLink>
    <div class="px-4 pb-3">
      <label class="relative block">
        <span class="absolute inset-y-0 left-3 flex items-center text-xs text-slate-400">🔍</span>
        <input
          v-model="conversationStore.searchQuery"
          :placeholder="t('chat.searchPlaceholder')"
          class="w-full rounded-full border border-slate-700 bg-slate-900/80 py-2 pl-9 pr-3 text-sm text-slate-200 placeholder:text-slate-500 focus:border-brand focus:outline-none"
          type="search"
        />
      </label>
    </div>
    <div class="flex-1 overflow-y-auto px-2 pb-4">
      <AppLoading v-if="conversationStore.listLoading" />
      <AppEmptyState v-else-if="!conversationStore.filteredConversations.length">
        <template #title>{{ t('chat.noConversationsTitle') }}</template>
        <template #description>{{ t('chat.noConversationsHint') }}</template>
        <template #action>
          <button
            class="mt-4 rounded-full border border-brand/50 px-4 py-1.5 text-sm text-brand transition hover:bg-brand/10"
            type="button"
            @click="handleCreateConversation"
          >
            {{ t('chat.newConversation') }}
          </button>
        </template>
      </AppEmptyState>
      <ul v-else class="space-y-1">
        <li v-for="item in conversationStore.filteredConversations" :key="item.id">
          <button
            class="w-full rounded-xl px-3 py-2 text-left transition"
            :class="[
              item.id === conversationStore.currentId
                ? 'bg-brand/10 text-white shadow-inner'
                : 'bg-slate-900/40 text-slate-300 hover:bg-slate-900/70',
            ]"
            type="button"
            @click="conversationStore.selectConversation(item.id)"
          >
            <div class="flex items-center justify-between text-xs text-slate-400">
              <span class="truncate font-medium text-slate-200">{{ item.title }}</span>
              <span>{{ formatRelativeTime(item.updated_at) }}</span>
            </div>
          </button>
        </li>
      </ul>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import AppEmptyState from '@/components/AppEmptyState.vue'
import AppLoading from '@/components/AppLoading.vue'
import { useConversationStore } from '@/stores'
import { formatRelativeTime } from '@/utils/time'

const conversationStore = useConversationStore()
const { t } = useI18n()

const handleCreateConversation = async () => {
  await conversationStore.createNewConversation()
}
</script>
