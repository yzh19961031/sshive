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
        <n-form-item label="租户" path="tenantId">
          <n-select
            v-model:value="form.tenantId"
            :options="tenantOptions"
            :loading="loadingTenants"
            placeholder="请选择租户"
            value-field="value"
            label-field="label"
          />
        </n-form-item>
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="admin" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="••••••••"
            show-password-on="click"
            @keydown.enter="handleLogin"
          />
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
import { ref, onMounted } from 'vue'
import { NForm, NFormItem, NInput, NSelect, NButton } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'

const auth = useAuthStore()
const form = ref({ tenantId: null as number | null, username: '', password: '' })
const loading = ref(false)
const loadingTenants = ref(false)
const error = ref('')

interface SelectOption { label: string; value: number }
const tenantOptions = ref<SelectOption[]>([])

onMounted(async () => {
  loadingTenants.value = true
  try {
    const res = await authApi.listTenantsPublic()
    const list: { id: number; name: string }[] = res.data.data ?? []
    tenantOptions.value = list.map(t => ({ label: t.name, value: t.id }))
    if (tenantOptions.value.length === 1 && tenantOptions.value[0]) {
      form.value.tenantId = tenantOptions.value[0].value
    }
  } catch {
    // ignore, user can retry
  } finally {
    loadingTenants.value = false
  }
})

const rules = {
  tenantId: [{ required: true, type: 'number' as const, message: '请选择租户' }],
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }],
}

async function handleLogin() {
  if (!form.value.tenantId) {
    error.value = '请选择租户'
    return
  }
  error.value = ''
  loading.value = true
  try {
    await auth.login(form.value.username, form.value.password, form.value.tenantId)
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    error.value = err.response?.data?.message ?? '登录失败'
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
