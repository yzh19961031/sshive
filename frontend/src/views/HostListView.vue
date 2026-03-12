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
            <!-- Card view -->
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
            <!-- List view -->
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

// Group by first tag (no tag → "未分组")
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
