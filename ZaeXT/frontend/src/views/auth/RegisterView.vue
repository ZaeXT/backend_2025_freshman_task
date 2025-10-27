<template>
  <div
    class="grid min-h-screen place-items-center bg-gradient-to-br from-slate-900 via-slate-950 to-black px-4 py-12"
  >
    <div
      class="w-full max-w-md rounded-3xl border border-white/10 bg-white/5 p-8 shadow-2xl backdrop-blur"
    >
      <h1 class="text-2xl font-semibold text-white">{{ t('auth.registerTitle') }}</h1>
      <p class="mt-2 text-sm text-slate-300">{{ t('auth.registerSubtitle') }}</p>

      <form class="mt-8 space-y-4" @submit.prevent="handleSubmit">
        <div>
          <label class="text-xs uppercase tracking-widest text-slate-400">{{
            t('auth.username')
          }}</label>
          <input
            v-model="username"
            class="mt-1 w-full rounded-xl border border-white/20 bg-white/10 px-4 py-2 text-sm text-white placeholder:text-slate-400 focus:border-brand focus:outline-none"
            type="text"
            required
            minlength="3"
            autocomplete="username"
          />
        </div>
        <div>
          <label class="text-xs uppercase tracking-widest text-slate-400">{{
            t('auth.password')
          }}</label>
          <input
            v-model="password"
            class="mt-1 w-full rounded-xl border border-white/20 bg-white/10 px-4 py-2 text-sm text-white placeholder:text-slate-400 focus:border-brand focus:outline-none"
            type="password"
            required
            minlength="6"
            autocomplete="new-password"
          />
        </div>
        <div>
          <label class="text-xs uppercase tracking-widest text-slate-400">{{
            t('auth.confirmPassword')
          }}</label>
          <input
            v-model="confirmPassword"
            class="mt-1 w-full rounded-xl border border-white/20 bg-white/10 px-4 py-2 text-sm text-white placeholder:text-slate-400 focus:border-brand focus:outline-none"
            type="password"
            required
            minlength="6"
            autocomplete="new-password"
          />
        </div>
        <button
          class="w-full rounded-full bg-brand px-4 py-2 text-sm font-medium text-white shadow-elevated transition hover:bg-brand-dark disabled:cursor-not-allowed disabled:bg-brand/40"
          type="submit"
          :disabled="authStore.loading"
        >
          {{ authStore.loading ? t('common.loading') : t('auth.register') }}
        </button>
      </form>

      <p class="mt-6 text-center text-xs text-slate-400">
        {{ t('auth.haveAccount') }}
        <RouterLink class="text-brand hover:underline" :to="{ name: 'login' }">
          {{ t('auth.loginNow') }}
        </RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores'
import { emitToast } from '@/utils/notifications'

const authStore = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')

const handleSubmit = async () => {
  if (password.value !== confirmPassword.value) {
    emitToast({ type: 'warning', message: t('auth.passwordMismatch') })
    return
  }
  await authStore.performRegister({ username: username.value, password: password.value })
  router.replace({ name: 'login' })
}
</script>
