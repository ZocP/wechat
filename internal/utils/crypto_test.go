package utils

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoUtil_KeyPadding(t *testing.T) {
	t.Run("short key is padded to 32 bytes", func(t *testing.T) {
		util := NewCryptoUtil("short")
		assert.Len(t, util.key, 32)
		assert.Equal(t, []byte("short"), util.key[:5])
		for i := 5; i < 32; i++ {
			assert.Equal(t, byte(0), util.key[i], "byte at position %d should be 0", i)
		}
	})

	t.Run("long key is truncated to 32 bytes", func(t *testing.T) {
		longKey := strings.Repeat("a", 40)
		util := NewCryptoUtil(longKey)
		assert.Len(t, util.key, 32)
		assert.Equal(t, longKey[:32], string(util.key))
	})

	t.Run("exact 32 byte key is unchanged", func(t *testing.T) {
		exactKey := strings.Repeat("k", 32)
		util := NewCryptoUtil(exactKey)
		assert.Len(t, util.key, 32)
		assert.Equal(t, exactKey, string(util.key))
	})

	t.Run("empty key is padded to 32 bytes", func(t *testing.T) {
		util := NewCryptoUtil("")
		assert.Len(t, util.key, 32)
		for i := 0; i < 32; i++ {
			assert.Equal(t, byte(0), util.key[i])
		}
	})
}

func TestCryptoUtil_EncryptDecrypt(t *testing.T) {
	util := NewCryptoUtil("test-key-for-unit-testing-32char")

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"unicode", "你好世界 🌍"},
		{"long text", strings.Repeat("abcdefghij", 100)},
		{"special chars", "!@#$%^&*()_+-={}[]|\\:;\"'<>,.?/"},
		{"newlines and tabs", "line1\nline2\ttab"},
		{"single char", "a"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := util.Encrypt(tc.plaintext)
			require.NoError(t, err)

			if tc.plaintext != "" {
				assert.NotEqual(t, tc.plaintext, ciphertext)
			}

			decrypted, err := util.Decrypt(ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}

func TestCryptoUtil_NonDeterministic(t *testing.T) {
	util := NewCryptoUtil("test-key")
	plaintext := "hello"

	c1, err := util.Encrypt(plaintext)
	require.NoError(t, err)
	c2, err := util.Encrypt(plaintext)
	require.NoError(t, err)

	assert.NotEqual(t, c1, c2, "GCM nonce should make each encryption unique")

	// Both should decrypt to the same value
	d1, err := util.Decrypt(c1)
	require.NoError(t, err)
	d2, err := util.Decrypt(c2)
	require.NoError(t, err)
	assert.Equal(t, d1, d2)
}

func TestCryptoUtil_WrongKey(t *testing.T) {
	util1 := NewCryptoUtil("key-one-for-encryption-32-chars!")
	util2 := NewCryptoUtil("key-two-for-decryption-32-chars!")

	ciphertext, err := util1.Encrypt("secret data")
	require.NoError(t, err)

	_, err = util2.Decrypt(ciphertext)
	assert.Error(t, err)
}

func TestCryptoUtil_DecryptErrors(t *testing.T) {
	util := NewCryptoUtil("test-key")

	t.Run("invalid base64", func(t *testing.T) {
		_, err := util.Decrypt("not-valid-base64!!!")
		assert.Error(t, err)
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		short := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
		_, err := util.Decrypt(short)
		assert.Error(t, err)
	})

	t.Run("empty input", func(t *testing.T) {
		_, err := util.Decrypt("")
		assert.Error(t, err)
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		ciphertext, err := util.Encrypt("hello")
		require.NoError(t, err)

		data, err := base64.StdEncoding.DecodeString(ciphertext)
		require.NoError(t, err)
		data[len(data)-1] ^= 0xff
		corrupted := base64.StdEncoding.EncodeToString(data)

		_, err = util.Decrypt(corrupted)
		assert.Error(t, err)
	})
}
