<template>
  <div class="flex w-full items-center justify-center py-12 text-center text-sm text-slate-400">
    <svg class="mr-3 h-5 w-5 animate-spin text-brand" viewBox="0 0 24 24" fill="none">
      <circle
        class="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="4"
      ></circle>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
    </svg>
    <slot>Loading...</slot>
  </div>
</template>
