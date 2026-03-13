// frontend/src/api/auth.ts
import { http } from './client'

export interface TenantOption {
  id: number
  name: string
}

export const authApi = {
  login: (data: { username: string; password: string; tenant_id: number }) =>
    http.post('/auth/login', data),
  logout: () => http.post('/auth/logout'),
  listTenantsPublic: () =>
    http.get<{ data: TenantOption[] }>('/tenants/public'),
}
