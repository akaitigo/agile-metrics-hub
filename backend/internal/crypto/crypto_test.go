package crypto_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/akaitigo/agile-metrics-hub/internal/crypto"
)

func generateTestKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return hex.EncodeToString(key)
}

func TestEncryptDecrypt(t *testing.T) {
	keyHex := generateTestKey(t)
	enc, err := crypto.NewEncryptor(keyHex)
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	original := "pk_12345678_ABCDEFGHIJKLMNOPQRST"
	ciphertext, err := enc.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != original {
		t.Errorf("expected %q, got %q", original, decrypted)
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	keyHex := generateTestKey(t)
	enc, err := crypto.NewEncryptor(keyHex)
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	ct1, _ := enc.Encrypt("same-plaintext")
	ct2, _ := enc.Encrypt("same-plaintext")

	if hex.EncodeToString(ct1) == hex.EncodeToString(ct2) {
		t.Error("encrypting the same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestNewEncryptor_InvalidKey(t *testing.T) {
	_, err := crypto.NewEncryptor("tooshort")
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	keyHex := generateTestKey(t)
	enc, err := crypto.NewEncryptor(keyHex)
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	_, err = enc.Decrypt([]byte("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid ciphertext")
	}
}
