<template>
  <li
    class="rounded-xl border border-slate-200/80 bg-white/70 p-4 dark:border-slate-700/60 dark:bg-slate-900/60"
  >
    <div class="flex items-start justify-between gap-3">
      <div>
        <p class="text-sm font-semibold text-slate-800 dark:text-slate-200">{{ node.name }}</p>
        <p class="text-xs text-slate-400">ID: {{ node.id }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-1 text-xs">
        <button
          class="rounded-full border border-slate-300 px-2 py-1 text-slate-500 hover:border-brand hover:text-brand"
          type="button"
          @click="addChild"
        >
          {{ t('categories.addChild') }}
        </button>
        <button
          class="rounded-full border border-slate-300 px-2 py-1 text-slate-500 hover:border-brand hover:text-brand"
          type="button"
          @click="rename"
        >
          {{ t('categories.rename') }}
        </button>
        <button
          class="rounded-full border border-rose-200 px-2 py-1 text-rose-500 hover:bg-rose-500/10"
          type="button"
          @click="remove"
        >
          {{ t('categories.remove') }}
        </button>
      </div>
    </div>
    <ul
      v-if="node.children?.length"
      class="mt-3 space-y-2 border-l border-dashed border-slate-300/60 pl-4"
    >
      <CategoryNode v-for="child in node.children" :key="child.id" :node="child" />
    </ul>
  </li>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { useCategoryStore } from '@/stores'
import type { CategoryTreeNode } from '@/types'

const props = defineProps<{ node: CategoryTreeNode }>()

const categoryStore = useCategoryStore()
const { t } = useI18n()

const addChild = async () => {
  const name = window.prompt(t('categories.promptName'))
  if (!name) return
  await categoryStore.addCategory({ name, parent_id: props.node.id })
}

const rename = async () => {
  const name = window.prompt(t('categories.promptRename'), props.node.name)
  if (!name || name === props.node.name) return
  await categoryStore.patchCategory(props.node.id, { name })
}

const remove = async () => {
  if (!window.confirm(t('categories.confirmDelete', { name: props.node.name }))) return
  await categoryStore.removeCategory(props.node.id)
}
</script>
