<template>
  <div
    class="grid min-h-screen place-items-center bg-gradient-to-br from-slate-900 via-slate-950 to-black px-4 py-12"
  >
    <div
      class="w-full max-w-md rounded-3xl border border-white/10 bg-white/5 p-8 shadow-2xl backdrop-blur"
    >
      <h1 class="text-2xl font-semibold text-white">{{ t('auth.loginTitle') }}</h1>
      <p class="mt-2 text-sm text-slate-300">{{ t('auth.loginSubtitle') }}</p>

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
            autocomplete="current-password"
          />
        </div>
        <button
          class="w-full rounded-full bg-brand px-4 py-2 text-sm font-medium text-white shadow-elevated transition hover:bg-brand-dark disabled:cursor-not-allowed disabled:bg-brand/40"
          type="submit"
          :disabled="authStore.loading"
        >
          {{ authStore.loading ? t('common.loading') : t('auth.login') }}
        </button>
      </form>

      <p class="mt-6 text-center text-xs text-slate-400">
        {{ t('auth.noAccount') }}
        <RouterLink class="text-brand hover:underline" :to="{ name: 'register' }">
          {{ t('auth.registerNow') }}
        </RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const username = ref('')
const password = ref('')

onMounted(() => {
  if (authStore.isAuthenticated) {
    router.replace({ name: 'chat' })
  }
})

const handleSubmit = async () => {
  await authStore.performLogin({ username: username.value, password: password.value })
  const redirect = (route.query.redirect as string) || '/'
  router.replace(redirect)
}
</script>
