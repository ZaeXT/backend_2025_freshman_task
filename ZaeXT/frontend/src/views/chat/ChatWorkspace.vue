<template>
  <div class="flex h-screen flex-col overflow-hidden bg-slate-100 dark:bg-slate-950">
    <main class="flex flex-1 min-h-0 overflow-hidden">
      <ChatSidebar class="hidden w-72 flex-shrink-0 lg:flex" />

      <div class="relative flex flex-1 min-h-0 flex-col overflow-hidden">
        <div
          class="absolute left-4 right-4 z-10 flex flex-col gap-3 lg:hidden"
          style="top: calc(env(safe-area-inset-top, 0px) + 4.5rem)"
        >
          <button
            class="inline-flex h-9 w-auto self-start items-center rounded-full border border-slate-300 bg-white/80 px-3 text-xs text-slate-600 shadow"
            type="button"
            @click="sidebarOpen = true"
          >
            {{ t('chat.openSidebar') }}
          </button>
        </div>
        <ChatHeader />
        <ChatMessageList />
        <ChatComposer />
      </div>

      <ChatInfoPanel class="hidden w-72 flex-shrink-0 xl:flex" />
    </main>

    <TransitionRoot :show="sidebarOpen">
      <Dialog as="div" class="relative z-50 lg:hidden" @close="sidebarOpen = false">
        <TransitionChild
          enter="duration-200 ease-out"
          enter-from="opacity-0"
          enter-to="opacity-100"
          leave="duration-150 ease-in"
          leave-from="opacity-100"
          leave-to="opacity-0"
        >
          <div class="fixed inset-0 bg-black/50" />
        </TransitionChild>

        <div class="fixed inset-0 flex">
          <TransitionChild
            enter="duration-200 ease-out"
            enter-from="-translate-x-full"
            enter-to="translate-x-0"
            leave="duration-150 ease-in"
            leave-from="translate-x-0"
            leave-to="-translate-x-full"
          >
            <DialogPanel class="relative h-full w-72 bg-slate-950">
              <ChatSidebar />
            </DialogPanel>
          </TransitionChild>
        </div>
      </Dialog>
    </TransitionRoot>
  </div>
</template>

<script setup lang="ts">
import { Dialog, DialogPanel, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import ChatComposer from '@/features/chat/ChatComposer.vue'
import ChatHeader from '@/features/chat/ChatHeader.vue'
import ChatInfoPanel from '@/features/chat/ChatInfoPanel.vue'
import ChatMessageList from '@/features/chat/ChatMessageList.vue'
import ChatSidebar from '@/features/chat/ChatSidebar.vue'
import { useAuthStore, useCategoryStore, useConversationStore, useModelStore } from '@/stores'

const authStore = useAuthStore()
const conversationStore = useConversationStore()
const modelStore = useModelStore()
const categoryStore = useCategoryStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const sidebarOpen = ref(false)

const ensureConversationSelected = async () => {
  const idParam = route.params.id
  const id =
    typeof idParam === 'string'
      ? Number(idParam)
      : Array.isArray(idParam)
        ? Number(idParam[0])
        : null
  if (id) {
    await conversationStore.selectConversation(id)
  }
}

onMounted(async () => {
  await authStore.fetchProfileSafely()
  await Promise.all([
    conversationStore.loadConversations(),
    modelStore.loadModels(),
    categoryStore.loadCategories(),
  ])
  await ensureConversationSelected()
})

watch(
  () => route.params.id,
  async (newId) => {
    if (!newId) return
    const parsed = Array.isArray(newId) ? Number(newId[0]) : Number(newId)
    if (!Number.isNaN(parsed)) {
      await conversationStore.selectConversation(parsed)
    }
  },
)

watch(
  () => conversationStore.currentId,
  (id) => {
    if (!id) return
    if (route.params.id?.toString() === id.toString()) return
    router.replace({ name: 'chat-with-id', params: { id } }).catch(() => undefined)
    sidebarOpen.value = false
  },
)
</script>
