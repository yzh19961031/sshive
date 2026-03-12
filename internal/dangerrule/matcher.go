package dangerrule

import (
	"regexp"
	"sync"
)

// Matcher 维护编译好的正则表达式缓存
type Matcher struct {
	mu    sync.RWMutex
	repo  *Repo
	cache map[int64][]*compiledRule // tenantID -> 编译好的规则列表
}

type compiledRule struct {
	id      int64
	pattern string
	re      *regexp.Regexp
}

var GlobalMatcher = &Matcher{
	repo:  &Repo{},
	cache: make(map[int64][]*compiledRule),
}

// Reload 重新从数据库加载并编译指定租户的规则
func (m *Matcher) Reload(tenantID int64) error {
	rules, err := m.repo.ListActive(tenantID)
	if err != nil {
		return err
	}
	compiled := make([]*compiledRule, 0, len(rules))
	for _, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			continue // 跳过无效正则
		}
		compiled = append(compiled, &compiledRule{
			id:      r.ID,
			pattern: r.Pattern,
			re:      re,
		})
	}
	m.mu.Lock()
	m.cache[tenantID] = compiled
	m.mu.Unlock()
	return nil
}

// Match 检查命令是否命中规则，返回命中的规则 pattern（未命中返回空字符串）
func (m *Matcher) Match(tenantID int64, command string) string {
	m.mu.RLock()
	rules, ok := m.cache[tenantID]
	m.mu.RUnlock()
	if !ok {
		// 首次使用时懒加载
		_ = m.Reload(tenantID)
		m.mu.RLock()
		rules = m.cache[tenantID]
		m.mu.RUnlock()
	}
	for _, r := range rules {
		if r.re.MatchString(command) {
			return r.pattern
		}
	}
	return ""
}
