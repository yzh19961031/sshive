// frontend/src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: () => import('@/views/DashboardView.vue') },
      { path: 'hosts', component: () => import('@/views/HostListView.vue') },
      { path: 'host-groups', component: () => import('@/views/HostGroupsView.vue') },
      { path: 'terminal', component: () => import('@/views/TerminalView.vue') },
      { path: 'sftp/:hostId', component: () => import('@/views/SftpView.vue') },
      { path: 'audit/sessions', component: () => import('@/views/AuditSessionsView.vue') },
      { path: 'audit/replay/:id', component: () => import('@/views/AuditReplayView.vue') },
      { path: 'settings/rules', component: () => import('@/views/DangerRulesView.vue') },
      { path: 'settings/users', component: () => import('@/views/UsersView.vue') },
      { path: 'settings/profile', component: () => import('@/views/ProfileView.vue') },
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Route guard: redirect to login if not authenticated
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
})
