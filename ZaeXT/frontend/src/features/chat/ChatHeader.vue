<template>
  <header
    class="flex flex-shrink-0 flex-wrap items-center justify-between gap-3 border-b border-white/10 px-6 py-4 dark:border-slate-800"
  >
    <div class="flex min-w-0 flex-1 items-center gap-3">
      <input
        v-if="isEditing"
        ref="titleInput"
        v-model="draftTitle"
        class="min-w-0 flex-1 rounded-lg border border-slate-300 bg-white/80 px-3 py-2 text-base font-semibold text-slate-900 focus:border-brand focus:outline-none dark:border-slate-700 dark:bg-slate-900/70 dark:text-white"
        :placeholder="t('chat.titlePlaceholder')"
        @keyup.enter="submitTitle"
        @keyup.esc="cancelEdit"
      />
      <h1
        v-else
        class="cursor-text text-lg font-semibold text-slate-900 transition hover:text-brand dark:text-white"
        @click="startEdit"
      >
        {{ activeConversation?.title || t('chat.titlePlaceholder') }}
      </h1>
      <button
        v-if="!isEditing"
        type="button"
        class="rounded-full border border-slate-300/50 px-3 py-1 text-xs text-slate-500 transition hover:border-brand hover:text-brand dark:border-slate-600 dark:text-slate-300"
        @click="startEdit"
      >
        {{ t('chat.rename') }}
      </button>
    </div>
    <div class="flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="rounded-full border border-slate-300/50 px-3 py-1 text-xs text-slate-500 transition hover:border-brand hover:text-brand dark:border-slate-600 dark:text-slate-300"
        @click="regenerateTitle"
      >
        {{ t('chat.regenerateTitle') }}
      </button>
      <button
        type="button"
        class="rounded-full border border-slate-300/50 px-3 py-1 text-xs text-slate-500 transition hover:border-rose-400 hover:text-rose-500 dark:border-slate-600 dark:text-slate-300"
        @click="moveToRecycleBin"
      >
        {{ t('chat.moveToRecycleBin') }}
      </button>
      <span class="hidden h-5 w-px bg-slate-300/60 dark:bg-slate-600 md:inline-block" />
      <RouterLink
        class="rounded-full border border-slate-300/50 px-3 py-1 text-xs text-slate-500 transition hover:border-brand hover:text-brand dark:border-slate-600 dark:text-slate-300"
        :to="{ name: 'profile' }"
      >
        {{ profileName }}
      </RouterLink>
      <button
        type="button"
        class="rounded-full border border-slate-300/60 px-3 py-1 text-xs text-slate-500 transition hover:border-rose-500 hover:text-rose-500 dark:border-slate-600 dark:text-slate-300"
        @click="handleLogout"
      >
        {{ t('profile.logout') }}
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRouter } from 'vue-router'

import { useAuthStore, useConversationStore } from '@/stores'

const conversationStore = useConversationStore()
const authStore = useAuthStore()
const { t } = useI18n()
const router = useRouter()

const { profile } = storeToRefs(authStore)

const isEditing = ref(false)
const draftTitle = ref('')
const titleInput = ref<HTMLInputElement | null>(null)

const activeConversation = computed(() => conversationStore.currentConversation)
const profileName = computed(() => profile.value?.username ?? t('profile.viewProfileShort'))

const startEdit = () => {
  if (!activeConversation.value) return
  draftTitle.value = activeConversation.value.title
  isEditing.value = true
  nextTick(() => titleInput.value?.focus())
}

const cancelEdit = () => {
  isEditing.value = false
}

const submitTitle = async () => {
  if (!activeConversation.value) return
  await conversationStore.mutateTitle(activeConversation.value.id, {
    title: draftTitle.value.trim(),
  })
  isEditing.value = false
}

const regenerateTitle = async () => {
  if (!activeConversation.value) return
  await conversationStore.mutateTitle(activeConversation.value.id, { title: '' })
}

const moveToRecycleBin = async () => {
  if (!activeConversation.value) return
  if (!window.confirm(t('chat.confirmMoveToRecycleBin'))) {
    return
  }
  await conversationStore.removeConversation(activeConversation.value.id)
}

const handleLogout = () => {
  authStore.performLogout()
  router.replace({ name: 'login' }).catch(() => undefined)
}

watch(
  () => conversationStore.currentId,
  () => {
    isEditing.value = false
  },
)
</script>
