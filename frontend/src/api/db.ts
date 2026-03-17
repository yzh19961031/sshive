// frontend/src/api/db.ts
import { http } from './client'

export interface DBServer {
  id: number
  name: string
  type: 'mysql' | 'postgres'
  host: string
  port: number
  username: string
  database: string
}

export interface QueryResult {
  columns: string[]
  rows: any[][]
  total: number
}

export const dbApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    http.get<{ data: { total: number; list: DBServer[] } }>('/db-servers', { params }),
  create: (data: Omit<DBServer, 'id'> & { password: string }) =>
    http.post<{ data: DBServer }>('/db-servers', data),
  update: (id: number, data: Partial<Omit<DBServer, 'id'> & { password?: string }>) =>
    http.put<{ data: DBServer }>(`/db-servers/${id}`, data),
  delete: (id: number) => http.delete(`/db-servers/${id}`),
  query: (id: number, sql: string) =>
    http.post<{ data: QueryResult }>(`/db-servers/${id}/query`, { sql }),
  databases: (id: number) =>
    http.get<{ data: string[] }>(`/db-servers/${id}/databases`),
  tables: (id: number, db: string) =>
    http.get<{ data: string[] }>(`/db-servers/${id}/databases/${db}/tables`),
}
