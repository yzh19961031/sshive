// frontend/src/stores/theme.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeId = 'dark-pro' | 'cyber-neon' | 'light-clean' | 'solarized' | 'dracula' | 'nord'

export const themes: { id: ThemeId; label: string }[] = [
  { id: 'dark-pro',     label: '暗黑专业' },
  { id: 'cyber-neon',   label: '赛博霓虹' },
  { id: 'light-clean',  label: '浅色简约' },
  { id: 'solarized',    label: 'Solarized' },
  { id: 'dracula',      label: 'Dracula' },
  { id: 'nord',         label: 'Nord' },
]

export const useThemeStore = defineStore('theme', () => {
  const current = ref<ThemeId>(
    (localStorage.getItem('theme') as ThemeId) ?? 'dark-pro'
  )

  function apply(id: ThemeId) {
    current.value = id
    document.documentElement.setAttribute('data-theme', id)
    localStorage.setItem('theme', id)
  }

  function init() { apply(current.value) }

  return { current, themes, apply, init }
})
