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
)

const maxTrackerBuf = 64 * 1024

// CommandTracker 将用户输入的命令与其 SSH 输出关联起来。
// 当用户按下回车后，等待下一次出现 shell 提示符，提取中间的输出作为结果。
type CommandTracker struct {
	mu         sync.Mutex
	buf        strings.Builder
	currentCmd string
	waiting    bool
}

// SetCommand 在用户确认命令（非拦截的 Enter）时调用，开始收集该命令的输出。
func (t *CommandTracker) SetCommand(cmd string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentCmd = cmd
	t.waiting = true
	t.buf.Reset()
}

// FeedOutput 处理一段 SSH 输出。
// 当检测到命令执行完毕（出现 shell 提示符），返回 (cmd, result, true)。
func (t *CommandTracker) FeedOutput(data string) (cmd, result string, done bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.buf.WriteString(data)
	// 防止无界增长，超出后保留后半段
	if t.buf.Len() > maxTrackerBuf {
		s := t.buf.String()
		t.buf.Reset()
		t.buf.WriteString(s[len(s)-maxTrackerBuf/2:])
	}

	if !t.waiting {
		return "", "", false
	}

	// 判断本次 chunk 是否以 shell 提示符结尾
	clean := stripAnsi(data)
	if !shellPromptRe.MatchString(strings.TrimRight(clean, " \t")) {
		return "", "", false
	}

	result = extractResult(t.buf.String(), t.currentCmd)
	cmd = t.currentCmd
	t.waiting = false
	t.currentCmd = ""
	t.buf.Reset()
	return cmd, result, true
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
	// 去除末尾的空行和 shell 提示符行
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
