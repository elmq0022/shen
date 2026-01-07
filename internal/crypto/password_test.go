package crypto_test

import (
	"testing"

	"github.com/elmq0022/shen/internal/crypto"
)

func TestPasswordRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mySecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "simple password",
			password: "password",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "thisIsAVeryLongPasswordWithLotsOfCharacters1234567890!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "password with spaces",
			password: "my password has spaces",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the password
			hashed, err := crypto.HashedPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashedPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify the password matches
			match, err := crypto.CheckPassword(tt.password, hashed)
			if err != nil {
				t.Errorf("CheckPassword() unexpected error = %v", err)
				return
			}

			if !match {
				t.Errorf("CheckPassword() = %v, want true for correct password", match)
			}

			// Verify wrong password doesn't match
			wrongMatch, err := crypto.CheckPassword("wrongPassword", hashed)
			if err != nil {
				t.Errorf("CheckPassword() with wrong password unexpected error = %v", err)
				return
			}

			if wrongMatch {
				t.Errorf("CheckPassword() = %v, want false for incorrect password", wrongMatch)
			}
		})
	}
}

func TestCheckPasswordInvalidFormat(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		hashedFormat string
		wantErr      bool
	}{
		{
			name:         "no separator",
			password:     "password",
			hashedFormat: "invalidsalt",
			wantErr:      true,
		},
		{
			name:         "invalid hex in salt",
			password:     "password",
			hashedFormat: "zzz.123456",
			wantErr:      true,
		},
		{
			name:         "invalid hex in hash",
			password:     "password",
			hashedFormat: "123456.zzz",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := crypto.CheckPassword(tt.password, tt.hashedFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordHashUniqueness(t *testing.T) {
	password := "samePassword"

	hash1, err := crypto.HashedPassword(password)
	if err != nil {
		t.Fatalf("HashedPassword() error = %v", err)
	}

	hash2, err := crypto.HashedPassword(password)
	if err != nil {
		t.Fatalf("HashedPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashedPassword() produced identical hashes for same password (salts should be unique)")
	}

	// Both hashes should still verify the same password
	match1, err := crypto.CheckPassword(password, hash1)
	if err != nil || !match1 {
		t.Errorf("First hash failed to verify: match=%v, err=%v", match1, err)
	}

	match2, err := crypto.CheckPassword(password, hash2)
	if err != nil || !match2 {
		t.Errorf("Second hash failed to verify: match=%v, err=%v", match2, err)
	}
}
