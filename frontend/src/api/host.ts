// frontend/src/api/host.ts
import { http } from './client'

export interface Host {
  id: number
  name: string
  ip: string
  port: number
  auth_type: string
  credential_id: number
  status: number
  tags: string[]
  created_at: string
}

export const hostApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get<{ data: { total: number; list: Host[] } }>('/hosts', { params }),
  create: (data: Omit<Host, 'id' | 'created_at'>) =>
    http.post<{ data: Host }>('/hosts', data),
  update: (id: number, data: Partial<Host>) =>
    http.put<{ data: Host }>(`/hosts/${id}`, data),
  delete: (id: number) => http.delete(`/hosts/${id}`),
}
