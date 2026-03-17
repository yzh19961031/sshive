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

export interface DBQueryLog {
  id: number
  tenant_id: number
  db_server_id: number
  user_id: number
  database: string
  sql: string
  duration_ms: number
  rows_returned: number
  error_msg: string
  created_at: string
}

export interface QueryHistory {
  sql: string
  database: string
  durationMs: number
  rowsReturned: number
  error: string
  executedAt: Date
}

export const dbApi = {
  list: () =>
    http.get<{ code: number; data: DBServer[] }>('/db-servers'),
  create: (data: Omit<DBServer, 'id'> & { password: string }) =>
    http.post('/db-servers', data),
  delete: (id: number) =>
    http.delete(`/db-servers/${id}`),
  query: (id: number, sql: string, database?: string) =>
    http.post<{ code: number; data: QueryResult }>(`/db-servers/${id}/query`, { sql, database }),
  databases: (id: number) =>
    http.get<{ code: number; data: string[] }>(`/db-servers/${id}/databases`),
  tables: (id: number, db: string) =>
    http.get<{ code: number; data: string[] }>(`/db-servers/${id}/databases/${db}/tables`),
  listQueryLogs: (params: {
    db_server_id?: number
    user_id?: number
    page?: number
    page_size?: number
    start_time?: string
    end_time?: string
  }) =>
    http.get<{ code: number; data: { list: DBQueryLog[]; total: number } }>('/db-query-logs', { params }),
}
