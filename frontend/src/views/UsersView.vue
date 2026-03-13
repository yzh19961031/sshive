<!-- frontend/src/views/UsersView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
      <n-button type="primary" size="small" @click="openCreate">+ 新建用户</n-button>
    </div>

    <n-data-table :columns="columns" :data="users" :loading="loading"
      :pagination="pagination" remote @update:page="loadPage" />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="modalVisible" preset="card" :title="editingId ? '编辑用户' : '新建用户'"
      style="width:440px" :mask-closable="false">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80px">
        <n-form-item v-if="!editingId" label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="2-50字符" />
        </n-form-item>
        <n-form-item v-if="!editingId" label="密码" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="至少8位" show-password-on="click" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="可选" />
        </n-form-item>
        <n-form-item v-if="editingId" label="状态" path="status">
          <n-select v-model:value="form.status" :options="statusOptions" />
        </n-form-item>
        <n-form-item label="角色" path="role_ids">
          <n-select v-model:value="form.role_ids" multiple :options="roleOptions"
            placeholder="选择角色" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div style="display:flex;justify-content:flex-end;gap:8px">
          <n-button @click="modalVisible = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NDataTable, NButton, NTag, NModal, NForm, NFormItem, NInput, NSelect } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { userApi } from '@/api/user'

const users = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

// Modal state
const modalVisible = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formRef = ref<FormInst>()
const form = ref({ username: '', password: '', email: '', status: 1, role_ids: [] as number[] })

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 2 },
]

const roleOptions = ref<{ label: string; value: number }[]>([])

const rules: FormRules = {
  username: [{ required: true, min: 2, message: '用户名至少2位', trigger: 'blur' }],
  password: [{ required: true, min: 8, message: '密码至少8位', trigger: 'blur' }],
}

function openCreate() {
  editingId.value = null
  form.value = { username: '', password: '', email: '', status: 1, role_ids: [] }
  modalVisible.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  form.value = { username: row.username, password: '', email: row.email ?? '', status: row.status, role_ids: [] }
  modalVisible.value = true
}

async function save() {
  try { await formRef.value?.validate() } catch { return }
  saving.value = true
  try {
    if (editingId.value) {
      await userApi.update(editingId.value, { email: form.value.email, status: form.value.status })
      if (form.value.role_ids.length > 0) {
        await userApi.assignRoles(editingId.value, form.value.role_ids)
      }
    } else {
      const res = await userApi.create({ username: form.value.username, password: form.value.password, email: form.value.email })
      const newId = res.data.data?.id
      if (newId && form.value.role_ids.length > 0) {
        await userApi.assignRoles(newId, form.value.role_ids)
      }
    }
    modalVisible.value = false
    loadPage(pagination.value.page)
  } finally {
    saving.value = false
  }
}

async function disableUser(id: number) {
  await userApi.disable(id)
  loadPage(pagination.value.page)
}

const columns = [
  { title: '用户名', key: 'username' },
  { title: '邮箱', key: 'email', render: (row: any) => row.email || '—' },
  {
    title: '状态', key: 'status', width: 80,
    render: (row: any) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' },
      { default: () => row.status === 1 ? '启用' : '禁用' })
  },
  { title: '创建时间', key: 'created_at', render: (row: any) => new Date(row.created_at).toLocaleString() },
  {
    title: '操作', key: 'actions',
    render: (row: any) => h('div', { style: 'display:flex;gap:6px' }, [
      h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => '编辑' }),
      row.status === 1
        ? h(NButton, { size: 'tiny', type: 'error', onClick: () => disableUser(row.id) }, { default: () => '禁用' })
        : null,
    ])
  },
]

async function loadPage(page: number) {
  loading.value = true
  try {
    const res = await userApi.list({ page, page_size: 20 })
    users.value = res.data.data?.list ?? []
    pagination.value.itemCount = res.data.data?.total ?? 0
    pagination.value.page = page
  } finally {
    loading.value = false
  }
}

async function loadRoles() {
  try {
    const res = await userApi.listRoles()
    const list: any[] = res.data.data ?? []
    roleOptions.value = list.map(r => ({ label: r.description || r.name, value: r.id }))
  } catch (e) {
    console.error('Failed to load roles:', e)
  }
}

onMounted(() => {
  loadPage(1)
  loadRoles()
})
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
