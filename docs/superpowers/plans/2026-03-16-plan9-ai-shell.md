# Plan 9 — AI Shell 助手

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在终端侧边栏提供自然语言 → Shell 命令转换，用户确认后插入终端，不自动执行

**Architecture:** 后端新建 `internal/ai/` 包，调用 Claude API（流式 SSE），系统提示词注入目标主机 OS 信息；前端终端页右侧抽屉，输入框 + 流式渲染结果 + 「插入命令」按钮，命令通过 xterm `.paste()` 写入终端输入行

**Tech Stack:** Go + Gin SSE + Anthropic Go SDK；Vue 3 + TypeScript；xterm.js paste API

---

## 环境变量

```bash
SSHIVE_AI_API_KEY=sk-ant-...      # Anthropic API Key
SSHIVE_AI_MODEL=claude-haiku-4-5-20251001  # 默认用 Haiku，快且便宜
```

---

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/config/config.go` | 新增 `AI.APIKey`、`AI.Model` 配置字段 |
| `internal/ai/handler.go` | `POST /api/ai/shell` — 接收 prompt + host_id，流式 SSE 返回命令 |
| `internal/ai/service.go` | 调用 Anthropic API，构造系统提示词，流式输出 |
| `cmd/server/main.go` | 注册 AI 路由 |
| `frontend/src/api/ai.ts` | `streamShellCommand(prompt, hostId)` — 读取 SSE 流 |
| `frontend/src/views/TerminalView.vue` | 右侧 AI 抽屉，流式渲染，插入按钮 |

---

## Chunk 1: 后端 — AI 接口

### Task 1: 配置扩展

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: 在 Config 结构体加 AI 字段**

```go
// internal/config/config.go
type Config struct {
    DB       DBConfig    `yaml:"db"`
    Redis    RedisConfig `yaml:"redis"`
    AI       AIConfig    `yaml:"ai"`       // 新增
    EncryptKey string    `yaml:"encrypt_key"`
    JWTSecret  string    `yaml:"jwt_secret"`
    CastDir    string    `yaml:"cast_dir"`
    Port       int       `yaml:"port"`
}

type AIConfig struct {
    APIKey string `yaml:"api_key"`
    Model  string `yaml:"model"`
}
```

- [ ] **Step 2: 在 config.yaml.example（或 README）说明新配置项**

```yaml
ai:
  api_key: "sk-ant-..."
  model: "claude-haiku-4-5-20251001"
```

- [ ] **Step 3: Commit**

```bash
git add internal/config/config.go
git commit -m "feat(ai): add AI config fields"
```

---

### Task 2: AI Service

**Files:**
- Create: `internal/ai/service.go`

- [ ] **Step 1: 安装 Anthropic Go SDK**

```bash
go get github.com/anthropics/anthropic-sdk-go
```

- [ ] **Step 2: 创建 service.go**

```go
// internal/ai/service.go
package ai

import (
    "context"
    "fmt"

    anthropic "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
    "github.com/sshive/sshive/internal/config"
)

type Service struct {
    client *anthropic.Client
    model  string
}

func NewService() *Service {
    cfg := config.C.AI
    model := cfg.Model
    if model == "" {
        model = "claude-haiku-4-5-20251001"
    }
    client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
    return &Service{client: &client, model: model}
}

const systemPrompt = `你是一个 Linux 运维助手。用户会用自然语言描述想要做的操作，你需要返回对应的 Shell 命令。

规则：
- 只返回命令本身，不要解释，不要 markdown 代码块
- 如果需要多条命令，用换行分隔
- 命令必须安全，不要包含 rm -rf / 等破坏性操作
- 根据用户提供的 OS 信息调整命令语法
- 如果需求不明确，返回最常见的安全写法`

// StreamShellCommand 流式返回 Shell 命令，通过 chunk channel 发送每个文本块
func (s *Service) StreamShellCommand(ctx context.Context, prompt, osInfo string, chunk chan<- string) error {
    userMsg := prompt
    if osInfo != "" {
        userMsg = fmt.Sprintf("[目标主机 OS: %s]\n\n%s", osInfo, prompt)
    }

    stream := s.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
        Model:     anthropic.Model(s.model),
        MaxTokens: 512,
        System: []anthropic.TextBlockParam{
            {Text: systemPrompt},
        },
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
        },
    })

    for stream.Next() {
        event := stream.Current()
        if delta, ok := event.Delta.(anthropic.ContentBlockDeltaEventDelta); ok {
            if delta.Text != "" {
                select {
                case chunk <- delta.Text:
                case <-ctx.Done():
                    return ctx.Err()
                }
            }
        }
    }
    return stream.Err()
}
```

- [ ] **Step 3: go build ./... 确认编译**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/ai/service.go go.mod go.sum
git commit -m "feat(ai): AI service with streaming shell command generation"
```

---

### Task 3: AI Handler（SSE）

**Files:**
- Create: `internal/ai/handler.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 创建 handler.go**

```go
// internal/ai/handler.go
package ai

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
)

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

type ShellRequest struct {
    Prompt string `json:"prompt" binding:"required"`
    OSInfo string `json:"os_info"` // 可选，前端传当前主机 OS
}

// Shell 接收自然语言，SSE 流式返回 Shell 命令
func (h *Handler) Shell(c *gin.Context) {
    var req ShellRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
        return
    }

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("X-Accel-Buffering", "no")

    chunk := make(chan string, 32)
    errCh := make(chan error, 1)

    go func() {
        errCh <- h.svc.StreamShellCommand(c.Request.Context(), req.Prompt, req.OSInfo, chunk)
        close(chunk)
    }()

    c.Stream(func(w io.Writer) bool {
        select {
        case text, ok := <-chunk:
            if !ok {
                fmt.Fprintf(w, "event: done\ndata: \n\n")
                return false
            }
            fmt.Fprintf(w, "data: %s\n\n", text)
            return true
        case <-c.Request.Context().Done():
            return false
        }
    })
}
```

补充 import `"io"` 到 handler.go。

- [ ] **Step 2: 注册路由（cmd/server/main.go）**

在 `authed` 路由组里加：

```go
aiSvc := ai.NewService()
aiH := ai.NewHandler(aiSvc)
authed.POST("/ai/shell", aiH.Shell)
```

同时在 import 里加 `"github.com/sshive/sshive/internal/ai"`。

- [ ] **Step 3: go build ./... 确认编译**

```bash
go build ./...
```

- [ ] **Step 4: 用 curl 快速测试（需配置好 API Key）**

```bash
curl -X POST http://localhost:8080/api/ai/shell \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"列出当前目录下7天前修改的文件"}' \
  --no-buffer
# 期望：流式输出 find . -mtime +7 之类的命令
```

- [ ] **Step 5: Commit**

```bash
git add internal/ai/handler.go cmd/server/main.go
git commit -m "feat(ai): SSE handler for shell command generation"
```

---

## Chunk 2: 前端 — AI 抽屉

### Task 4: API 层 & 终端集成

**Files:**
- Create: `frontend/src/api/ai.ts`
- Modify: `frontend/src/views/TerminalView.vue`

- [ ] **Step 1: 创建 ai.ts**

```typescript
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
```

- [ ] **Step 2: TerminalView.vue 加 AI 抽屉按钮和面板**

在 tab-bar 右侧（主题选择器旁）加 AI 按钮：

```vue
<!-- tab-bar 里 tab-actions 内，主题选择器后面 -->
<button class="ai-toggle-btn" :class="{ active: aiPanelOpen }"
  @click="aiPanelOpen = !aiPanelOpen" title="AI Shell 助手">
  ✦ AI
</button>
```

在 `terminal-area` 外层包一层 flex 容器，右侧加 AI 面板：

```vue
<div class="terminal-wrapper">
  <div class="terminal-area" :class="`split-${splitCount}`">
    <!-- 原有分屏内容不变 -->
  </div>

  <!-- AI 侧边栏 -->
  <div v-if="aiPanelOpen" class="ai-panel">
    <div class="ai-panel-header">
      <span>✦ AI Shell 助手</span>
      <button class="ai-close" @click="aiPanelOpen = false">×</button>
    </div>
    <div class="ai-result" v-if="aiResult || aiLoading">
      <pre class="ai-code">{{ aiResult }}<span v-if="aiLoading" class="ai-cursor">▌</span></pre>
      <button v-if="aiResult && !aiLoading" class="ai-insert-btn" @click="insertCommand">
        插入终端
      </button>
    </div>
    <div class="ai-input-row">
      <textarea
        v-model="aiPrompt"
        class="ai-textarea"
        placeholder="描述你想做的操作…&#10;例：找出占用磁盘最大的10个目录"
        rows="3"
        @keydown.enter.ctrl="askAI"
      />
      <button class="ai-send-btn" :disabled="aiLoading" @click="askAI">
        {{ aiLoading ? '生成中…' : '生成命令' }}
      </button>
    </div>
  </div>
</div>
```

- [ ] **Step 3: script 里加 AI 相关状态和函数**

```typescript
import { streamShellCommand } from '@/api/ai'

const aiPanelOpen = ref(false)
const aiPrompt = ref('')
const aiResult = ref('')
const aiLoading = ref(false)

async function askAI() {
  if (!aiPrompt.value.trim() || aiLoading.value) return
  aiResult.value = ''
  aiLoading.value = true
  await streamShellCommand({
    prompt: aiPrompt.value,
    onChunk: (text) => { aiResult.value += text },
    onDone: () => { aiLoading.value = false },
    onError: (err) => { aiResult.value = `错误：${err}`; aiLoading.value = false },
  })
}

function insertCommand() {
  // 找当前 active 分屏的 xterm，调用 paste 写入输入行（不执行）
  const xterm = splitXterms[activeSplitIdx.value]
  if (!xterm) return
  const cmd = aiResult.value.trim()
  const tab = store.getTab(splitTabIds.value[activeSplitIdx.value])
  if (tab?.ws?.readyState === WebSocket.OPEN) {
    tab.ws.send(cmd)  // 发到 SSH stdin（PTY 会回显，但不执行）
  }
}
```

- [ ] **Step 4: 加 AI 面板样式**

```css
/* terminal-wrapper 包裹层 */
.terminal-wrapper { display: flex; flex: 1; overflow: hidden; }
.terminal-area { flex: 1; } /* 移除原来的 flex:1 改到这里 */

/* AI 侧边栏 */
.ai-panel {
  width: 300px; flex-shrink: 0;
  display: flex; flex-direction: column;
  background: color-mix(in srgb, var(--terminal-bg) 90%, #000);
  border-left: 1px solid color-mix(in srgb, var(--terminal-fg) 12%, transparent);
}
.ai-panel-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 12px; font-size: 13px; font-weight: 600;
  color: var(--accent);
  border-bottom: 1px solid color-mix(in srgb, var(--terminal-fg) 10%, transparent);
}
.ai-close { background: none; border: none; color: inherit; cursor: pointer; font-size: 16px; }
.ai-result { flex: 1; overflow-y: auto; padding: 10px 12px; }
.ai-code {
  font-family: 'JetBrains Mono', monospace; font-size: 12px;
  color: var(--terminal-fg); white-space: pre-wrap; word-break: break-all;
  background: color-mix(in srgb, var(--accent) 8%, transparent);
  padding: 8px 10px; border-radius: 4px; margin: 0;
}
.ai-cursor { animation: blink 1s step-end infinite; }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }
.ai-insert-btn {
  margin-top: 8px; width: 100%;
  background: var(--accent); color: #fff; border: none;
  border-radius: 4px; padding: 6px; cursor: pointer; font-size: 12px;
}
.ai-insert-btn:hover { opacity: 0.85; }
.ai-input-row { padding: 10px 12px; display: flex; flex-direction: column; gap: 6px;
  border-top: 1px solid color-mix(in srgb, var(--terminal-fg) 10%, transparent); }
.ai-textarea {
  width: 100%; resize: none; font-size: 12px;
  background: color-mix(in srgb, var(--terminal-fg) 6%, transparent);
  border: 1px solid color-mix(in srgb, var(--terminal-fg) 18%, transparent);
  border-radius: 4px; padding: 6px 8px; color: var(--terminal-fg);
  font-family: inherit;
}
.ai-textarea:focus { outline: none; border-color: var(--accent); }
.ai-send-btn {
  background: color-mix(in srgb, var(--accent) 20%, transparent);
  color: var(--accent); border: 1px solid color-mix(in srgb, var(--accent) 40%, transparent);
  border-radius: 4px; padding: 5px; cursor: pointer; font-size: 12px;
}
.ai-send-btn:hover:not(:disabled) { background: color-mix(in srgb, var(--accent) 30%, transparent); }
.ai-send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.ai-toggle-btn {
  background: transparent; border: 1px solid color-mix(in srgb, var(--terminal-fg) 20%, transparent);
  color: color-mix(in srgb, var(--terminal-fg) 55%, transparent);
  border-radius: 4px; padding: 2px 8px; cursor: pointer; font-size: 11px;
  font-weight: 600; transition: all 0.15s;
}
.ai-toggle-btn:hover, .ai-toggle-btn.active {
  border-color: var(--accent); color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
}
```

- [ ] **Step 5: 前端构建**

```bash
cd frontend && npm run build
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/ai.ts frontend/src/views/TerminalView.vue
git commit -m "feat(ai): AI shell panel with streaming command generation"
```

---

## Chunk 3: 联调验收

- [ ] **Step 1: config.yaml 加 AI 配置**

```yaml
ai:
  api_key: "sk-ant-..."
  model: "claude-haiku-4-5-20251001"
```

- [ ] **Step 2: 重新编译启动**

```bash
go build -o server ./cmd/server/ && ./server
```

- [ ] **Step 3: 手动验收**
  1. 打开终端页，右上角出现「✦ AI」按钮
  2. 点击打开侧边栏，输入「查看内存使用情况」→ 点「生成命令」
  3. 命令流式显示出来（如 `free -h`）
  4. 点「插入终端」→ 命令出现在 SSH 终端输入行，未自动执行
  5. 手动按 Enter 执行

- [ ] **Step 4: 最终 Commit & Push**

```bash
git add -A
git commit -m "feat: AI shell assistant complete"
git push
```
