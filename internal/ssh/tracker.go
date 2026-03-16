package ssh

import (
	"regexp"
	"strings"
	"sync"
)

var (
	// 匹配 ANSI 转义序列 / 控制字符
	ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b[()][AB012]|\r`)
	// 匹配 shell 提示符末尾（支持 $、#，含颜色等前缀）
	shellPromptRe = regexp.MustCompile(`[#$]\s*$`)
	// 用于从最后一行提取命令：去掉提示符前缀
	promptPrefixRe = regexp.MustCompile(`^.*?[#$]\s*`)
)

const (
	maxContextBuf = 64 * 1024 // 64KB 滚动上下文缓冲（用于提取命令）
	maxResultSize = 10 * 1024 // 10KB 结果上限（防止 top/tail -f 等命令撑爆）
)

// CommandTracker 将用户输入的命令与其 SSH 输出关联起来。
//
// 关键设计：命令从「输出缓冲区的最后一行」提取（而非原始输入），
// 这样 Tab 补全、方向键回召等经过 PTY 渲染后的实际命令都能被准确捕获。
type CommandTracker struct {
	mu         sync.Mutex
	contextBuf strings.Builder // 滚动上下文（用于 OnEnter 时提取命令）
	resultBuf  strings.Builder // 当前命令输出（用于结果提取）
	currentCmd string
	waiting    bool
}

// FeedOutput 处理一段 SSH 输出。
// 返回 (cmd, result, true) 表示当前命令执行完毕。
func (t *CommandTracker) FeedOutput(data string) (cmd, result string, done bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 始终追加到上下文缓冲（用于下次 OnEnter）
	t.contextBuf.WriteString(data)
	if t.contextBuf.Len() > maxContextBuf {
		s := t.contextBuf.String()
		t.contextBuf.Reset()
		t.contextBuf.WriteString(s[len(s)-maxContextBuf/2:])
	}

	if !t.waiting {
		return "", "", false
	}

	t.resultBuf.WriteString(data)

	// 结果超过上限（如 top/tail -f）：放弃本次记录，等下一条命令
	if t.resultBuf.Len() > maxResultSize {
		t.waiting = false
		t.currentCmd = ""
		t.resultBuf.Reset()
		return "", "", false
	}

	// 判断本次 chunk 是否以 shell 提示符结尾
	clean := stripAnsi(data)
	if !shellPromptRe.MatchString(strings.TrimRight(clean, " \t")) {
		return "", "", false
	}

	result = extractResult(t.resultBuf.String(), t.currentCmd)
	cmd = t.currentCmd
	t.waiting = false
	t.currentCmd = ""
	t.resultBuf.Reset()
	return cmd, result, true
}

// OnEnter 在用户按下回车时调用。
// 从上下文缓冲最后一行提取实际命令（包含 Tab 补全后的文本），
// 然后开始收集该命令的输出。
func (t *CommandTracker) OnEnter() {
	t.mu.Lock()
	defer t.mu.Unlock()

	cmd := extractCommandFromLastLine(t.contextBuf.String())
	if cmd == "" {
		return
	}
	t.currentCmd = cmd
	t.waiting = true
	t.resultBuf.Reset()
}

// extractCommandFromLastLine 从终端输出中提取最后一行并去除提示符前缀。
func extractCommandFromLastLine(output string) string {
	clean := stripAnsi(output)
	lines := strings.Split(strings.TrimRight(clean, "\n"), "\n")
	lastLine := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastLine = lines[i]
			break
		}
	}
	if lastLine == "" {
		return ""
	}
	// 去除 "[user@host ~]$ " 之类的提示符前缀
	cmd := promptPrefixRe.ReplaceAllString(lastLine, "")
	return strings.TrimSpace(cmd)
}

// stripAnsi 移除 ANSI 转义序列和回车符。
func stripAnsi(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}

// extractResult 从原始输出中去除命令 echo 和尾部提示符，返回干净的结果文本。
func extractResult(raw, cmd string) string {
	clean := stripAnsi(raw)
	// 去除命令 echo（第一次出现）
	if cmd != "" {
		if idx := strings.Index(clean, cmd); idx >= 0 {
			clean = clean[idx+len(cmd):]
		}
	}
	// 去除末尾空行和 shell 提示符行
	lines := strings.Split(clean, "\n")
	for len(lines) > 0 {
		trimmed := strings.TrimSpace(lines[len(lines)-1])
		if trimmed == "" || shellPromptRe.MatchString(trimmed) {
			lines = lines[:len(lines)-1]
		} else {
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
