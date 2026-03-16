package ssh

import (
	"strings"

	"github.com/sshive/sshive/internal/dangerrule"
)

// Interceptor 缓冲用户输入，在回车时检测高危指令
type Interceptor struct {
	tenantID int64
	buf      strings.Builder
}

func NewInterceptor(tenantID int64) *Interceptor {
	return &Interceptor{tenantID: tenantID}
}

// InterceptResult 拦截结果
type InterceptResult struct {
	Blocked     bool
	Command     string
	MatchedRule string
}

// processBackspaces 处理退格符（0x7f / 0x08），返回视觉上等价的字符串。
func processBackspaces(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\x7f' || r == '\b' {
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// Feed 处理一段输入数据。
// 若未触发回车，返回 nil 表示继续缓冲。
// 调用方只在 result.Blocked==false 时才透传数据给 SSH 服务器。
func (ic *Interceptor) Feed(data string) *InterceptResult {
	ic.buf.WriteString(data)
	if !strings.ContainsAny(data, "\r\n") {
		return nil
	}
	cmd := processBackspaces(strings.TrimRight(ic.buf.String(), "\r\n"))
	ic.buf.Reset()

	if cmd == "" {
		return &InterceptResult{Blocked: false, Command: cmd}
	}

	matched := dangerrule.GlobalMatcher.Match(ic.tenantID, cmd)
	if matched != "" {
		return &InterceptResult{Blocked: true, Command: cmd, MatchedRule: matched}
	}
	return &InterceptResult{Blocked: false, Command: cmd}
}
