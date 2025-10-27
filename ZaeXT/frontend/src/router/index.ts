import NProgress from 'nprogress'
import { storeToRefs } from 'pinia'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { TOKEN_KEY } from '@/utils/token'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'chat',
    component: () => import('@/views/chat/ChatWorkspace.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/chat/:id',
    name: 'chat-with-id',
    component: () => import('@/views/chat/ChatWorkspace.vue'),
    meta: { requiresAuth: true },
    props: true,
  },
  {
    path: '/models',
    name: 'models',
    component: () => import('@/views/models/ModelsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/categories',
    name: 'categories',
    component: () => import('@/views/categories/CategoriesView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/recycle-bin',
    name: 'recycle-bin',
    component: () => import('@/views/recycle-bin/RecycleBinView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/profile/ProfileView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/error/:code?',
    name: 'error',
    component: () => import('@/views/errors/ErrorView.vue'),
    props: (route) => ({
      title: route.params.code ?? 'Error',
      description: (route.query.message as string) ?? 'Unexpected error occurred.',
    }),
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/errors/ErrorView.vue'),
    props: { title: '404', description: 'The page you requested could not be found.' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to, _from, next) => {
  NProgress.start()
  const token = typeof window !== 'undefined' ? window.localStorage.getItem(TOKEN_KEY) : null
  if (to.meta.requiresAuth && !token) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }

  if (token && to.meta.guestOnly) {
    next({ name: 'chat' })
    return
  }

  if (token && to.meta.requiresAuth) {
    const authStore = useAuthStore()
    const { profile, loading } = storeToRefs(authStore)
    if (!profile.value && !loading.value) {
      await authStore.fetchProfileSafely()
    }
  }

  next()
})

router.afterEach(() => {
  NProgress.done()
})

export { router }
