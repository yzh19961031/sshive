<!-- frontend/src/views/ProfileView.vue -->
<template>
  <div class="page">
    <h2 class="page-title">个人设置</h2>

    <section class="section">
      <h3 class="section-title">界面主题</h3>
      <div class="theme-grid">
        <div v-for="t in themeStore.themes" :key="t.id"
          :class="['theme-card', themeStore.current === t.id && 'active']"
          @click="themeStore.apply(t.id)">
          <div class="theme-preview">
            <div class="preview-sidebar" />
            <div class="preview-content">
              <div class="preview-bar" />
              <div class="preview-lines">
                <div /><div style="width:70%" /><div style="width:85%" />
              </div>
            </div>
          </div>
          <span class="theme-name">{{ t.label }}</span>
          <span v-if="themeStore.current === t.id" class="theme-check">✓</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'
const themeStore = useThemeStore()
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-title { margin: 0 0 24px; font-size: 16px; font-weight: 600; }
.section { margin-bottom: 32px; }
.section-title { margin: 0 0 16px; font-size: 14px; font-weight: 500; color: var(--text-secondary); }
.theme-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 12px; }
.theme-card { cursor: pointer; border: 2px solid var(--border); border-radius: 8px;
  overflow: hidden; position: relative; transition: border-color 0.15s; }
.theme-card:hover { border-color: var(--text-muted); }
.theme-card.active { border-color: var(--accent); }
.theme-preview { height: 80px; display: flex; }
.preview-sidebar { width: 20px; background: #010409; }
.preview-content { flex: 1; background: #0d1117; padding: 6px; display: flex; flex-direction: column; gap: 4px; }
.preview-bar { height: 6px; background: #1f6feb; border-radius: 2px; width: 60%; }
.preview-lines > div { height: 4px; background: #30363d; border-radius: 1px; margin-bottom: 3px; width: 100%; }
.theme-name { display: block; padding: 6px 8px; font-size: 12px; text-align: center; }
.theme-check { position: absolute; top: 6px; right: 6px; color: var(--accent); font-size: 14px; }
</style>
