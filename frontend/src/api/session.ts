// frontend/src/api/session.ts
import { http } from './client'

export interface LogItem {
  id: number
  session_id: number
  type: 'input' | 'output'
  content: string
  created_at: string
}

export const sessionApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/sessions', { params }),
  getLogs: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/logs`, { params }),
  getCommands: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/commands`, { params }),
  getReplayUrl: (id: number) => {
    const token = localStorage.getItem('token') ?? ''
    return `/api/sessions/${id}/replay?token=${encodeURIComponent(token)}`
  },
}
