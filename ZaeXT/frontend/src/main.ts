import './styles/main.css'
import 'nprogress/nprogress.css'
import 'highlight.js/styles/github-dark.css'

import { createApp } from 'vue'

import App from './App.vue'
import { i18n } from './i18n'
import { router } from './router'
import { pinia } from './stores'

async function bootstrap() {
  const app = createApp(App)

  app.use(pinia)
  app.use(router)
  app.use(i18n)

  await router.isReady()
  app.mount('#app')
}

bootstrap()
