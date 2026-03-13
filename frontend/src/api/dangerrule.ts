// frontend/src/api/dangerrule.ts
import { http } from './client'

export const dangerRuleApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get('/danger-rules', { params }),
  create: (data: { pattern: string; description?: string; action?: string }) =>
    http.post('/danger-rules', data),
  update: (id: number, data: { pattern?: string; description?: string; enabled?: number }) =>
    http.put(`/danger-rules/${id}`, data),
  delete: (id: number) => http.delete(`/danger-rules/${id}`),
}
