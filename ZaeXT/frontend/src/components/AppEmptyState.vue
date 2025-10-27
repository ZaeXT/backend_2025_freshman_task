<template>
  <div class="flex w-full flex-col items-center justify-center gap-2 py-12 text-center">
    <div
      class="flex h-16 w-16 items-center justify-center rounded-full border border-dashed border-slate-300 text-slate-400 dark:border-slate-600"
    >
      <slot name="icon">🗒️</slot>
    </div>
    <h3 class="text-sm font-medium text-slate-600 dark:text-slate-300">
      <slot name="title">暂无数据</slot>
    </h3>
    <p class="max-w-xs text-xs text-slate-400 dark:text-slate-500">
      <slot name="description">稍后再试或创建新内容。</slot>
    </p>
    <slot name="action"></slot>
  </div>
</template>
