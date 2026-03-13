<!-- frontend/src/views/AuditReplayView.vue -->
<template>
  <div class="page">
    <div class="page-header">
      <router-link to="/audit/sessions" class="back-link">← 返回</router-link>
      <h2 class="page-title">会话回放 #{{ sessionId }}</h2>
    </div>
    <div ref="playerContainer" class="player-container" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
// @ts-ignore
import * as AsciinemaPlayer from 'asciinema-player'
import 'asciinema-player/dist/bundle/asciinema-player.css'
import { sessionApi } from '@/api/session'

const route = useRoute()
const sessionId = route.params.id as string
const playerContainer = ref<HTMLElement>()

onMounted(() => {
  if (!playerContainer.value) return
  AsciinemaPlayer.create(
    sessionApi.getReplayUrl(Number(sessionId)),
    playerContainer.value,
    { cols: 220, rows: 50, fit: 'width', theme: 'monokai' }
  )
})
</script>

<style scoped>
.page { height: 100%; overflow-y: auto; padding: 20px 24px; }
.page-header { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; }
.back-link { color: var(--accent); text-decoration: none; font-size: 13px; }
.page-title { margin: 0; font-size: 16px; font-weight: 600; }
.player-container { max-width: 900px; }
</style>
