package dangerrule_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 直接测试正则逻辑，不依赖数据库
func TestDangerPatterns(t *testing.T) {
	cases := []struct {
		pattern string
		cmd     string
		match   bool
	}{
		{`rm\s+-rf\s+/`, "rm -rf /", true},
		{`rm\s+-rf\s+/`, "rm -rf /tmp/log", true},
		{`rm\s+-rf\s+/`, "rm -rf ./mydir", false},
		{`shutdown|reboot|halt`, "shutdown now", true},
		{`shutdown|reboot|halt`, "reboot", true},
		{`shutdown|reboot|halt`, "echo hello", false},
		{`mkfs\..*`, "mkfs.ext4 /dev/sdb", true},
		{`mkfs\..*`, "ls /dev/sdb", false},
		{`dd\s+if=.*of=/dev/`, "dd if=/dev/zero of=/dev/sda", true},
		{`dd\s+if=.*of=/dev/`, "dd if=image.iso of=/tmp/out", false},
	}
	for _, tc := range cases {
		re := regexp.MustCompile(tc.pattern)
		got := re.MatchString(tc.cmd)
		assert.Equal(t, tc.match, got, "pattern=%q cmd=%q", tc.pattern, tc.cmd)
	}
}
