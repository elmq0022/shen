package crypto_test

import (
	"encoding/hex"
	"testing"

	"github.com/elmq0022/shen/internal/crypto"
)

func TestGenerateSessionToken(t *testing.T) {
	token, err := crypto.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken() error = %v", err)
	}

	// Token should be 64 characters (32 bytes hex-encoded)
	if len(token) != 64 {
		t.Errorf("GenerateSessionToken() length = %d, want 64", len(token))
	}

	// Should be valid hex
	_, err = hex.DecodeString(token)
	if err != nil {
		t.Errorf("GenerateSessionToken() produced invalid hex: %v", err)
	}
}

func TestGenerateSessionTokenUniqueness(t *testing.T) {
	token1, err := crypto.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken() error = %v", err)
	}

	token2, err := crypto.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken() error = %v", err)
	}

	if token1 == token2 {
		t.Error("GenerateSessionToken() produced identical tokens (should be unique)")
	}
}

func TestHashToken(t *testing.T) {
	token := "test-session-token"
	hash := crypto.HashToken(token)

	// SHA-256 hash should be 64 characters (32 bytes hex-encoded)
	if len(hash) != 64 {
		t.Errorf("HashToken() length = %d, want 64", len(hash))
	}

	// Should be valid hex
	_, err := hex.DecodeString(hash)
	if err != nil {
		t.Errorf("HashToken() produced invalid hex: %v", err)
	}

	// Same input should produce same hash
	hash2 := crypto.HashToken(token)
	if hash != hash2 {
		t.Error("HashToken() produced different hashes for same input")
	}

	// Different input should produce different hash
	hash3 := crypto.HashToken("different-token")
	if hash == hash3 {
		t.Error("HashToken() produced same hash for different inputs")
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "known value 1",
			token: "test",
			want:  "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			name:  "known value 2",
			token: "session",
			want:  "3f3af1ecebbd1410ab417ec0d27bbfcb5d340e177ae159b59fc8626c2dfd9175",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := crypto.HashToken(tt.token)
			if got != tt.want {
				t.Errorf("HashToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
