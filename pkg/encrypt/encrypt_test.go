package encrypt_test

import (
	"testing"

	"github.com/sshive/sshive/pkg/encrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := "super-secret-password-123"
	ciphertext, err := encrypt.Encrypt(key, plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	decrypted, err := encrypt.Decrypt(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDifferentEachTime(t *testing.T) {
	key := make([]byte, 32)
	c1, _ := encrypt.Encrypt(key, "hello")
	c2, _ := encrypt.Encrypt(key, "hello")
	assert.NotEqual(t, c1, c2)
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1
	ciphertext, _ := encrypt.Encrypt(key1, "secret")
	_, err := encrypt.Decrypt(key2, ciphertext)
	assert.Error(t, err)
}
