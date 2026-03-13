// frontend/src/api/credential.ts
import { http } from './client'

export interface Credential {
  id: number
  name: string
  type: 'password' | 'key'
  username: string
  created_at: string
}

export const credentialApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get<{ data: { total: number; list: Credential[] } }>('/credentials', { params }),
  create: (data: { name: string; type: 'password' | 'key'; username: string; secret: string }) =>
    http.post<{ data: Credential }>('/credentials', data),
}
