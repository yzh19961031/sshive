// frontend/src/api/hostgroup.ts
import { http } from './client'

export interface HostGroup {
  id: number
  name: string
  tenant_id: number
  created_at: string
}

export const hostGroupApi = {
  list: () =>
    http.get<{ data: HostGroup[] }>('/host-groups'),
  create: (data: { name: string }) =>
    http.post<{ data: HostGroup }>('/host-groups', data),
  update: (id: number, data: { name: string }) =>
    http.put<{ data: HostGroup }>(`/host-groups/${id}`, data),
  delete: (id: number) => http.delete(`/host-groups/${id}`),
}
