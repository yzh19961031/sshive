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
        <n-button type="primary" size="small" @click="openAddModal">+ 添加主机</n-button>
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
            <div v-if="viewMode === 'card'" class="card-grid">
              <div v-for="host in group.hosts" :key="host.id" class="host-card">
                <div class="card-header">
                  <div :class="['status-dot', host.status === 1 ? 'online' : 'offline']" />
                  <span class="host-name">{{ host.name }}</span>
                  <span v-if="host.jump_host_id" style="margin-left:4px;font-size:10px;color:var(--text-secondary)">⤳跳板</span>
                </div>
                <div class="host-meta">{{ host.ip }}:{{ host.port }}</div>
                <div class="host-tags">
                  <span v-for="tag in host.tags" :key="tag" class="tag">{{ tag }}</span>
                </div>
                <div class="card-actions">
                  <n-button size="tiny" type="primary" @click="connectSSH(host)">SSH</n-button>
                  <n-button size="tiny" @click="connectSFTP(host)">SFTP</n-button>
                  <n-button size="tiny" @click="openEditModal(host)">编辑</n-button>
                  <n-button size="tiny" type="error" @click="deleteHost(host)">删除</n-button>
                </div>
              </div>
            </div>
            <div v-else class="host-list">
              <div v-for="host in group.hosts" :key="host.id" class="list-row">
                <div :class="['status-dot', host.status === 1 ? 'online' : 'offline']" />
                <span class="host-name">{{ host.name }}</span>
                <span v-if="host.jump_host_id" style="margin-left:4px;font-size:10px;color:var(--text-secondary)">⤳跳板</span>
                <span class="host-ip">{{ host.ip }}:{{ host.port }}</span>
                <div class="list-actions">
                  <n-button size="tiny" type="primary" @click="connectSSH(host)">SSH</n-button>
                  <n-button size="tiny" @click="connectSFTP(host)">SFTP</n-button>
                  <n-button size="tiny" @click="openEditModal(host)">编辑</n-button>
                  <n-button size="tiny" type="error" @click="deleteHost(host)">删除</n-button>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </div>
    <n-empty v-else description="暂无主机，点击右上角添加" style="margin-top: 60px" />

    <!-- Add Host Modal -->
    <n-modal v-model:show="showAddModal" preset="card" title="添加主机" style="width: 540px" :mask-closable="false">
      <n-form ref="addFormRef" :model="addForm" :rules="addRules" label-placement="left" label-width="80">
        <n-form-item label="主机名称" path="name">
          <n-input v-model:value="addForm.name" placeholder="例如：生产-Web-01" />
        </n-form-item>
        <n-form-item label="IP 地址" path="ip">
          <n-input v-model:value="addForm.ip" placeholder="192.168.1.100" />
        </n-form-item>
        <n-form-item label="端口" path="port">
          <n-input-number v-model:value="addForm.port" :min="1" :max="65535" style="width:100%" />
        </n-form-item>
        <n-form-item label="认证方式" path="auth_type">
          <n-select
            v-model:value="addForm.auth_type"
            :options="authTypeOptions"
            @update:value="onAuthTypeChange"
          />
        </n-form-item>

        <!-- 凭据选择 + 内联新建 -->
        <n-form-item label="凭据" path="credential_id">
          <div style="width:100%">
            <div style="display:flex;gap:8px;align-items:center">
              <n-select
                v-model:value="addForm.credential_id"
                :options="filteredCredentialOptions"
                :loading="loadingCredentials"
                placeholder="选择已有凭据"
                style="flex:1"
                clearable
              />
              <n-button size="small" @click="showNewCred = !showNewCred">
                {{ showNewCred ? '收起' : '+ 新建' }}
              </n-button>
            </div>

            <!-- 内联新建凭据表单 -->
            <transition name="slide">
              <div v-if="showNewCred" class="new-cred-box">
                <div class="new-cred-title">新建凭据</div>
                <n-form-item label="凭据名称" :show-feedback="false" style="margin-bottom:8px">
                  <n-input v-model:value="newCred.name" placeholder="例如：root-密码" size="small" />
                </n-form-item>
                <n-form-item label="用户名" :show-feedback="false" style="margin-bottom:8px">
                  <n-input v-model:value="newCred.username" placeholder="root" size="small" />
                </n-form-item>
                <n-form-item :label="addForm.auth_type === 'password' ? '密码' : '私钥'" :show-feedback="false" style="margin-bottom:8px">
                  <n-input
                    v-model:value="newCred.secret"
                    :type="addForm.auth_type === 'password' ? 'password' : 'textarea'"
                    :placeholder="addForm.auth_type === 'password' ? '登录密码' : '粘贴 PEM 私钥内容'"
                    show-password-on="click"
                    size="small"
                    :rows="addForm.auth_type === 'key' ? 4 : 1"
                  />
                </n-form-item>
                <n-button
                  type="primary" size="small" :loading="savingCred"
                  :disabled="!newCred.name || !newCred.username || !newCred.secret"
                  @click="createCredential"
                >
                  保存凭据并选中
                </n-button>
                <span v-if="credError" style="margin-left:8px;color:var(--danger);font-size:12px">{{ credError }}</span>
              </div>
            </transition>
          </div>
        </n-form-item>

        <n-form-item label="跳板机" path="jump_host_id">
          <n-select
            v-model:value="addForm.jump_host_id"
            :options="jumpHostOptions"
            placeholder="不使用跳板机"
            clearable
          />
        </n-form-item>

        <n-form-item label="分组" path="group_id">
          <div style="width:100%">
            <div style="display:flex;gap:8px;align-items:center">
              <n-select
                v-model:value="addForm.group_id"
                :options="groupOptions"
                placeholder="选择分组（可选）"
                clearable
                style="flex:1"
              />
              <n-button size="small" @click="showNewGroup = !showNewGroup">
                {{ showNewGroup ? '收起' : '+ 新建' }}
              </n-button>
            </div>
            <transition name="slide">
              <div v-if="showNewGroup" class="new-cred-box">
                <div class="new-cred-title">新建分组</div>
                <n-form-item label="分组名称" :show-feedback="false" style="margin-bottom:8px">
                  <n-input v-model:value="newGroupName" placeholder="例如：生产环境" size="small" />
                </n-form-item>
                <n-button
                  type="primary" size="small" :loading="savingGroup"
                  :disabled="!newGroupName"
                  @click="createGroup(addForm)"
                >
                  保存分组并选中
                </n-button>
                <span v-if="groupError" style="margin-left:8px;color:var(--danger);font-size:12px">{{ groupError }}</span>
              </div>
            </transition>
          </div>
        </n-form-item>

        <n-form-item label="标签" path="tags">
          <n-dynamic-tags v-model:value="addForm.tags" />
        </n-form-item>
      </n-form>
      <template #footer>
        <p v-if="hostError" style="color:var(--danger);font-size:12px;margin:0 0 8px">{{ hostError }}</p>
        <div style="display:flex;justify-content:flex-end;gap:8px">
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="submitAddHost">确认添加</n-button>
        </div>
      </template>
    </n-modal>

    <!-- Edit Host Modal -->
    <n-modal v-model:show="showEditModal" preset="card" title="编辑主机" style="width: 540px" :mask-closable="false">
      <n-form ref="editFormRef" :model="editForm" :rules="editRules" label-placement="left" label-width="80">
        <n-form-item label="主机名称" path="name">
          <n-input v-model:value="editForm.name" placeholder="例如：生产-Web-01" />
        </n-form-item>
        <n-form-item label="IP 地址" path="ip">
          <n-input v-model:value="editForm.ip" placeholder="192.168.1.100" />
        </n-form-item>
        <n-form-item label="端口" path="port">
          <n-input-number v-model:value="editForm.port" :min="1" :max="65535" style="width:100%" />
        </n-form-item>
        <n-form-item label="认证方式" path="auth_type">
          <n-select
            v-model:value="editForm.auth_type"
            :options="authTypeOptions"
            @update:value="onAuthTypeChangeEdit"
          />
        </n-form-item>

        <!-- 凭据选择 + 内联新建 -->
        <n-form-item label="凭据" path="credential_id">
          <div style="width:100%">
            <div style="display:flex;gap:8px;align-items:center">
              <n-select
                v-model:value="editForm.credential_id"
                :options="filteredCredentialOptionsEdit"
                :loading="loadingCredentials"
                placeholder="选择已有凭据"
                style="flex:1"
                clearable
              />
              <n-button size="small" @click="showNewCredEdit = !showNewCredEdit">
                {{ showNewCredEdit ? '收起' : '+ 新建' }}
              </n-button>
            </div>

            <!-- 内联新建凭据表单 -->
            <transition name="slide">
              <div v-if="showNewCredEdit" class="new-cred-box">
                <div class="new-cred-title">新建凭据</div>
                <n-form-item label="凭据名称" :show-feedback="false" style="margin-bottom:8px">
                  <n-input v-model:value="newCredEdit.name" placeholder="例如：root-密码" size="small" />
                </n-form-item>
                <n-form-item label="用户名" :show-feedback="false" style="margin-bottom:8px">
                  <n-input v-model:value="newCredEdit.username" placeholder="root" size="small" />
                </n-form-item>
                <n-form-item :label="editForm.auth_type === 'password' ? '密码' : '私钥'" :show-feedback="false" style="margin-bottom:8px">
                  <n-input
                    v-model:value="newCredEdit.secret"
                    :type="editForm.auth_type === 'password' ? 'password' : 'textarea'"
                    :placeholder="editForm.auth_type === 'password' ? '登录密码' : '粘贴 PEM 私钥内容'"
                    show-password-on="click"
                    size="small"
                    :rows="editForm.auth_type === 'key' ? 4 : 1"
                  />
                </n-form-item>
                <n-button
                  type="primary" size="small" :loading="savingCredEdit"
                  :disabled="!newCredEdit.name || !newCredEdit.username || !newCredEdit.secret"
                  @click="createCredentialForEdit"
                >
                  保存凭据并选中
                </n-button>
                <span v-if="credErrorEdit" style="margin-left:8px;color:var(--danger);font-size:12px">{{ credErrorEdit }}</span>
              </div>
            </transition>
          </div>
        </n-form-item>

        <n-form-item label="跳板机" path="jump_host_id">
          <n-select
            v-model:value="editForm.jump_host_id"
            :options="jumpHostOptionsForEdit"
            placeholder="不使用跳板机"
            clearable
          />
        </n-form-item>

        <n-form-item label="分组" path="group_id">
          <div style="width:100%">
            <div style="display:flex;gap:8px;align-items:center">
              <n-select
                v-model:value="editForm.group_id"
                :options="groupOptions"
                placeholder="选择分组（可选）"
                clearable
                style="flex:1"
              />
              <n-button size="small" @click="showNewGroup = !showNewGroup">
                {{ showNewGroup ? '收起' : '+ 新建' }}
              </n-button>
            </div>
            <transition name="slide">
              <div v-if="showNewGroup" class="new-cred-box">
                <div class="new-cred-title">新建分组</div>
                <n-form-item label="分组名称" :show-feedback="false" style="margin-bottom:8px">
                  <n-input v-model:value="newGroupName" placeholder="例如：生产环境" size="small" />
                </n-form-item>
                <n-button
                  type="primary" size="small" :loading="savingGroup"
                  :disabled="!newGroupName"
                  @click="createGroup(editForm)"
                >
                  保存分组并选中
                </n-button>
                <span v-if="groupError" style="margin-left:8px;color:var(--danger);font-size:12px">{{ groupError }}</span>
              </div>
            </transition>
          </div>
        </n-form-item>

        <n-form-item label="标签" path="tags">
          <n-dynamic-tags v-model:value="editForm.tags" />
        </n-form-item>
      </n-form>
      <template #footer>
        <p v-if="editHostError" style="color:var(--danger);font-size:12px;margin:0 0 8px">{{ editHostError }}</p>
        <div style="display:flex;justify-content:flex-end;gap:8px">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" :loading="editSaving" @click="submitEditHost">保存修改</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton, NInput, NEmpty, NModal, NForm, NFormItem,
  NSelect, NInputNumber, NDynamicTags,
} from 'naive-ui'
import { hostApi, type Host } from '@/api/host'
import { credentialApi, type Credential } from '@/api/credential'
import { hostGroupApi, type HostGroup } from '@/api/hostgroup'

const router = useRouter()
const hosts = ref<Host[]>([])
const groups = ref<HostGroup[]>([])
const search = ref('')
const viewMode = ref<'card' | 'list'>('card')
const showAddModal = ref(false)
const collapsed = reactive<Record<string, boolean>>({})

// ── Credentials ──────────────────────────────────────────
const credentials = ref<Credential[]>([])
const loadingCredentials = ref(false)

// 根据当前选择的 auth_type 过滤凭据
const filteredCredentialOptions = computed(() =>
  credentials.value
    .filter(c => c.type === addForm.auth_type)
    .map(c => ({
      label: `${c.name} (${c.username})`,
      value: c.id,
    }))
)

const authTypeOptions = [
  { label: '密码认证', value: 'password' },
  { label: '密钥认证', value: 'key' },
]

// ── 内联新建凭据 ──────────────────────────────────────────
const showNewCred = ref(false)
const savingCred = ref(false)
const credError = ref('')
const newCred = reactive({ name: '', username: '', secret: '' })

async function createCredential() {
  if (!newCred.name || !newCred.username || !newCred.secret) return
  credError.value = ''
  savingCred.value = true
  try {
    const res = await credentialApi.create({
      name: newCred.name,
      type: addForm.auth_type,
      username: newCred.username,
      secret: newCred.secret,
    })
    const created = res.data.data
    credentials.value.push(created)
    addForm.credential_id = created.id
    showNewCred.value = false
    newCred.name = ''
    newCred.username = ''
    newCred.secret = ''
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    credError.value = err.response?.data?.message ?? '创建失败'
  } finally {
    savingCred.value = false
  }
}

// ── Host form ─────────────────────────────────────────────
const addFormRef = ref()
const groupOptions = computed(() =>
  groups.value.map(g => ({ label: g.name, value: g.id }))
)

const jumpHostOptions = computed(() =>
  hosts.value.map(h => ({ label: `${h.name} (${h.ip})`, value: h.id }))
)

const addForm = reactive({
  name: '',
  ip: '',
  port: 22,
  auth_type: 'password' as 'password' | 'key',
  credential_id: null as number | null,
  tags: [] as string[],
  group_id: null as number | null,
  jump_host_id: undefined as number | undefined,
})
const saving = ref(false)
const hostError = ref('')

const addRules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [
    { required: true, message: '请输入 IP 地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入有效的 IP 地址', trigger: 'blur' },
  ],
  port: [{ required: true, type: 'number' as const, message: '请输入端口' }],
  auth_type: [{ required: true, message: '请选择认证方式' }],
  credential_id: [{ required: true, type: 'number' as const, message: '请选择或新建凭据' }],
}

function onAuthTypeChange() {
  addForm.credential_id = null
  showNewCred.value = false
}

// ── Inline Group Creation ──────────────────────────────────
const showNewGroup = ref(false)
const newGroupName = ref('')
const savingGroup = ref(false)
const groupError = ref('')

async function createGroup(form: { group_id: number | null }) {
  if (!newGroupName.value) return
  groupError.value = ''
  savingGroup.value = true
  try {
    const res = await hostGroupApi.create({ name: newGroupName.value })
    const created = res.data.data
    groups.value.push(created)
    form.group_id = created.id
    showNewGroup.value = false
    newGroupName.value = ''
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    groupError.value = err.response?.data?.message ?? '创建失败'
  } finally {
    savingGroup.value = false
  }
}

// ── Edit Host ──────────────────────────────────────────────
const showEditModal = ref(false)
const editFormRef = ref()
const editingHostId = ref<number>(0)
const editSaving = ref(false)
const editHostError = ref('')
const showNewCredEdit = ref(false)
const savingCredEdit = ref(false)
const credErrorEdit = ref('')
const newCredEdit = reactive({ name: '', username: '', secret: '' })

const editForm = reactive({
  name: '',
  ip: '',
  port: 22,
  auth_type: 'password' as 'password' | 'key',
  credential_id: null as number | null,
  tags: [] as string[],
  group_id: null as number | null,
  jump_host_id: undefined as number | undefined,
})

const editRules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [
    { required: true, message: '请输入 IP 地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入有效的 IP 地址', trigger: 'blur' },
  ],
  port: [{ required: true, type: 'number' as const, message: '请输入端口' }],
  auth_type: [{ required: true, message: '请选择认证方式' }],
  credential_id: [{ required: true, type: 'number' as const, message: '请选择或新建凭据' }],
}

const jumpHostOptionsForEdit = computed(() =>
  hosts.value
    .filter(h => h.id !== editingHostId.value)
    .map(h => ({ label: `${h.name} (${h.ip})`, value: h.id }))
)

// 根据 editForm.auth_type 过滤凭据
const filteredCredentialOptionsEdit = computed(() =>
  credentials.value
    .filter(c => c.type === editForm.auth_type)
    .map(c => ({ label: `${c.name} (${c.username})`, value: c.id }))
)

function onAuthTypeChangeEdit() {
  editForm.credential_id = null
  showNewCredEdit.value = false
}

async function createCredentialForEdit() {
  if (!newCredEdit.name || !newCredEdit.username || !newCredEdit.secret) return
  credErrorEdit.value = ''
  savingCredEdit.value = true
  try {
    const res = await credentialApi.create({
      name: newCredEdit.name,
      type: editForm.auth_type,
      username: newCredEdit.username,
      secret: newCredEdit.secret,
    })
    const created = res.data.data
    credentials.value.push(created)
    editForm.credential_id = created.id
    showNewCredEdit.value = false
    newCredEdit.name = ''
    newCredEdit.username = ''
    newCredEdit.secret = ''
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    credErrorEdit.value = err.response?.data?.message ?? '创建失败'
  } finally {
    savingCredEdit.value = false
  }
}

function openEditModal(host: Host) {
  editingHostId.value = host.id
  editForm.name = host.name
  editForm.ip = host.ip
  editForm.port = host.port
  editForm.auth_type = host.auth_type as 'password' | 'key'
  editForm.credential_id = host.credential_id
  editForm.tags = [...(host.tags ?? [])]
  editForm.group_id = host.group_id
  editForm.jump_host_id = host.jump_host_id
  editHostError.value = ''
  showNewCredEdit.value = false
  showNewGroup.value = false
  newCredEdit.name = ''
  newCredEdit.username = ''
  newCredEdit.secret = ''
  showEditModal.value = true
}

async function submitEditHost() {
  try {
    await editFormRef.value?.validate()
  } catch {
    return
  }
  if (!editForm.credential_id) {
    editHostError.value = '请选择或新建凭据'
    return
  }
  editSaving.value = true
  editHostError.value = ''
  try {
    await hostApi.update(editingHostId.value, {
      name: editForm.name,
      ip: editForm.ip,
      port: editForm.port,
      auth_type: editForm.auth_type,
      credential_id: editForm.credential_id,
      tags: editForm.tags,
      group_id: editForm.group_id,
      jump_host_id: editForm.jump_host_id,
    })
    showEditModal.value = false
    await loadHosts()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    editHostError.value = err.response?.data?.message ?? '保存失败'
  } finally {
    editSaving.value = false
  }
}

async function loadHosts() {
  const res = await hostApi.list({ page: 1, page_size: 100 })
  hosts.value = res.data.data?.list ?? []
}

async function loadGroups() {
  const res = await hostGroupApi.list()
  groups.value = res.data.data ?? []
}

async function loadCredentials() {
  loadingCredentials.value = true
  try {
    const res = await credentialApi.list({ page: 1, page_size: 100 })
    credentials.value = res.data.data?.list ?? []
  } finally {
    loadingCredentials.value = false
  }
}

onMounted(() => {
  loadHosts()
  loadCredentials()
  loadGroups()
})

function openAddModal() {
  addForm.name = ''
  addForm.ip = ''
  addForm.port = 22
  addForm.auth_type = 'password'
  addForm.credential_id = null
  addForm.tags = []
  addForm.group_id = null
  addForm.jump_host_id = undefined
  hostError.value = ''
  showNewCred.value = false
  showNewGroup.value = false
  newCred.name = ''
  newCred.username = ''
  newCred.secret = ''
  showAddModal.value = true
}

async function submitAddHost() {
  try {
    await addFormRef.value?.validate()
  } catch {
    return
  }
  if (!addForm.credential_id) {
    hostError.value = '请选择或新建凭据'
    return
  }
  saving.value = true
  hostError.value = ''
  try {
    await hostApi.create({
      name: addForm.name,
      ip: addForm.ip,
      port: addForm.port,
      auth_type: addForm.auth_type,
      credential_id: addForm.credential_id,
      status: 1,
      tags: addForm.tags,
      group_id: addForm.group_id,
      jump_host_id: addForm.jump_host_id,
    })
    showAddModal.value = false
    await loadHosts()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    hostError.value = err.response?.data?.message ?? '添加失败'
  } finally {
    saving.value = false
  }
}

async function deleteHost(host: Host) {
  if (!confirm(`确认删除主机「${host.name}」？`)) return
  try {
    await hostApi.delete(host.id)
    await loadHosts()
  } catch {
    hostError.value = '删除失败'
  }
}

const filteredHosts = computed(() =>
  hosts.value.filter(h =>
    !search.value ||
    h.name.includes(search.value) ||
    h.ip.includes(search.value)
  )
)

const groupedHosts = computed(() => {
  const map = new Map<string, Host[]>()
  const groupMap = new Map<number, string>(groups.value.map(g => [g.id, g.name]))
  for (const h of filteredHosts.value) {
    const label = h.group_id ? (groupMap.get(h.group_id) ?? '未分组') : '未分组'
    if (!map.has(label)) map.set(label, [])
    map.get(label)!.push(h)
  }
  return [...map.entries()].map(([label, hs]) => ({ label, hosts: hs }))
})

function toggleGroup(label: string) {
  collapsed[label] = !collapsed[label]
}

function connectSSH(host: Host) {
  const sessions: { hostId: number; hostName: string }[] = JSON.parse(sessionStorage.getItem('pendingSSH') ?? '[]')
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
.new-cred-box {
  margin-top: 10px;
  padding: 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 6px;
}
.new-cred-title {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  font-weight: 500;
}
.slide-enter-active, .slide-leave-active { transition: all 0.2s ease; overflow: hidden; }
.slide-enter-from, .slide-leave-to { max-height: 0; opacity: 0; }
.slide-enter-to, .slide-leave-from { max-height: 2000px; opacity: 1; }
</style>
