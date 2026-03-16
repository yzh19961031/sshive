<!-- frontend/src/layouts/MainLayout.vue -->
<template>
  <div class="layout">
    <nav class="sidebar" :class="{ collapsed }">
      <!-- Logo -->
      <div class="logo-row">
        <div class="logo">S</div>
        <span v-if="!collapsed" class="logo-text">SSHive</span>
      </div>

      <!-- Nav groups -->
      <div class="nav-body">
        <template v-for="group in navGroups" :key="group.label">
          <div v-if="!collapsed" class="group-label">{{ group.label }}</div>
          <div v-else class="group-divider" />
          <router-link
            v-for="item in group.items" :key="item.path"
            :to="item.path"
            class="nav-item"
            :title="collapsed ? item.label : undefined"
          >
            <span class="nav-icon">{{ item.icon }}</span>
            <span v-if="!collapsed" class="nav-text">{{ item.label }}</span>
          </router-link>
        </template>
      </div>

      <!-- Bottom: theme picker + collapse toggle + logout -->
      <div class="sidebar-bottom">
        <div class="nav-item" :title="collapsed ? '切换主题' : undefined" @click="cycleTheme">
          <span class="nav-icon">🎨</span>
          <span v-if="!collapsed" class="nav-text">{{ currentThemeLabel }}</span>
        </div>
        <div class="nav-item" :title="collapsed ? '退出登录' : undefined" @click="handleLogout">
          <span class="nav-icon">⏻</span>
          <span v-if="!collapsed" class="nav-text">退出登录</span>
        </div>
        <div class="nav-item collapse-btn" @click="toggleCollapse"
          :title="collapsed ? '展开侧边栏' : '收起侧边栏'">
          <span class="nav-icon">{{ collapsed ? '→' : '←' }}</span>
          <span v-if="!collapsed" class="nav-text">收起</span>
        </div>
      </div>
    </nav>

    <main class="content">
      <router-view v-slot="{ Component, route }">
        <keep-alive>
          <component :is="Component" :key="route.path" />
        </keep-alive>
      </router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useDialog } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'

const auth = useAuthStore()
const themeStore = useThemeStore()
const dialog = useDialog()
themeStore.init()

function handleLogout() {
  dialog.warning({
    title: '退出登录',
    content: '确定要退出登录吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: () => auth.logout(),
  })
}

// Sidebar collapse state
const collapsed = ref(localStorage.getItem('sidebar-collapsed') === '1')
function toggleCollapse() {
  collapsed.value = !collapsed.value
  localStorage.setItem('sidebar-collapsed', collapsed.value ? '1' : '0')
}

// Theme cycling
const currentThemeLabel = computed(() => {
  const t = themeStore.themes.find(t => t.id === themeStore.current)
  return t?.label ?? themeStore.current
})
function cycleTheme() {
  const ids = themeStore.themes.map(t => t.id)
  const idx = ids.indexOf(themeStore.current)
  themeStore.apply(ids[(idx + 1) % ids.length]!)
}

const navGroups = [
  {
    label: '运维',
    items: [
      { path: '/dashboard', icon: '📊', label: '仪表盘' },
      { path: '/hosts',       icon: '⬡',  label: '主机列表' },
      { path: '/host-groups', icon: '◫',  label: '主机分组' },
      { path: '/terminal',  icon: '⌨',  label: 'Terminal' },
    ],
  },
  {
    label: '安全',
    items: [
      { path: '/audit/sessions',  icon: '📋', label: '审计日志' },
      { path: '/audit/commands',  icon: '⌫',  label: '命令管控' },
      { path: '/settings/rules',  icon: '⚠', label: '高危规则' },
    ],
  },
  {
    label: '系统',
    items: [
      { path: '/settings/users',   icon: '👥', label: '用户管理' },
      { path: '/settings/profile', icon: '⚙', label: '个人设置' },
    ],
  },
]
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* ── Sidebar ── */
.sidebar {
  width: 160px;
  flex-shrink: 0;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 10px 0 10px;
  transition: width 0.2s ease;
  overflow: hidden;
}
.sidebar.collapsed {
  width: 56px;
}

/* Logo row */
.logo-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  margin-bottom: 12px;
  min-height: 32px;
}
.logo {
  width: 26px; height: 26px;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 12px;
  flex-shrink: 0;
}
.logo-text {
  font-weight: 700;
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

/* Group label */
.group-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-secondary);
  padding: 10px 14px 4px;
  white-space: nowrap;
}
.group-divider {
  height: 1px;
  background: var(--border);
  margin: 8px 10px;
}

/* Nav items */
.nav-body {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 34px;
  padding: 0 12px;
  border-radius: 6px;
  margin: 1px 6px;
  cursor: pointer;
  text-decoration: none;
  color: var(--text-secondary);
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
  position: relative;
}
.nav-item:hover,
.nav-item.router-link-active {
  background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent);
}
.nav-item.router-link-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 4px;
  bottom: 4px;
  width: 2px;
  background: var(--accent);
  border-radius: 2px;
}
.nav-icon { font-size: 15px; flex-shrink: 0; width: 20px; text-align: center; display: inline-block; }
.nav-text { font-size: 13px; overflow: hidden; text-overflow: ellipsis; }

/* Bottom section */
.sidebar-bottom {
  border-top: 1px solid var(--border);
  padding-top: 6px;
}
.collapse-btn {
  opacity: 0.6;
}
.collapse-btn:hover { opacity: 1; }

/* Main content */
.content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
}
</style>
