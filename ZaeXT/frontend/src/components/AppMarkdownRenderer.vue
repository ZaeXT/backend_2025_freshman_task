<template>
  <div ref="markdownEl" class="prose prose-slate max-w-none dark:prose-invert" v-html="sanitized" />
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import MarkdownIt from 'markdown-it'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps<{
  content: string
}>()

const baseMarkdown = new MarkdownIt()

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  highlight(str: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        const result = hljs.highlight(str, { language: lang, ignoreIllegals: true })
        return `<pre class="code-block"><code class="hljs language-${lang}">${result.value}</code></pre>`
      } catch (error) {
        console.error('[markdown] highlight error', error)
      }
    }
    const escaped = baseMarkdown.utils.escapeHtml(str)
    return `<pre class="code-block"><code class="hljs">${escaped}</code></pre>`
  },
})

const markdownEl = ref<HTMLDivElement | null>(null)

const rendered = computed(() => markdown.render(props.content || ''))

const sanitized = computed(() => DOMPurify.sanitize(rendered.value))

const attachCopyButtons = () => {
  nextTick(() => {
    const container = markdownEl.value
    if (!container) return
    const blocks = container.querySelectorAll('pre.code-block')
    blocks.forEach((block) => {
      if (block.querySelector('button.copy-button')) return
      const button = document.createElement('button')
      button.className =
        'copy-button absolute right-3 top-3 rounded-md bg-slate-800 px-2 py-1 text-xs text-slate-200 opacity-0 transition hover:bg-slate-700 group-hover:opacity-100'
      button.textContent = '复制'
      button.addEventListener('click', () => {
        const code = block.querySelector('code')?.textContent ?? ''
        if (navigator?.clipboard) {
          navigator.clipboard.writeText(code).then(() => {
            button.textContent = '已复制'
            setTimeout(() => {
              button.textContent = '复制'
            }, 2000)
          })
        }
      })
      block.classList.add('group', 'relative', 'rounded-xl', 'bg-slate-900/90', 'p-4')
      block.appendChild(button)
    })
  })
}

watch(
  () => sanitized.value,
  () => {
    attachCopyButtons()
  },
  { flush: 'post' },
)

onMounted(() => {
  attachCopyButtons()
})

onBeforeUnmount(() => {
  const blocks = markdownEl.value?.querySelectorAll('pre.code-block button.copy-button') ?? []
  blocks.forEach((button) => button.remove())
})
</script>

<style scoped>
.prose pre {
  @apply relative rounded-xl border border-slate-200/40 bg-slate-900/80 p-4 text-xs dark:border-slate-700/60;
}

.prose code {
  @apply font-mono;
}

.copy-button {
  @apply pointer-events-auto;
}

.prose
  :where(p:first-child, ul:first-child, ol:first-child, pre:first-child, blockquote:first-child) {
  margin-top: 0;
}

.prose :where(p:last-child, ul:last-child, ol:last-child, pre:last-child, blockquote:last-child) {
  margin-bottom: 0;
}

.prose {
  @apply break-words whitespace-pre-wrap;
}
</style>
