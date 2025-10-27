<template>
  <div
    class="flex flex-col items-center justify-center gap-3 rounded-2xl border border-rose-200/40 bg-rose-500/10 p-6 text-center dark:border-rose-500/30 dark:bg-rose-500/10"
  >
    <h3 class="text-sm font-semibold text-rose-600 dark:text-rose-300">
      <slot name="title">发生错误</slot>
    </h3>
    <p class="max-w-xs text-xs text-rose-500/80 dark:text-rose-200/80">
      <slot name="description">请稍后重试或联系管理员。</slot>
    </p>
    <slot name="action"></slot>
  </div>
</template>
