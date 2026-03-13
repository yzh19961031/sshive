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
