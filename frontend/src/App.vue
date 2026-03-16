<!-- frontend/src/App.vue -->
<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides">
    <n-dialog-provider>
      <router-view />
    </n-dialog-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NConfigProvider, NDialogProvider, darkTheme } from 'naive-ui'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
theme.init()

// Light theme uses null (naive-ui default), others use dark
const naiveTheme = computed(() =>
  theme.current === 'light-clean' ? null : darkTheme
)

const themeOverrides = computed(() => ({
  common: {
    primaryColor: '#409eff',
    primaryColorHover: '#60a5fa',
  }
}))
</script>
