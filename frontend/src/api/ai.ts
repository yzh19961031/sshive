// frontend/src/api/ai.ts
import { useAuthStore } from '@/stores/auth'

export interface ShellStreamOptions {
  prompt: string
  osInfo?: string
  onChunk: (text: string) => void
  onDone: () => void
  onError: (err: string) => void
}

export async function streamShellCommand(opts: ShellStreamOptions) {
  const auth = useAuthStore()
  const res = await fetch('/api/ai/shell', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${auth.token}`,
    },
    body: JSON.stringify({ prompt: opts.prompt, os_info: opts.osInfo ?? '' }),
  })

  if (!res.ok || !res.body) {
    opts.onError('请求失败')
    return
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    const lines = buf.split('\n')
    buf = lines.pop() ?? ''
    for (const line of lines) {
      if (line.startsWith('event: done')) { opts.onDone(); return }
      if (line.startsWith('data: ')) opts.onChunk(line.slice(6))
    }
  }
  opts.onDone()
}
