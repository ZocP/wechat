package tests

import (
	"encoding/base64"
	"pickup/internal/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoUtilKeyHandling(t *testing.T) {
	t.Run("short key still encrypts/decrypts correctly", func(t *testing.T) {
		util := utils.NewCryptoUtil("short")
		plaintext := "test data"
		ciphertext, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		decrypted, err := util.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("long key still encrypts/decrypts correctly", func(t *testing.T) {
		longKey := strings.Repeat("a", 40)
		util := utils.NewCryptoUtil(longKey)
		plaintext := "test data"
		ciphertext, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		decrypted, err := util.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("exact 32 byte key works", func(t *testing.T) {
		exactKey := strings.Repeat("x", 32)
		util := utils.NewCryptoUtil(exactKey)
		plaintext := "test data"
		ciphertext, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		decrypted, err := util.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestCryptoUtilEncryptDecrypt(t *testing.T) {
	util := utils.NewCryptoUtil("this-is-a-test-key-32-chars-long")

	t.Run("basic encrypt decrypt", func(t *testing.T) {
		plaintext := "hello world"
		ciphertext, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, ciphertext, "ciphertext should differ from plaintext")

		decrypted, err := util.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("empty string encrypt decrypt", func(t *testing.T) {
		ciphertext, err := util.Encrypt("")
		require.NoError(t, err)
		decrypted, err := util.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, "", decrypted)
	})

	t.Run("unicode encrypt decrypt", func(t *testing.T) {
		plaintext := "你好世界 🌍"
		ciphertext, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		decrypted, err := util.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("different encryptions produce different ciphertexts", func(t *testing.T) {
		plaintext := "hello world"
		c1, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		c2, err := util.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, c1, c2, "GCM with random nonce should produce different ciphertexts")
	})

	t.Run("different keys produce different ciphertexts", func(t *testing.T) {
		util2 := utils.NewCryptoUtil("another-key-that-is-32-chars!!")
		plaintext := "secret"
		c1, _ := util.Encrypt(plaintext)
		c2, _ := util2.Encrypt(plaintext)
		assert.NotEqual(t, c1, c2)
	})

	t.Run("wrong key cannot decrypt", func(t *testing.T) {
		util2 := utils.NewCryptoUtil("wrong-key-wrong-key-wrong-key-32")
		ciphertext, err := util.Encrypt("secret")
		require.NoError(t, err)
		_, err = util2.Decrypt(ciphertext)
		assert.Error(t, err)
	})
}

func TestCryptoUtilDecryptErrors(t *testing.T) {
	util := utils.NewCryptoUtil("example-key")

	t.Run("invalid base64", func(t *testing.T) {
		_, err := util.Decrypt("not-valid-base64!!!")
		assert.Error(t, err)
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		shortCipher := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
		_, err := util.Decrypt(shortCipher)
		assert.Error(t, err)
	})

	t.Run("empty ciphertext", func(t *testing.T) {
		_, err := util.Decrypt("")
		assert.Error(t, err)
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		ciphertext, err := util.Encrypt("hello")
		require.NoError(t, err)
		// Corrupt the ciphertext
		data, _ := base64.StdEncoding.DecodeString(ciphertext)
		data[len(data)-1] ^= 0xff
		corrupted := base64.StdEncoding.EncodeToString(data)
		_, err = util.Decrypt(corrupted)
		assert.Error(t, err)
	})
}
