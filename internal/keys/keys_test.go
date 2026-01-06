package keys_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"strings"
	"testing"

	"github.com/elmq0022/shen/internal/keys"
	"github.com/oklog/ulid/v2"
)

func TestGenerateAndEncryptJWTKey(t *testing.T) {
	kek := make([]byte, 32)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("failed to generate test KEK: %v", err)
	}

	kid, encryptedPrivatePEM, publicPEM, err := keys.GenerateAndEncryptJWTKey(kek)
	if err != nil {
		t.Fatalf("GenerateAndEncryptJWTKey() failed: %v", err)
	}

	if kid == "" {
		t.Error("kid should not be empty")
	}
	_, parseErr := ulid.Parse(kid)
	if parseErr != nil {
		t.Errorf("kid %q is not valid ULID format: %v", kid, parseErr)
	}

	if len(encryptedPrivatePEM) == 0 {
		t.Error("encryptedPrivatePEM should not be empty")
	}
	if len(encryptedPrivatePEM) < 128 {
		t.Errorf("encryptedPrivatePEM suspiciously short: got %d bytes", len(encryptedPrivatePEM))
	}

	if publicPEM == "" {
		t.Error("publicPEM should not be empty")
	}
	if !strings.Contains(publicPEM, "BEGIN PUBLIC KEY") {
		t.Error("publicPEM should contain PEM header")
	}

	block, _ := pem.Decode([]byte(publicPEM))
	if block == nil {
		t.Fatal("failed to decode public key PEM")
	}
	if block.Type != "PUBLIC KEY" {
		t.Errorf("expected PEM type 'PUBLIC KEY', got %q", block.Type)
	}
	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Errorf("failed to parse public key: %v", err)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	kek := make([]byte, 32)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("failed to generate test KEK: %v", err)
	}

	kid, encryptedPrivatePEM, publicPEM, err := keys.GenerateAndEncryptJWTKey(kek)
	if err != nil {
		t.Fatalf("GenerateAndEncryptJWTKey() failed: %v", err)
	}

	decryptedPrivatePEM, err := keys.DecryptPrivateKey(encryptedPrivatePEM, kek)
	if err != nil {
		t.Fatalf("DecryptPrivateKey() failed: %v", err)
	}

	if len(decryptedPrivatePEM) == 0 {
		t.Fatal("decrypted private key should not be empty")
	}
	if !strings.Contains(string(decryptedPrivatePEM), "BEGIN PRIVATE KEY") {
		t.Error("decrypted private key should contain PEM header")
	}

	block, _ := pem.Decode(decryptedPrivatePEM)
	if block == nil {
		t.Fatal("failed to decode decrypted private key PEM")
	}
	if block.Type != "PRIVATE KEY" {
		t.Errorf("expected PEM type 'PRIVATE KEY', got %q", block.Type)
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse decrypted private key: %v", err)
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("expected *rsa.PrivateKey, got %T", parsedKey)
	}

	if rsaKey.N.BitLen() != 2048 {
		t.Errorf("expected 2048-bit key, got %d bits", rsaKey.N.BitLen())
	}

	publicBlock, _ := pem.Decode([]byte(publicPEM))
	if publicBlock == nil {
		t.Fatal("failed to decode public key PEM")
	}
	parsedPublicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	expectedPublicKey := &rsaKey.PublicKey
	actualPublicKey, ok := parsedPublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("expected *rsa.PublicKey, got %T", parsedPublicKey)
	}

	if expectedPublicKey.N.Cmp(actualPublicKey.N) != 0 {
		t.Error("public key modulus does not match private key")
	}
	if expectedPublicKey.E != actualPublicKey.E {
		t.Error("public key exponent does not match private key")
	}

	if kid == "" {
		t.Error("kid should not be empty")
	}
}

func TestLoadKEK_Success(t *testing.T) {
	expectedKEK := make([]byte, 32)
	if _, err := rand.Read(expectedKEK); err != nil {
		t.Fatalf("failed to generate test KEK: %v", err)
	}
	kekBase64 := base64.StdEncoding.EncodeToString(expectedKEK)
	t.Setenv("SHEN_KEY_ENCRYPTION_KEY", kekBase64)

	kek, err := keys.LoadKEK()

	if err != nil {
		t.Fatalf("LoadKEK() failed: %v", err)
	}
	if len(kek) != 32 {
		t.Errorf("expected 32-byte KEK, got %d bytes", len(kek))
	}
	if string(kek) != string(expectedKEK) {
		t.Error("loaded KEK does not match expected KEK")
	}
}

func TestLoadKEK_NoEnvVar(t *testing.T) {
	os.Unsetenv("SHEN_KEY_ENCRYPTION_KEY")

	kek, err := keys.LoadKEK()

	if err == nil {
		t.Fatal("expected error when SHEN_KEY_ENCRYPTION_KEY is not set, got nil")
	}
	if !strings.Contains(err.Error(), "SHEN_KEY_ENCRYPTION_KEY") {
		t.Errorf("error should mention SHEN_KEY_ENCRYPTION_KEY, got: %v", err)
	}
	if kek != nil {
		t.Errorf("kek should be nil on error, got %v", kek)
	}
}

func TestLoadKEK_InvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		kekValue string
		wantErr  string
	}{
		{
			name:     "not base64",
			kekValue: "not-valid-base64!!!",
			wantErr:  "failed to base64 decode",
		},
		{
			name:     "wrong length (16 bytes)",
			kekValue: base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr:  "KEK was not 32 bytes",
		},
		{
			name:     "wrong length (64 bytes)",
			kekValue: base64.StdEncoding.EncodeToString(make([]byte, 64)),
			wantErr:  "KEK was not 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHEN_KEY_ENCRYPTION_KEY", tt.kekValue)

			_, err := keys.LoadKEK()

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestDecryptPrivateKey_TamperedData(t *testing.T) {
	kek := make([]byte, 32)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("failed to generate test KEK: %v", err)
	}

	_, encryptedPrivatePEM, _, err := keys.GenerateAndEncryptJWTKey(kek)
	if err != nil {
		t.Fatalf("GenerateAndEncryptJWTKey() failed: %v", err)
	}

	tampered := make([]byte, len(encryptedPrivatePEM))
	copy(tampered, encryptedPrivatePEM)
	tampered[len(tampered)/2] ^= 0xFF

	_, err = keys.DecryptPrivateKey(tampered, kek)

	if err == nil {
		t.Fatal("expected error when decrypting tampered data, got nil")
	}
	if !strings.Contains(err.Error(), "error decrypting data") {
		t.Errorf("expected decryption error, got: %v", err)
	}
}

func TestDecryptPrivateKey_WrongKEK(t *testing.T) {
	kek1 := make([]byte, 32)
	if _, err := rand.Read(kek1); err != nil {
		t.Fatalf("failed to generate test KEK: %v", err)
	}

	_, encryptedPrivatePEM, _, err := keys.GenerateAndEncryptJWTKey(kek1)
	if err != nil {
		t.Fatalf("GenerateAndEncryptJWTKey() failed: %v", err)
	}

	kek2 := make([]byte, 32)
	if _, err := rand.Read(kek2); err != nil {
		t.Fatalf("failed to generate second KEK: %v", err)
	}

	_, err = keys.DecryptPrivateKey(encryptedPrivatePEM, kek2)

	if err == nil {
		t.Fatal("expected error when decrypting with wrong KEK, got nil")
	}
}

func TestDecryptPrivateKey_TooShort(t *testing.T) {
	kek := make([]byte, 32)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("failed to generate test KEK: %v", err)
	}

	shortData := make([]byte, 5)

	_, err := keys.DecryptPrivateKey(shortData, kek)

	if err == nil {
		t.Fatal("expected error for data shorter than nonce size, got nil")
	}
	if !strings.Contains(err.Error(), "encrypted key too short") {
		t.Errorf("expected 'too short' error, got: %v", err)
	}
}
