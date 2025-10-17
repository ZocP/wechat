package tests

import (
	"encoding/base64"
	"pickup/internal/utils"
	"strings"
	"testing"
)

func TestNewCryptoUtilKeyLength(t *testing.T) {
	t.Run("pad short key to 32 bytes", func(t *testing.T) {
		util := utils.NewCryptoUtil("short")
		if got := len(util.key); got != 32 {
			t.Fatalf("expected key length 32, got %d", got)
		}
		// ensure padded bytes are zeros
		for i := len("short"); i < 32; i++ {
			if util.key[i] != 0 {
				t.Fatalf("expected padding byte at position %d to be 0, got %d", i, util.key[i])
			}
		}
	})

	t.Run("truncate long key to 32 bytes", func(t *testing.T) {
		longKey := strings.Repeat("a", 40)
		util := utils.NewCryptoUtil(longKey)
		if got := len(util.key); got != 32 {
			t.Fatalf("expected key length 32, got %d", got)
		}
		expectedPrefix := longKey[:32]
		if string(util.key) != expectedPrefix {
			t.Fatalf("expected key prefix %q, got %q", expectedPrefix, string(util.key))
		}
	})
}

func TestCryptoUtilEncryptDecrypt(t *testing.T) {
	util := utils.NewCryptoUtil("this-is-a-test-key")

	plaintext := "hello world"
	ciphertext, err := util.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt returned error: %v", err)
	}

	decrypted, err := util.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt returned error: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected decrypted text %q, got %q", plaintext, decrypted)
	}

	// ciphertext should not match plaintext
	if ciphertext == plaintext {
		t.Fatalf("ciphertext should not equal plaintext")
	}
}

func TestCryptoUtilDecryptErrors(t *testing.T) {
	util := utils.NewCryptoUtil("example-key")

	t.Run("invalid base64", func(t *testing.T) {
		if _, err := util.Decrypt("not-base64"); err == nil {
			t.Fatal("expected error for invalid base64 input")
		}
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		shortCipher := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
		if _, err := util.Decrypt(shortCipher); err == nil {
			t.Fatal("expected error for short ciphertext")
		}
	})
}
