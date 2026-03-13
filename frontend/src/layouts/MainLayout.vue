<!-- frontend/src/layouts/MainLayout.vue -->
<template>
  <div class="layout">
    <!-- 44px icon sidebar -->
    <nav class="sidebar">
      <div class="logo">S</div>
      <router-link v-for="item in navItems" :key="item.path"
        :to="item.path" class="nav-item" :title="item.label">
        <span class="nav-icon">{{ item.icon }}</span>
      </router-link>
      <div class="sidebar-bottom">
        <div class="nav-item" title="退出" @click="auth.logout()">
          <span class="nav-icon">⏻</span>
        </div>
      </div>
    </nav>
    <!-- Content area -->
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'

const auth = useAuthStore()
const theme = useThemeStore()
theme.init()

const navItems = [
  { path: '/hosts',            icon: '⬡', label: '主机列表' },
  { path: '/terminal',         icon: '⌨', label: 'Terminal' },
  { path: '/audit/sessions',   icon: '📋', label: '审计日志' },
  { path: '/settings/rules',   icon: '⚠', label: '高危规则' },
  { path: '/settings/users',   icon: '👥', label: '用户管理' },
  { path: '/settings/profile', icon: '⚙', label: '个人设置' },
]
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}
.sidebar {
  width: 44px;
  flex-shrink: 0;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 0;
  gap: 4px;
}
.logo {
  width: 28px; height: 28px;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 13px;
  margin-bottom: 8px;
}
.nav-item {
  width: 34px; height: 34px;
  border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer;
  text-decoration: none;
  color: var(--text-secondary);
  transition: background 0.15s, color 0.15s;
}
.nav-item:hover,
.nav-item.router-link-active {
  background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent);
}
.nav-icon { font-size: 16px; }
.sidebar-bottom { margin-top: auto; }
.content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
}
</style>
