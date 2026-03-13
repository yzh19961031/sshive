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

const auth = useAuthStore()

const route = useRoute()
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
  ws = new WebSocket(`${proto}//${location.host}/api/ws/sftp/${hostId}?token=${auth.token}`)
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
