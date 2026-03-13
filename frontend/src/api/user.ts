// frontend/src/api/user.ts
import { http } from './client'

export interface UserItem {
  id: number
  username: string
  email: string
  status: number
  created_at: string
}

export interface RoleItem {
  id: number
  name: string
  description: string
}

export const userApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/users', { params }),
  create: (data: { username: string; password: string; email?: string }) =>
    http.post('/users', data),
  update: (id: number, data: { email?: string; status?: number }) =>
    http.put(`/users/${id}`, data),
  disable: (id: number) =>
    http.delete(`/users/${id}`),
  assignRoles: (id: number, role_ids: number[]) =>
    http.put(`/users/${id}/roles`, { role_ids }),
  listRoles: () =>
    http.get('/roles'),
}
