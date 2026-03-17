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
  list: () => http.get<{ code: number; data: DBServer[] }>('/api/db-servers'),
  create: (data: Omit<DBServer, 'id'> & { password: string }) =>
    http.post<{ data: DBServer }>('/api/db-servers', data),
  delete: (id: number) => http.delete(`/api/db-servers/${id}`),
  query: (id: number, sql: string) =>
    http.post<{ data: QueryResult }>(`/api/db-servers/${id}/query`, { sql }),
  databases: (id: number) =>
    http.get<{ data: string[] }>(`/api/db-servers/${id}/databases`),
  tables: (id: number, db: string) =>
    http.get<{ data: string[] }>(`/api/db-servers/${id}/databases/${db}/tables`),
}
