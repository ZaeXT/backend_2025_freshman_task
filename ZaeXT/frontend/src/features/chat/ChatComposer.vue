<template>
  <footer
    class="flex-shrink-0 border-t border-white/10 bg-slate-950/70 px-6 py-4 backdrop-blur dark:border-slate-800"
  >
    <form class="mx-auto flex max-w-3xl flex-col gap-3" @submit.prevent="handleSubmit">
      <textarea
        v-model="message"
        :placeholder="t('chat.inputPlaceholder')"
        class="h-28 w-full resize-none rounded-2xl border border-slate-700 bg-slate-900/70 px-4 py-3 text-sm text-slate-100 shadow-inner focus:border-brand focus:outline-none"
        :disabled="isStreaming"
      />
      <div class="flex flex-wrap items-center justify-between gap-3 text-xs text-slate-400">
        <div class="flex flex-wrap items-center gap-3">
          <label class="flex items-center gap-2">
            <span>{{ t('chat.modelLabel') }}</span>
            <select
              v-model="selectedModel"
              class="rounded-full border border-slate-600 bg-slate-900 px-3 py-1.5 text-xs text-slate-200 focus:border-brand focus:outline-none"
            >
              <option v-for="model in availableModels" :key="model.id" :value="model.id">
                {{ model.name }}
              </option>
              <option v-if="lockedModels.length" disabled>──────────</option>
              <option
                v-for="model in lockedModels"
                :key="`locked-${model.id}`"
                :value="model.id"
                disabled
              >
                {{ model.name }} · {{ t('chat.upgradeRequired') }}
              </option>
            </select>
          </label>
          <label class="flex cursor-pointer items-center gap-2">
            <input v-model="enableThinking" class="accent-brand" type="checkbox" />
            <span>{{ t('chat.enableThinking') }}</span>
          </label>
        </div>
        <div class="flex items-center gap-2">
          <button
            v-if="isStreaming"
            type="button"
            class="rounded-full border border-rose-500/70 px-4 py-1.5 text-xs font-medium text-rose-400 transition hover:bg-rose-500/10"
            @click="conversationStore.abortStreaming"
          >
            {{ t('chat.stopGenerating') }}
          </button>
          <button
            v-else
            type="submit"
            class="rounded-full bg-brand px-4 py-1.5 text-xs font-medium text-white shadow-elevated transition hover:bg-brand-dark disabled:cursor-not-allowed disabled:bg-brand/40"
            :disabled="!message.trim()"
          >
            {{ t('chat.send') }}
          </button>
        </div>
      </div>
    </form>
  </footer>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { useConversationStore, useModelStore } from '@/stores'
import { emitToast } from '@/utils/notifications'

const conversationStore = useConversationStore()
const modelStore = useModelStore()
const { t } = useI18n()

const message = ref('')
const enableThinking = ref(false)
const selectedModel = ref<string | undefined>(undefined)

const { streamingState } = storeToRefs(conversationStore)
const isStreaming = computed(() => streamingState.value.isStreaming)

const { availableModels, lockedModels } = storeToRefs(modelStore)

watch(
  () => availableModels.value,
  (models) => {
    if (!models.length) return
    if (!selectedModel.value) {
      selectedModel.value = models[0]?.id
    }
  },
  { immediate: true },
)

const handleSubmit = async () => {
  const original = message.value
  const trimmed = original.trim()
  if (!trimmed) return
  if (
    selectedModel.value &&
    !availableModels.value.some((model) => model.id === selectedModel.value)
  ) {
    emitToast({ type: 'warning', message: t('chat.modelNotAccessible') })
    return
  }
  message.value = ''
  try {
    await conversationStore.sendMessage({
      message: trimmed,
      enable_thinking: enableThinking.value,
      model_id: selectedModel.value,
    })
  } catch (error) {
    message.value = original
    console.error('[chat] send message failed', error)
  }
}
</script>
