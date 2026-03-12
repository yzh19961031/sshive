package auth_test

import (
	"testing"

	"github.com/sshive/sshive/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
	assert.True(t, auth.HasPermission("super_admin", "tenant:manage"))
	assert.True(t, auth.HasPermission("operator", "host:connect"))
	assert.False(t, auth.HasPermission("operator", "user:manage"))
	assert.False(t, auth.HasPermission("auditor", "host:connect"))
	assert.False(t, auth.HasPermission("unknown_role", "host:connect"))
}
