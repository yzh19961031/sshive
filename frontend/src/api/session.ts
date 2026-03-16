// frontend/src/api/session.ts
import { http } from './client'

export interface LogItem {
  id: number
  session_id: number
  type: 'input' | 'output'
  content: string
  created_at: string
}

export interface CommandItem {
  id: number
  session_id: number
  command: string
  action: 'execute' | 'blocked'
  result: string
  created_at: string
  host_name: string
  username: string
}

export const sessionApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/sessions', { params }),
  getLogs: (id: number, params?: { page?: number; page_size?: number }) =>
    http.get(`/sessions/${id}/logs`, { params }),
  getCommands: (id: number, params?: {
    page?: number; page_size?: number
    command?: string; action?: string
    start_time?: string; end_time?: string
  }) => http.get(`/sessions/${id}/commands`, { params }),
  getReplayUrl: (id: number) => {
    const token = localStorage.getItem('token') ?? ''
    return `/api/sessions/${id}/replay?token=${encodeURIComponent(token)}`
  },
  listAllCommands: (params?: { page?: number; page_size?: number }) =>
    http.get('/commands', { params }),
}
