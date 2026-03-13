// frontend/src/api/stats.ts
import { http } from './client'

export interface StatsData {
  total_hosts: number
  today_sessions: number
  active_sessions: number
  today_danger_events: number
}

export const statsApi = {
  get: () => http.get<{ code: number; data: StatsData }>('/stats'),
}
