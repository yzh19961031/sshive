// frontend/src/api/ai.ts
import { useAuthStore } from '@/stores/auth'

export interface AIMessage {
  role: 'user' | 'assistant'
  content: string
}

export interface ShellStreamOptions {
  messages: AIMessage[]
  osInfo?: string
  onChunk: (text: string) => void
  onDone: () => void
  onError: (err: string) => void
}

export interface SQLStreamOptions {
  messages: AIMessage[]
  dbContext?: string
  onChunk: (text: string) => void
  onDone: () => void
  onError: (err: string) => void
}

async function sseStream(url: string, body: object, opts: { onChunk: (t: string) => void; onDone: () => void; onError: (e: string) => void }) {
  const auth = useAuthStore()
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${auth.token}` },
    body: JSON.stringify(body),
  })
  if (!res.ok || !res.body) { opts.onError('请求失败'); return }
  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = '', lastEvent = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    const lines = buf.split('\n')
    buf = lines.pop() ?? ''
    for (const line of lines) {
      if (line.startsWith('event: ')) { lastEvent = line.slice(7).trim() }
      else if (line.startsWith('data: ')) {
        const data = line.slice(6)
        if (lastEvent === 'done') { opts.onDone(); return }
        if (lastEvent === 'error') { opts.onError(data); return }
        opts.onChunk(data); lastEvent = ''
      }
    }
  }
  opts.onDone()
}

export async function streamSQLCommand(opts: SQLStreamOptions) {
  await sseStream('/api/ai/sql', { messages: opts.messages, db_context: opts.dbContext ?? '' }, opts)
}

export async function streamShellCommand(opts: ShellStreamOptions) {
  await sseStream('/api/ai/shell', { messages: opts.messages, os_info: opts.osInfo ?? '' }, opts)
}
