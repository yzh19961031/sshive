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
