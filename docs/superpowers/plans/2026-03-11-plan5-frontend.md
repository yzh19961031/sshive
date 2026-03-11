# SSHive Plan 5: 前端（Vue 3 + 全部页面）

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现完整的 Vue 3 前端，包含图标侧边栏布局、6套主题、主机列表（分组+视图切换）、WebSSH Terminal（Tab+分屏）、SFTP文件管理、审计日志、审计回放、设置页面。

**Architecture:** Vite + Vue 3 + TypeScript + Naive UI + Pinia + Vue Router。CSS 变量实现主题切换，xterm.js 渲染 Terminal，asciinema-player 实现录像回放。前端构建产物通过 go:embed 嵌入后端二进制。

**Tech Stack:** Vue 3, Vite, TypeScript, Naive UI, Pinia, Vue Router 4, xterm.js, asciinema-player, axios

**前置条件:** Plan 1-4 全部完成，后端 API 可访问。

---

## 文件结构

```
frontend/                          # 前端根目录（独立于 Go 模块）
├── package.json
├── vite.config.ts
├── tsconfig.json
├── index.html
└── src/
    ├── main.ts                    # 入口
    ├── App.vue                    # 根组件（主题挂载）
    ├── router/
    │   └── index.ts               # 路由配置
    ├── stores/
    │   ├── auth.ts                # 登录态、JWT 存储
    │   └── theme.ts               # 主题状态
    ├── api/
    │   ├── client.ts              # axios 实例（带 JWT 拦截器）
    │   ├── auth.ts                # 登录/登出接口
    │   ├── host.ts                # 主机接口
    │   ├── credential.ts          # 凭据接口
    │   ├── session.ts             # 审计会话接口
    │   └── dangerrule.ts          # 高危规则接口
    ├── styles/
    │   ├── themes.css             # 6套主题 CSS 变量
    │   └── global.css             # 全局样式
    ├── layouts/
    │   └── MainLayout.vue         # 图标侧边栏 + 内容区布局
    └── views/
        ├── LoginView.vue          # 登录页
        ├── HostListView.vue       # 主机列表（分组+视图切换）
        ├── TerminalView.vue       # WebSSH Terminal（Tab+分屏）
        ├── SftpView.vue           # SFTP 文件管理
        ├── AuditSessionsView.vue  # 审计会话列表
        ├── AuditReplayView.vue    # 会话录像回放
        ├── DangerRulesView.vue    # 高危规则管理
        ├── UsersView.vue          # 用户管理
        └── ProfileView.vue        # 个人设置（主题切换）
```

---

## Chunk 1: 项目初始化 + 主题系统

### Task 1: Vite + Vue 3 项目初始化

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.ts`

- [ ] **Step 1: 初始化项目**

```bash
cd /Volumes/data/personal/code/sshive
npm create vite@latest frontend -- --template vue-ts
cd frontend
```

- [ ] **Step 2: 安装依赖**

```bash
npm install naive-ui@2.38.1
npm install pinia@2.1.7
npm install vue-router@4.3.0
npm install axios@1.6.7
npm install @xterm/xterm@5.3.0 @xterm/addon-fit@0.10.0 @xterm/addon-web-links@0.11.0
npm install asciinema-player@3.7.0
npm install
```

- [ ] **Step 3: 配置 Vite 代理（开发时转发到 Go 后端）**

```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': resolve(__dirname, 'src') }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true, // 代理 WebSocket
      }
    }
  },
  build: {
    outDir: '../web', // 输出到 Go 的 web 目录
  }
})
```

- [ ] **Step 4: 创建 main.ts**

```typescript
// frontend/src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import './styles/themes.css'
import './styles/global.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

- [ ] **Step 5: 编译验证**

```bash
npm run build
```

Expected: 成功构建，输出到 `../web/`

- [ ] **Step 6: Commit**

```bash
cd ..
git add frontend/
git commit -m "chore: init vue3 frontend with vite + naive-ui"
```

---

### Task 2: 6套主题 CSS 变量

**Files:**
- Create: `frontend/src/styles/themes.css`
- Create: `frontend/src/styles/global.css`

- [ ] **Step 1: 定义主题 CSS 变量**

```css
/* frontend/src/styles/themes.css */

/* ── 暗黑专业（默认） ── */
[data-theme="dark-pro"], :root {
  --bg-base: #0d1117;
  --bg-surface: #161b22;
  --bg-elevated: #21262d;
  --bg-sidebar: #010409;
  --border: #30363d;
  --text-primary: #e6edf3;
  --text-secondary: #8b949e;
  --text-muted: #6e7681;
  --accent: #1f6feb;
  --accent-hover: #388bfd;
  --success: #3fb950;
  --warning: #d29922;
  --danger: #f85149;
  --terminal-bg: #0d1117;
  --terminal-fg: #c9d1d9;
}

/* ── 赛博霓虹 ── */
[data-theme="cyber-neon"] {
  --bg-base: #070b14;
  --bg-surface: #0d1526;
  --bg-elevated: #152040;
  --bg-sidebar: #040810;
  --border: #00e5ff22;
  --text-primary: #e0f7ff;
  --text-secondary: #7ecfdf;
  --text-muted: #4a8a9a;
  --accent: #00e5ff;
  --accent-hover: #40f0ff;
  --success: #69ff47;
  --warning: #ffcc00;
  --danger: #ff4081;
  --terminal-bg: #060a12;
  --terminal-fg: #00e5ff;
}

/* ── 浅色简约 ── */
[data-theme="light-clean"] {
  --bg-base: #f8fafc;
  --bg-surface: #ffffff;
  --bg-elevated: #f1f5f9;
  --bg-sidebar: #1e1e2e;
  --border: #e2e8f0;
  --text-primary: #0f172a;
  --text-secondary: #64748b;
  --text-muted: #94a3b8;
  --accent: #6366f1;
  --accent-hover: #4f46e5;
  --success: #10b981;
  --warning: #f59e0b;
  --danger: #ef4444;
  --terminal-bg: #1e1e2e;
  --terminal-fg: #cdd6f4;
}

/* ── Solarized Dark ── */
[data-theme="solarized"] {
  --bg-base: #002b36;
  --bg-surface: #073642;
  --bg-elevated: #084652;
  --bg-sidebar: #001f27;
  --border: #094551;
  --text-primary: #839496;
  --text-secondary: #657b83;
  --text-muted: #586e75;
  --accent: #268bd2;
  --accent-hover: #2aa198;
  --success: #859900;
  --warning: #b58900;
  --danger: #dc322f;
  --terminal-bg: #002b36;
  --terminal-fg: #839496;
}

/* ── Dracula ── */
[data-theme="dracula"] {
  --bg-base: #282a36;
  --bg-surface: #313442;
  --bg-elevated: #44475a;
  --bg-sidebar: #191a21;
  --border: #44475a;
  --text-primary: #f8f8f2;
  --text-secondary: #6272a4;
  --text-muted: #4d5374;
  --accent: #bd93f9;
  --accent-hover: #cfa9ff;
  --success: #50fa7b;
  --warning: #ffb86c;
  --danger: #ff5555;
  --terminal-bg: #282a36;
  --terminal-fg: #f8f8f2;
}

/* ── Nord ── */
[data-theme="nord"] {
  --bg-base: #2e3440;
  --bg-surface: #3b4252;
  --bg-elevated: #434c5e;
  --bg-sidebar: #242933;
  --border: #4c566a;
  --text-primary: #eceff4;
  --text-secondary: #9099aa;
  --text-muted: #677281;
  --accent: #88c0d0;
  --accent-hover: #81a1c1;
  --success: #a3be8c;
  --warning: #ebcb8b;
  --danger: #bf616a;
  --terminal-bg: #2e3440;
  --terminal-fg: #d8dee9;
}
```

- [ ] **Step 2: 全局样式**

```css
/* frontend/src/styles/global.css */
*, *::before, *::after { box-sizing: border-box; }

html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  background: var(--bg-base);
  color: var(--text-primary);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  font-size: 14px;
}

::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--bg-elevated); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--text-muted); }
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/styles/
git commit -m "feat: add 6 built-in themes with CSS variables"
```

---

### Task 3: Pinia Stores（Auth + Theme）

**Files:**
- Create: `frontend/src/stores/auth.ts`
- Create: `frontend/src/stores/theme.ts`

- [ ] **Step 1: Auth Store**

```typescript
// frontend/src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { router } from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') ?? '')
  const isLoggedIn = computed(() => !!token.value)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('token', t)
  }

  async function login(username: string, password: string, tenantId: number) {
    const res = await authApi.login({ username, password, tenant_id: tenantId })
    setToken(res.data.data.token)
    router.push('/hosts')
  }

  async function logout() {
    try { await authApi.logout() } catch {}
    token.value = ''
    localStorage.removeItem('token')
    router.push('/login')
  }

  return { token, isLoggedIn, login, logout }
})
```

- [ ] **Step 2: Theme Store**

```typescript
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

  // 初始化时应用已保存的主题
  function init() { apply(current.value) }

  return { current, themes, apply, init }
})
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/stores/
git commit -m "feat: add auth and theme pinia stores"
```

---

### Task 4: API Client + 接口封装

**Files:**
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/api/auth.ts`
- Create: `frontend/src/api/host.ts`
- Create: `frontend/src/api/session.ts`
- Create: `frontend/src/api/dangerrule.ts`

- [ ] **Step 1: axios 实例**

```typescript
// frontend/src/api/client.ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

export const http = axios.create({ baseURL: '/api' })

http.interceptors.request.use(cfg => {
  const auth = useAuthStore()
  if (auth.token) cfg.headers.Authorization = `Bearer ${auth.token}`
  return cfg
})

http.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
    }
    return Promise.reject(err)
  }
)
```

- [ ] **Step 2: Auth API**

```typescript
// frontend/src/api/auth.ts
import { http } from './client'

export const authApi = {
  login: (data: { username: string; password: string; tenant_id: number }) =>
    http.post('/auth/login', data),
  logout: () => http.post('/auth/logout'),
}
```

- [ ] **Step 3: Host API**

```typescript
// frontend/src/api/host.ts
import { http } from './client'

export interface Host {
  id: number
  name: string
  ip: string
  port: number
  auth_type: string
  credential_id: number
  status: number
  tags: string[]
  created_at: string
}

export const hostApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get<{ data: { total: number; list: Host[] } }>('/hosts', { params }),
  create: (data: Omit<Host, 'id' | 'created_at'>) =>
    http.post<{ data: Host }>('/hosts', data),
  update: (id: number, data: Partial<Host>) =>
    http.put<{ data: Host }>(`/hosts/${id}`, data),
  delete: (id: number) => http.delete(`/hosts/${id}`),
}
```

- [ ] **Step 4: Session API**

```typescript
// frontend/src/api/session.ts
import { http } from './client'

export const sessionApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/sessions', { params }),
  getLogs: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/logs`, { params }),
  getCommands: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/commands`, { params }),
  getReplayUrl: (id: number) => `/api/sessions/${id}/replay`,
}
```

- [ ] **Step 5: DangerRule API**

```typescript
// frontend/src/api/dangerrule.ts
import { http } from './client'

export const dangerRuleApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/danger-rules', { params }),
  create: (data: { pattern: string; description?: string; action?: string }) =>
    http.post('/danger-rules', data),
  update: (id: number, data: { pattern?: string; description?: string; enabled?: number }) =>
    http.put(`/danger-rules/${id}`, data),
  delete: (id: number) => http.delete(`/danger-rules/${id}`),
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/
git commit -m "feat: add API client and interface wrappers"
```

---

## Chunk 2: 路由 + 布局 + 登录

### Task 5: Vue Router + 主布局

**Files:**
- Create: `frontend/src/router/index.ts`
- Create: `frontend/src/layouts/MainLayout.vue`
- Create: `frontend/src/App.vue`

- [ ] **Step 1: 配置路由**

```typescript
// frontend/src/router/index.ts
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', redirect: '/hosts' },
      { path: 'hosts', component: () => import('@/views/HostListView.vue') },
      { path: 'terminal', component: () => import('@/views/TerminalView.vue') },
      { path: 'sftp/:hostId', component: () => import('@/views/SftpView.vue') },
      { path: 'audit/sessions', component: () => import('@/views/AuditSessionsView.vue') },
      { path: 'audit/replay/:id', component: () => import('@/views/AuditReplayView.vue') },
      { path: 'settings/rules', component: () => import('@/views/DangerRulesView.vue') },
      { path: 'settings/users', component: () => import('@/views/UsersView.vue') },
      { path: 'settings/profile', component: () => import('@/views/ProfileView.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/hosts' }
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫：未登录跳转登录页
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
})
```

- [ ] **Step 2: 图标侧边栏主布局**

```vue
<!-- frontend/src/layouts/MainLayout.vue -->
<template>
  <div class="layout">
    <!-- 44px 图标侧边栏 -->
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
    <!-- 内容区 -->
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
  { path: '/hosts',           icon: '⬡', label: '主机列表' },
  { path: '/terminal',        icon: '⌨', label: 'Terminal' },
  { path: '/audit/sessions',  icon: '📋', label: '审计日志' },
  { path: '/settings/rules',  icon: '⚠', label: '高危规则' },
  { path: '/settings/users',  icon: '👥', label: '用户管理' },
  { path: '/settings/profile',icon: '⚙', label: '个人设置' },
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
```

- [ ] **Step 3: App.vue**

```vue
<!-- frontend/src/App.vue -->
<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides">
    <router-view />
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NConfigProvider, darkTheme } from 'naive-ui'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
theme.init()

// 浅色主题用 null（naive-ui 默认），其余用暗色
const naiveTheme = computed(() =>
  theme.current === 'light-clean' ? null : darkTheme
)

const themeOverrides = computed(() => ({
  common: {
    primaryColor: 'var(--accent)',
    primaryColorHover: 'var(--accent-hover)',
  }
}))
</script>
```

- [ ] **Step 4: 编译验证**

```bash
cd frontend && npm run build
```

Expected: 无报错

- [ ] **Step 5: Commit**

```bash
cd .. && git add frontend/src/router/ frontend/src/layouts/ frontend/src/App.vue
git commit -m "feat: add router, main layout with icon sidebar"
```

---

### Task 6: 登录页

**Files:**
- Create: `frontend/src/views/LoginView.vue`

- [ ] **Step 1: 实现登录页**

```vue
<!-- frontend/src/views/LoginView.vue -->
<template>
  <div class="login-page">
    <div class="login-card">
      <div class="brand">
        <div class="brand-logo">S</div>
        <h1 class="brand-name">SSHive</h1>
        <p class="brand-sub">运维审计平台</p>
      </div>
      <n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
        <n-form-item label="租户 ID" path="tenantId">
          <n-input-number v-model:value="form.tenantId" :min="1" placeholder="1" style="width:100%" />
        </n-form-item>
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="admin" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="••••••••"
            show-password-on="click" @keydown.enter="handleLogin" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="handleLogin">
          登录
        </n-button>
      </n-form>
      <p v-if="error" class="error-msg">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NForm, NFormItem, NInput, NInputNumber, NButton } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const form = ref({ tenantId: 1, username: '', password: '' })
const loading = ref(false)
const error = ref('')

const rules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }],
}

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(form.value.username, form.value.password, form.value.tenantId)
  } catch (e: any) {
    error.value = e.response?.data?.message ?? '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-base);
}
.login-card {
  width: 380px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 40px;
}
.brand { text-align: center; margin-bottom: 32px; }
.brand-logo {
  width: 48px; height: 48px;
  border-radius: 12px;
  background: var(--accent);
  color: #fff;
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 22px; font-weight: 700;
  margin-bottom: 12px;
}
.brand-name { margin: 0; font-size: 22px; font-weight: 700; color: var(--text-primary); }
.brand-sub { margin: 4px 0 0; color: var(--text-secondary); font-size: 13px; }
.error-msg { margin-top: 12px; color: var(--danger); font-size: 13px; text-align: center; }
</style>
```

- [ ] **Step 2: 编译验证**

```bash
cd frontend && npm run build
```

- [ ] **Step 3: Commit**

```bash
cd .. && git add frontend/src/views/LoginView.vue
git commit -m "feat: add login page"
```

---

## Chunk 3: 核心页面

### Task 7: 主机列表页（分组 + 视图切换）

**Files:**
- Create: `frontend/src/views/HostListView.vue`

- [ ] **Step 1: 实现主机列表页**

```vue
<!-- frontend/src/views/HostListView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">主机列表</h2>
      <div class="header-actions">
        <n-input v-model:value="search" placeholder="搜索主机..." size="small"
          style="width:200px" clearable />
        <div class="view-toggle">
          <button :class="['toggle-btn', viewMode === 'card' && 'active']" @click="viewMode = 'card'">⊞</button>
          <button :class="['toggle-btn', viewMode === 'list' && 'active']" @click="viewMode = 'list'">☰</button>
        </div>
        <n-button type="primary" size="small" @click="showAddModal = true">+ 添加主机</n-button>
      </div>
    </div>

    <div class="host-groups" v-if="groupedHosts.length">
      <div v-for="group in groupedHosts" :key="group.label" class="group">
        <div class="group-header" @click="toggleGroup(group.label)">
          <span class="group-arrow">{{ collapsed[group.label] ? '▶' : '▾' }}</span>
          <span class="group-label">{{ group.label }}</span>
          <span class="group-count">{{ group.hosts.length }}</span>
        </div>
        <transition name="slide">
          <div v-if="!collapsed[group.label]">
            <!-- 卡片视图 -->
            <div v-if="viewMode === 'card'" class="card-grid">
              <div v-for="host in group.hosts" :key="host.id" class="host-card">
                <div class="card-header">
                  <div :class="['status-dot', host.status === 1 ? 'online' : 'offline']" />
                  <span class="host-name">{{ host.name }}</span>
                </div>
                <div class="host-meta">{{ host.ip }}:{{ host.port }}</div>
                <div class="host-tags">
                  <span v-for="tag in host.tags" :key="tag" class="tag">{{ tag }}</span>
                </div>
                <div class="card-actions">
                  <n-button size="tiny" type="primary" @click="connectSSH(host)">SSH</n-button>
                  <n-button size="tiny" @click="connectSFTP(host)">SFTP</n-button>
                </div>
              </div>
            </div>
            <!-- 列表视图 -->
            <div v-else class="host-list">
              <div v-for="host in group.hosts" :key="host.id" class="list-row">
                <div :class="['status-dot', host.status === 1 ? 'online' : 'offline']" />
                <span class="host-name">{{ host.name }}</span>
                <span class="host-ip">{{ host.ip }}:{{ host.port }}</span>
                <div class="list-actions">
                  <n-button size="tiny" type="primary" @click="connectSSH(host)">SSH</n-button>
                  <n-button size="tiny" @click="connectSFTP(host)">SFTP</n-button>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </div>
    <n-empty v-else description="暂无主机" style="margin-top: 60px" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NInput, NEmpty } from 'naive-ui'
import { hostApi, type Host } from '@/api/host'

const router = useRouter()
const hosts = ref<Host[]>([])
const search = ref('')
const viewMode = ref<'card' | 'list'>('card')
const showAddModal = ref(false)
const collapsed = reactive<Record<string, boolean>>({})

onMounted(async () => {
  const res = await hostApi.list({ page: 1, page_size: 100 })
  hosts.value = res.data.data?.list ?? []
})

const filteredHosts = computed(() =>
  hosts.value.filter(h =>
    !search.value ||
    h.name.includes(search.value) ||
    h.ip.includes(search.value)
  )
)

// 按第一个 tag 分组（无 tag 归入"未分组"）
const groupedHosts = computed(() => {
  const map = new Map<string, Host[]>()
  for (const h of filteredHosts.value) {
    const label = h.tags?.[0] ?? '未分组'
    if (!map.has(label)) map.set(label, [])
    map.get(label)!.push(h)
  }
  return [...map.entries()].map(([label, hs]) => ({ label, hosts: hs }))
})

function toggleGroup(label: string) {
  collapsed[label] = !collapsed[label]
}

function connectSSH(host: Host) {
  // 将主机信息存入 sessionStorage，Terminal 页面读取
  const sessions: any[] = JSON.parse(sessionStorage.getItem('pendingSSH') ?? '[]')
  sessions.push({ hostId: host.id, hostName: host.name })
  sessionStorage.setItem('pendingSSH', JSON.stringify(sessions))
  router.push('/terminal')
}

function connectSFTP(host: Host) {
  router.push(`/sftp/${host.id}`)
}
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; margin-bottom: 20px; gap: 12px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
.header-actions { display: flex; align-items: center; gap: 8px; margin-left: auto; }
.view-toggle { display: flex; border: 1px solid var(--border); border-radius: 5px; overflow: hidden; }
.toggle-btn { background: transparent; border: none; color: var(--text-secondary);
  padding: 4px 8px; cursor: pointer; font-size: 14px; }
.toggle-btn.active { background: var(--accent); color: #fff; }
.group { margin-bottom: 16px; }
.group-header { display: flex; align-items: center; gap: 8px; padding: 6px 0;
  cursor: pointer; color: var(--text-secondary); font-size: 13px; user-select: none; }
.group-arrow { font-size: 10px; }
.group-count { background: var(--bg-elevated); padding: 1px 7px; border-radius: 10px; font-size: 11px; }
.card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px; padding: 4px 0 8px; }
.host-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 8px;
  padding: 12px; display: flex; flex-direction: column; gap: 6px; }
.card-header { display: flex; align-items: center; gap: 6px; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.status-dot.online { background: var(--success); }
.status-dot.offline { background: var(--danger); }
.host-name { font-weight: 500; font-size: 13px; }
.host-meta { color: var(--text-secondary); font-size: 12px; }
.host-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.tag { background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent); font-size: 11px; padding: 1px 6px; border-radius: 3px; }
.card-actions { display: flex; gap: 6px; margin-top: 4px; }
.host-list { display: flex; flex-direction: column; gap: 3px; padding: 4px 0 8px; }
.list-row { display: flex; align-items: center; gap: 10px; background: var(--bg-surface);
  border: 1px solid var(--border); border-radius: 6px; padding: 8px 12px; }
.host-ip { color: var(--text-secondary); font-size: 12px; flex: 1; }
.list-actions { display: flex; gap: 6px; margin-left: auto; }
.slide-enter-active, .slide-leave-active { transition: all 0.2s ease; overflow: hidden; }
.slide-enter-from, .slide-leave-to { max-height: 0; opacity: 0; }
.slide-enter-to, .slide-leave-from { max-height: 2000px; opacity: 1; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/HostListView.vue
git commit -m "feat: add host list page with grouped view and card/list toggle"
```

---

### Task 8: WebSSH Terminal 页（Tab + 分屏）

**Files:**
- Create: `frontend/src/views/TerminalView.vue`

- [ ] **Step 1: 实现 Terminal 页**

```vue
<!-- frontend/src/views/TerminalView.vue -->
<template>
  <div class="terminal-page">
    <!-- Tab 栏 -->
    <div class="tab-bar">
      <div v-for="tab in tabs" :key="tab.id"
        :class="['tab', tab.id === activeTab && 'active']"
        @click="activeTab = tab.id">
        <div :class="['tab-dot', tab.connected ? 'online' : 'offline']" />
        <span>{{ tab.name }}</span>
        <button class="tab-close" @click.stop="closeTab(tab.id)">×</button>
      </div>
      <div class="tab-actions">
        <button class="split-btn" :class="{ active: splitMode }" @click="toggleSplit" title="分屏">⊞</button>
      </div>
    </div>

    <!-- Terminal 区域 -->
    <div :class="['terminal-area', splitMode && 'split']">
      <div v-for="(tab, i) in displayTabs" :key="tab.id"
        class="terminal-pane"
        :ref="el => setPaneRef(el as HTMLElement, i)">
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'

interface Tab {
  id: string
  name: string
  hostId: number
  connected: boolean
  term?: Terminal
  fit?: FitAddon
  ws?: WebSocket
}

const auth = useAuthStore()
const theme = useThemeStore()
const tabs = ref<Tab[]>([])
const activeTab = ref('')
const splitMode = ref(false)
const paneRefs: HTMLElement[] = []

// 分屏时最多显示 4 个，否则只显示当前 tab
const displayTabs = computed(() =>
  splitMode.value ? tabs.value.slice(0, 4) : tabs.value.filter(t => t.id === activeTab.value)
)

function setPaneRef(el: HTMLElement | null, i: number) {
  if (el) paneRefs[i] = el
}

onMounted(async () => {
  // 读取待连接的主机（从主机列表页传过来）
  const pending: { hostId: number; hostName: string }[] =
    JSON.parse(sessionStorage.getItem('pendingSSH') ?? '[]')
  sessionStorage.removeItem('pendingSSH')

  for (const p of pending) {
    await openTab(p.hostId, p.hostName)
  }
  if (tabs.value.length === 0) {
    // 无待连接主机时创建空 tab 提示
  }
})

async function openTab(hostId: number, name: string) {
  const id = `tab-${Date.now()}-${hostId}`
  const tab: Tab = { id, name, hostId, connected: false }
  tabs.value.push(tab)
  activeTab.value = id

  await nextTick()
  await initTerminal(tab)
}

async function initTerminal(tab: Tab) {
  const idx = displayTabs.value.findIndex(t => t.id === tab.id)
  const container = paneRefs[idx]
  if (!container) return

  const term = new Terminal({
    theme: { background: 'var(--terminal-bg)', foreground: 'var(--terminal-fg)' },
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    fontSize: 14,
    cursorBlink: true,
    allowTransparency: true,
  })
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(container)
  fit.fit()
  tab.term = term
  tab.fit = fit

  // 连接 WebSocket
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const ws = new WebSocket(`${proto}//${location.host}/api/ws/ssh/${tab.hostId}`)
  ws.binaryType = 'arraybuffer'
  tab.ws = ws

  ws.onopen = () => {
    // 发送初始尺寸
    ws.send(JSON.stringify({ type: 'init', width: term.cols, height: term.rows }))
    tab.connected = true
  }
  ws.onmessage = (e) => {
    const data = e.data instanceof ArrayBuffer
      ? new Uint8Array(e.data)
      : e.data
    term.write(data)
  }
  ws.onclose = () => { tab.connected = false; term.write('\r\n\x1b[31m[连接已断开]\x1b[0m\r\n') }

  term.onData(data => {
    if (ws.readyState === WebSocket.OPEN) ws.send(data)
  })
  term.onResize(({ cols, rows }) => {
    if (ws.readyState === WebSocket.OPEN)
      ws.send(JSON.stringify({ type: 'resize', width: cols, height: rows }))
  })

  // 监听容器大小变化
  const ro = new ResizeObserver(() => fit.fit())
  ro.observe(container)
}

function closeTab(id: string) {
  const tab = tabs.value.find(t => t.id === id)
  if (tab?.ws) tab.ws.close()
  if (tab?.term) tab.term.dispose()
  tabs.value = tabs.value.filter(t => t.id !== id)
  if (activeTab.value === id && tabs.value.length)
    activeTab.value = tabs.value[tabs.value.length - 1].id
}

function toggleSplit() {
  splitMode.value = !splitMode.value
  nextTick(() => tabs.value.forEach(t => t.fit?.fit()))
}

onUnmounted(() => {
  tabs.value.forEach(t => { t.ws?.close(); t.term?.dispose() })
})
</script>

<style scoped>
.terminal-page { display: flex; flex-direction: column; height: 100%; background: var(--terminal-bg); }
.tab-bar { display: flex; align-items: center; background: var(--bg-sidebar);
  border-bottom: 1px solid var(--border); height: 36px; padding: 0 8px; gap: 2px; overflow-x: auto; }
.tab { display: flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 4px 4px 0 0;
  cursor: pointer; font-size: 12px; color: var(--text-secondary); white-space: nowrap;
  border: 1px solid transparent; border-bottom: none; transition: all 0.15s; }
.tab.active { background: var(--terminal-bg); color: var(--text-primary);
  border-color: var(--border); }
.tab-dot { width: 6px; height: 6px; border-radius: 50%; }
.tab-dot.online { background: var(--success); }
.tab-dot.offline { background: var(--text-muted); }
.tab-close { background: none; border: none; color: inherit; cursor: pointer;
  padding: 0 2px; opacity: 0.5; font-size: 14px; line-height: 1; }
.tab-close:hover { opacity: 1; color: var(--danger); }
.tab-actions { margin-left: auto; display: flex; gap: 4px; }
.split-btn { background: transparent; border: 1px solid var(--border); color: var(--text-secondary);
  border-radius: 4px; padding: 2px 8px; cursor: pointer; font-size: 13px; }
.split-btn.active { border-color: var(--accent); color: var(--accent); }
.terminal-area { flex: 1; overflow: hidden; display: flex; gap: 2px; background: var(--border); }
.terminal-pane { flex: 1; overflow: hidden; background: var(--terminal-bg); }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/TerminalView.vue
git commit -m "feat: add WebSSH terminal page with tab and split screen"
```

---

### Task 9: 其余页面（SFTP / 审计 / 设置）

**Files:**
- Create: `frontend/src/views/SftpView.vue`
- Create: `frontend/src/views/AuditSessionsView.vue`
- Create: `frontend/src/views/AuditReplayView.vue`
- Create: `frontend/src/views/DangerRulesView.vue`
- Create: `frontend/src/views/ProfileView.vue`
- Create: `frontend/src/views/UsersView.vue`

- [ ] **Step 1: SFTP 文件管理页（核心功能）**

```vue
<!-- frontend/src/views/SftpView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">文件管理</h2>
      <span class="path-bar">{{ currentPath }}</span>
    </div>
    <div class="sftp-toolbar">
      <n-button size="small" :disabled="currentPath === '/'" @click="goUp">↑ 上级</n-button>
      <n-button size="small" type="primary" @click="triggerUpload">上传文件</n-button>
      <n-button size="small" @click="mkdir">新建目录</n-button>
      <input ref="fileInput" type="file" style="display:none" @change="uploadFile" />
    </div>
    <div class="file-list">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-for="file in files" :key="file.name" class="file-row"
        @dblclick="file.is_dir ? navigate(file.name) : download(file.name)">
        <span class="file-icon">{{ file.is_dir ? '📁' : '📄' }}</span>
        <span class="file-name">{{ file.name }}</span>
        <span class="file-size">{{ file.is_dir ? '--' : formatSize(file.size) }}</span>
        <div class="file-actions">
          <n-button v-if="!file.is_dir" size="tiny" @click.stop="download(file.name)">下载</n-button>
          <n-button size="tiny" type="error" @click.stop="deleteFile(file.name)">删除</n-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { NButton } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const hostId = route.params.hostId as string
const currentPath = ref('/')
const files = ref<any[]>([])
const loading = ref(false)
const fileInput = ref<HTMLInputElement>()
let ws: WebSocket | null = null

onMounted(() => {
  connect()
})

function connect() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${proto}//${location.host}/api/ws/sftp/${hostId}`)
  ws.onopen = () => listDir(currentPath.value)
  ws.onmessage = (e) => handleMessage(JSON.parse(e.data))
}

function send(msg: any) {
  if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify(msg))
}

function handleMessage(msg: any) {
  if (msg.type === 'result' && msg.files) {
    files.value = msg.files.sort((a: any, b: any) =>
      (b.is_dir ? 1 : 0) - (a.is_dir ? 1 : 0) || a.name.localeCompare(b.name))
    loading.value = false
  } else if (msg.type === 'file_data') {
    // 下载：等待二进制消息
    ws!.addEventListener('message', function handler(e: MessageEvent) {
      const blob = new Blob([e.data])
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url; a.download = msg.name; a.click()
      URL.revokeObjectURL(url)
      ws!.removeEventListener('message', handler)
    }, { once: false })
  }
}

function listDir(path: string) {
  loading.value = true
  send({ type: 'list', path, req_id: 'list' })
}

function navigate(name: string) {
  currentPath.value = `${currentPath.value}/${name}`.replace('//', '/')
  listDir(currentPath.value)
}

function goUp() {
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  currentPath.value = '/' + parts.join('/')
  listDir(currentPath.value)
}

function download(name: string) {
  send({ type: 'download', path: `${currentPath.value}/${name}`.replace('//', '/'), req_id: 'dl' })
}

function deleteFile(name: string) {
  if (!confirm(`确认删除 ${name}?`)) return
  send({ type: 'delete', path: `${currentPath.value}/${name}`.replace('//', '/'), req_id: 'del' })
  setTimeout(() => listDir(currentPath.value), 500)
}

function triggerUpload() { fileInput.value?.click() }

function uploadFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  const path = `${currentPath.value}/${file.name}`.replace('//', '/')
  send({ type: 'upload', path, name: file.name, size: file.size, req_id: 'up' })
  const reader = new FileReader()
  reader.onload = (ev) => {
    ws?.send(ev.target?.result as ArrayBuffer)
    setTimeout(() => listDir(currentPath.value), 500)
  }
  reader.readAsArrayBuffer(file)
}

function mkdir() {
  const name = prompt('目录名称')
  if (!name) return
  send({ type: 'mkdir', path: `${currentPath.value}/${name}`.replace('//', '/'), req_id: 'mkdir' })
  setTimeout(() => listDir(currentPath.value), 500)
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1048576) return `${(bytes/1024).toFixed(1)}K`
  return `${(bytes/1048576).toFixed(1)}M`
}

onUnmounted(() => ws?.close())
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; padding: 20px 24px; }
.page-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
.path-bar { color: var(--text-secondary); font-family: monospace; font-size: 13px;
  background: var(--bg-elevated); padding: 3px 10px; border-radius: 4px; }
.sftp-toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.file-list { flex: 1; overflow-y: auto; background: var(--bg-surface);
  border: 1px solid var(--border); border-radius: 8px; }
.loading { padding: 40px; text-align: center; color: var(--text-muted); }
.file-row { display: flex; align-items: center; gap: 10px; padding: 8px 14px;
  border-bottom: 1px solid var(--border); cursor: pointer; transition: background 0.1s; }
.file-row:last-child { border-bottom: none; }
.file-row:hover { background: var(--bg-elevated); }
.file-icon { font-size: 14px; }
.file-name { flex: 1; font-size: 13px; }
.file-size { color: var(--text-secondary); font-size: 12px; min-width: 60px; text-align: right; }
.file-actions { display: flex; gap: 6px; opacity: 0; transition: opacity 0.15s; }
.file-row:hover .file-actions { opacity: 1; }
</style>
```

- [ ] **Step 2: 审计会话列表页**

```vue
<!-- frontend/src/views/AuditSessionsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">审计日志</h2>
    </div>
    <n-data-table :columns="columns" :data="sessions" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage" />
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDataTable, NButton, NTag } from 'naive-ui'
import { sessionApi } from '@/api/session'

const router = useRouter()
const sessions = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '主机 ID', key: 'host_id', width: 100 },
  { title: '用户 ID', key: 'user_id', width: 100 },
  { title: '客户端 IP', key: 'client_ip', width: 140 },
  {
    title: '状态', key: 'status', width: 80,
    render: (row: any) => h(NTag, { type: row.status === 'active' ? 'success' : 'default', size: 'small' },
      { default: () => row.status })
  },
  { title: '开始时间', key: 'started_at', render: (row: any) => new Date(row.started_at).toLocaleString() },
  {
    title: '操作', key: 'actions',
    render: (row: any) => h('div', { style: 'display:flex;gap:6px' }, [
      h(NButton, { size: 'tiny', onClick: () => router.push(`/audit/replay/${row.id}`) }, { default: () => '回放' }),
    ])
  },
]

async function loadPage(page: number) {
  loading.value = true
  const res = await sessionApi.list({ page, page_size: 20 })
  sessions.value = res.data.data?.list ?? []
  pagination.value.itemCount = res.data.data?.total ?? 0
  pagination.value.page = page
  loading.value = false
}

onMounted(() => loadPage(1))
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
```

- [ ] **Step 3: 审计录像回放页**

```vue
<!-- frontend/src/views/AuditReplayView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <router-link to="/audit/sessions" class="back-link">← 返回</router-link>
      <h2 class="page-title">会话回放 #{{ sessionId }}</h2>
    </div>
    <div ref="playerContainer" class="player-container" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import * as AsciinemaPlayer from 'asciinema-player'
import 'asciinema-player/dist/bundle/asciinema-player.css'
import { sessionApi } from '@/api/session'

const route = useRoute()
const sessionId = route.params.id as string
const playerContainer = ref<HTMLElement>()

onMounted(() => {
  if (!playerContainer.value) return
  AsciinemaPlayer.create(
    sessionApi.getReplayUrl(Number(sessionId)),
    playerContainer.value,
    { cols: 220, rows: 50, fit: 'width', theme: 'monokai' }
  )
})
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; }
.back-link { color: var(--accent); text-decoration: none; font-size: 13px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
.player-container { max-width: 900px; }
</style>
```

- [ ] **Step 4: 高危规则管理页**

```vue
<!-- frontend/src/views/DangerRulesView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">高危指令规则</h2>
      <n-button type="primary" size="small" @click="showModal = true">+ 添加规则</n-button>
    </div>
    <n-data-table :columns="columns" :data="rules" :loading="loading" />

    <n-modal v-model:show="showModal" title="添加规则" preset="card" style="width:480px">
      <n-form :model="form">
        <n-form-item label="正则表达式" required>
          <n-input v-model:value="form.pattern" placeholder="rm\s+-rf\s+/" />
        </n-form-item>
        <n-form-item label="说明">
          <n-input v-model:value="form.description" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-button @click="showModal = false">取消</n-button>
        <n-button type="primary" :loading="saving" @click="save">保存</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NSwitch } from 'naive-ui'
import { dangerRuleApi } from '@/api/dangerrule'

const rules = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const saving = ref(false)
const form = ref({ pattern: '', description: '' })

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '正则表达式', key: 'pattern', render: (row: any) =>
    h('code', { style: 'font-family:monospace;color:var(--danger)' }, row.pattern) },
  { title: '说明', key: 'description' },
  { title: '动作', key: 'action', width: 80 },
  {
    title: '启用', key: 'enabled', width: 80,
    render: (row: any) => h(NSwitch, {
      value: row.enabled === 1,
      onUpdateValue: (v: boolean) => toggleRule(row.id, v)
    })
  },
  {
    title: '操作', key: 'actions', width: 80,
    render: (row: any) => h(NButton, {
      size: 'tiny', type: 'error',
      onClick: () => deleteRule(row.id)
    }, { default: () => '删除' })
  }
]

async function load() {
  loading.value = true
  const res = await dangerRuleApi.list({ page: 1, page_size: 100 })
  rules.value = res.data.data?.list ?? []
  loading.value = false
}

async function save() {
  saving.value = true
  await dangerRuleApi.create({ pattern: form.value.pattern, description: form.value.description })
  showModal.value = false
  form.value = { pattern: '', description: '' }
  await load()
  saving.value = false
}

async function deleteRule(id: number) {
  if (!confirm('确认删除？')) return
  await dangerRuleApi.delete(id)
  await load()
}

async function toggleRule(id: number, enabled: boolean) {
  await dangerRuleApi.update(id, { enabled: enabled ? 1 : 0 })
  await load()
}

onMounted(load)
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
```

- [ ] **Step 5: 个人设置页（主题切换）**

```vue
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
          <div class="theme-preview" :data-theme-preview="t.id">
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
```

- [ ] **Step 6: 用户管理页（简单列表）**

```vue
<!-- frontend/src/views/UsersView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
    </div>
    <n-empty description="用户管理功能开发中" style="margin-top:60px" />
  </div>
</template>

<script setup lang="ts">
import { NEmpty } from 'naive-ui'
</script>

<style scoped>
.page { height: 100%; padding: 20px 24px; }
.page-header { margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
```

- [ ] **Step 7: 全量构建验证**

```bash
cd frontend && npm run build
```

Expected: 无报错，输出到 `../web/`

- [ ] **Step 8: Commit + Push**

```bash
cd ..
git add frontend/src/views/
git commit -m "feat: add all frontend pages (sftp, audit, danger rules, profile)"
git push origin main
```

---

## Chunk 4: Go embed 集成

### Task 10: 将前端构建产物嵌入 Go 二进制

**Files:**
- Create: `web/embed.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 创建 embed.go**

```go
// web/embed.go
package web

import "embed"

//go:embed dist
var Static embed.FS
```

- [ ] **Step 2: 在 main.go 中注册静态文件服务**

```go
// 在 main.go 中，路由注册前添加：
import (
    "io/fs"
    "net/http"
    "github.com/sshive/sshive/web"
)

// 静态文件（前端 SPA）
distFS, err := fs.Sub(web.Static, "dist")
if err != nil {
    slog.Error("embed fs error", "err", err)
    os.Exit(1)
}
r.StaticFS("/assets", http.FS(distFS))

// SPA fallback：所有非 /api 路由返回 index.html
r.NoRoute(func(c *gin.Context) {
    if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
        middleware.NotFound(c, "not found")
        return
    }
    data, _ := fs.ReadFile(distFS, "index.html")
    c.Data(http.StatusOK, "text/html; charset=utf-8", data)
})
```

- [ ] **Step 3: 先构建前端，再编译 Go**

```bash
cd frontend && npm run build && cd ..
go build ./cmd/server/
```

Expected: 二进制包含前端，无报错

- [ ] **Step 4: 启动冒烟测试**

```bash
SSHIVE_DB_DSN="..." SSHIVE_REDIS_ADDR="..." SSHIVE_ENCRYPT_KEY="..." \
SSHIVE_JWT_SECRET="test" ./server &

# 访问前端
open http://localhost:8080
# 预期：看到登录页

kill %1
```

- [ ] **Step 5: 最终 Commit + Push**

```bash
git add web/ cmd/server/main.go frontend/
git commit -m "feat: embed frontend into go binary via go:embed"
git push origin main
```

---

## 测试汇总

```bash
cd frontend && npm run build   # SUCCESS（前端无编译错误）
cd .. && go build ./...        # SUCCESS（后端含嵌入前端）
```
