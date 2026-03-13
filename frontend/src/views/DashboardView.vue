<!-- frontend/src/views/DashboardView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">仪表盘</h2>
    </div>

    <!-- Stat cards -->
    <div class="stat-grid">
      <div class="stat-card" v-for="card in statCards" :key="card.label">
        <div class="stat-icon">{{ card.icon }}</div>
        <div class="stat-info">
          <div class="stat-value">{{ card.value }}</div>
          <div class="stat-label">{{ card.label }}</div>
        </div>
      </div>
    </div>

    <!-- Bottom two panels -->
    <div class="bottom-grid">
      <!-- Recent sessions -->
      <div class="panel">
        <div class="panel-title">最近会话</div>
        <n-data-table :columns="sessionColumns" :data="recentSessions"
          :loading="sessionsLoading" :pagination="false" size="small" />
      </div>

      <!-- Quick connect -->
      <div class="panel">
        <div class="panel-title">快捷连接</div>
        <div v-if="quickHosts.length === 0" class="empty-tip">暂无最近连接</div>
        <div v-else class="quick-grid">
          <div v-for="h in quickHosts" :key="h.host_id"
            class="quick-card" @click="openTerminal(h.host_id)">
            <span class="quick-icon">⬡</span>
            <span class="quick-name">{{ h.host_name || h.host_id }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDataTable, NTag } from 'naive-ui'
import { statsApi } from '@/api/stats'
import type { StatsData } from '@/api/stats'
import { sessionApi } from '@/api/session'

const router = useRouter()

// Stats
const stats = ref<StatsData>({ total_hosts: 0, today_sessions: 0, active_sessions: 0, today_danger_events: 0 })
const statCards = computed(() => [
  { icon: '⬡', label: '主机总数',    value: stats.value.total_hosts },
  { icon: '💻', label: '今日会话',    value: stats.value.today_sessions },
  { icon: '🔗', label: '活跃会话',    value: stats.value.active_sessions },
  { icon: '⚠', label: '今日高危拦截', value: stats.value.today_danger_events },
])

// Recent sessions
const recentSessions = ref<any[]>([])
const sessionsLoading = ref(false)
const sessionColumns = [
  { title: '主机', key: 'host_name', render: (row: any) => row.host_name || String(row.host_id) },
  { title: '用户', key: 'username',  render: (row: any) => row.username || String(row.user_id) },
  {
    title: '状态', key: 'status', width: 70,
    render: (row: any) => h(NTag, { type: row.status === 'active' ? 'success' : 'default', size: 'small' },
      { default: () => row.status })
  },
  { title: '开始时间', key: 'started_at', render: (row: any) => new Date(row.started_at).toLocaleString() },
]

// Quick connect hosts (deduplicated from recent sessions)
const quickHosts = computed(() => {
  const seen = new Set<number>()
  const result: any[] = []
  for (const s of recentSessions.value) {
    if (!seen.has(s.host_id)) {
      seen.add(s.host_id)
      result.push(s)
    }
    if (result.length >= 5) break
  }
  return result
})

function openTerminal(hostId: number) {
  router.push({ path: '/terminal', query: { hostId: String(hostId) } })
}

async function loadData() {
  sessionsLoading.value = true
  try {
    const [statsRes, sessRes] = await Promise.all([
      statsApi.get(),
      sessionApi.list({ page: 1, page_size: 5 }),
    ])
    stats.value = statsRes.data.data ?? stats.value
    recentSessions.value = sessRes.data.data?.list ?? []
  } catch (e) {
    console.error('Dashboard data load failed:', e)
  } finally {
    sessionsLoading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { margin-bottom: 20px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}
.stat-card {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 18px;
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon { font-size: 28px; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }

.bottom-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.panel {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
}
.panel-title { font-size: 13px; font-weight: 600; margin-bottom: 12px; color: var(--text-secondary); }

.quick-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; }
.quick-card {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: border-color 0.15s, color 0.15s;
}
.quick-card:hover { border-color: var(--accent); color: var(--accent); }
.quick-icon { font-size: 18px; }
.quick-name { font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.empty-tip { color: var(--text-secondary); text-align: center; padding: 20px 0; font-size: 13px; }
</style>
