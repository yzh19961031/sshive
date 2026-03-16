<!-- frontend/src/views/HostGroupsView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">主机分组</h2>
      <n-button type="primary" size="small" @click="openAdd">+ 新建分组</n-button>
    </div>

    <n-data-table
      :columns="columns"
      :data="groups"
      :loading="loading"
      :pagination="false"
    />

    <!-- 新建/编辑弹框 -->
    <n-modal v-model:show="showModal" preset="card" :title="editId ? '编辑分组' : '新建分组'" style="width:360px" :mask-closable="false">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="70">
        <n-form-item label="分组名称" path="name">
          <n-input v-model:value="form.name" placeholder="例如：生产环境" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div style="display:flex;justify-content:flex-end;gap:8px">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="submit">确认</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, reactive } from 'vue'
import { NDataTable, NButton, NModal, NForm, NFormItem, NInput, NPopconfirm } from 'naive-ui'
import { hostGroupApi, type HostGroup } from '@/api/hostgroup'

const groups = ref<HostGroup[]>([])
const loading = ref(false)
const showModal = ref(false)
const saving = ref(false)
const editId = ref<number | null>(null)
const formRef = ref()
const form = reactive({ name: '' })

const rules = {
  name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }],
}

async function load() {
  loading.value = true
  try {
    const res = await hostGroupApi.list()
    groups.value = res.data.data ?? []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editId.value = null
  form.name = ''
  showModal.value = true
}

function openEdit(row: HostGroup) {
  editId.value = row.id
  form.name = row.name
  showModal.value = true
}

async function submit() {
  try { await formRef.value?.validate() } catch { return }
  saving.value = true
  try {
    if (editId.value) {
      await hostGroupApi.update(editId.value, { name: form.name })
    } else {
      await hostGroupApi.create({ name: form.name })
    }
    showModal.value = false
    await load()
  } finally {
    saving.value = false
  }
}

async function deleteGroup(id: number) {
  await hostGroupApi.delete(id)
  await load()
}

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '分组名称', key: 'name', ellipsis: { tooltip: true } },
  {
    title: '创建时间', key: 'created_at', width: 180,
    render: (row: HostGroup) => new Date(row.created_at).toLocaleString('zh-CN', { hour12: false }),
  },
  {
    title: '操作', key: 'actions', width: 130,
    render: (row: HostGroup) => h('div', { style: 'display:flex;gap:6px' }, [
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => openEdit(row) },
        { default: () => '编辑' }),
      h(NPopconfirm,
        { onPositiveClick: () => deleteGroup(row.id) },
        {
          trigger: () => h(NButton, { size: 'tiny', quaternary: true, type: 'error' }, { default: () => '删除' }),
          default: () => '确认删除该分组？主机不会被删除。',
        }
      ),
    ]),
  },
]

onMounted(load)
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
</style>
