package cast_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sshive/sshive/pkg/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCastWriter(t *testing.T) {
	path := t.TempDir() + "/test.cast"

	w, err := cast.NewWriter(path, 80, 24)
	require.NoError(t, err)

	w.WriteOutput(0, "Hello ")
	w.WriteOutput(100*time.Millisecond, "World\r\n")
	w.WriteInput(200*time.Millisecond, "ls\n")
	require.NoError(t, w.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.GreaterOrEqual(t, len(lines), 3) // header + 2 output + 1 input

	// 第一行是 header JSON
	assert.Contains(t, lines[0], `"version":2`)
	assert.Contains(t, lines[0], `"width":80`)
	assert.Contains(t, lines[0], `"height":24`)

	// 第二行是第一个 output 事件
	assert.Contains(t, lines[1], `"o"`)
	assert.Contains(t, lines[1], "Hello ")
}
