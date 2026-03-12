package credential_test

import (
	"encoding/base64"
	"testing"

	"github.com/sshive/sshive/internal/config"
	"github.com/sshive/sshive/pkg/encrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	key := base64.StdEncoding.EncodeToString(raw)
	config.C.EncryptKey = key

	encKey, err := base64.StdEncoding.DecodeString(config.C.EncryptKey)
	require.NoError(t, err)

	secret := "my-private-key-content"
	encrypted, err := encrypt.Encrypt(encKey, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encrypt.Decrypt(encKey, encrypted)
	require.NoError(t, err)
	assert.Equal(t, secret, decrypted)
}
