<template>
  <RouterView />
  <AppToaster />
  <AppSettingsWidget />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppSettingsWidget from '@/components/AppSettingsWidget.vue'
import AppToaster from '@/components/AppToaster.vue'
</script>
