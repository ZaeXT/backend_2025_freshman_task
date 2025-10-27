<template>
  <div
    class="pointer-events-none fixed inset-x-4 top-4 z-[1000] mx-auto flex max-w-md flex-col gap-3 sm:right-6 sm:left-auto"
  >
    <transition-group name="toast" tag="div">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="pointer-events-auto overflow-hidden rounded-2xl border border-white/20 bg-slate-900/90 shadow-xl backdrop-blur transition dark:border-slate-700"
      >
        <div class="flex items-start gap-3 px-4 py-3">
          <span class="mt-1 h-2 w-2 rounded-full" :class="indicatorClass(toast.type)"></span>
          <div class="flex-1 text-sm text-slate-100">
            <p v-if="toast.title" class="font-semibold">{{ toast.title }}</p>
            <p class="leading-relaxed text-slate-200">{{ toast.message }}</p>
          </div>
          <button
            type="button"
            class="text-slate-400 transition hover:text-white"
            @click="dismiss(toast.id)"
          >
            ×
          </button>
        </div>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive } from 'vue'

import type { ToastPayload } from '@/utils/notifications'
import { onToast } from '@/utils/notifications'

interface InternalToast extends ToastPayload {
  id: string
  timeout?: number
}

const toasts = reactive<InternalToast[]>([])

const indicatorClass = (type: ToastPayload['type']) => {
  switch (type) {
    case 'success':
      return 'bg-emerald-400'
    case 'error':
      return 'bg-rose-400'
    case 'warning':
      return 'bg-amber-400'
    default:
      return 'bg-sky-400'
  }
}

const pushToast = (payload: ToastPayload) => {
  const id =
    payload.id ??
    (typeof crypto !== 'undefined' && crypto.randomUUID
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random()}`)
  const toast: InternalToast = {
    id,
    ...payload,
    type: payload.type ?? 'info',
  }
  toasts.push(toast)

  if (payload.duration !== 0) {
    const duration = payload.duration ?? 3000
    toast.timeout = window.setTimeout(() => dismiss(id), duration)
  }
}

const dismiss = (id?: string) => {
  if (!id) return
  const index = toasts.findIndex((toast) => toast.id === id)
  if (index >= 0) {
    const [toast] = toasts.splice(index, 1)
    if (toast?.timeout) {
      window.clearTimeout(toast.timeout)
    }
  }
}

let unsubscribe: (() => void) | null = null

onMounted(() => {
  unsubscribe = onToast(pushToast)
})

onBeforeUnmount(() => {
  unsubscribe?.()
})
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-12px) scale(0.98);
}
</style>
