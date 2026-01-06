package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/oklog/ulid/v2"
)

// GenerateAndEncryptJWTKey generates a new RSA-2048 key pair for JWT signing,
// encrypts the private key with AES-256-GCM using the provided KEK,
// and returns all components needed for database storage.
//
// Parameters:
//   - kek: 32-byte key encryption key (from LoadKEK or equivalent)
//
// Returns:
//   - kid: unique key identifier (ULID format - opaque, sortable by time)
//   - encryptedPrivatePEM: encrypted private key in PEM format (PKCS#8)
//   - publicPEM: public key in PEM format (PKIX)
//   - error: any error encountered during generation or encryption
func GenerateAndEncryptJWTKey(kek []byte) (string, []byte, string, error) {
	privatePEM, publicPEM, err := generateRSAKeyPair()
	if err != nil {
		return "", nil, "", err
	}

	encryptedPrivatePEM, err := encryptPrivateKey(privatePEM, kek)
	if err != nil {
		return "", nil, "", err
	}

	kid := generateKID()

	return kid, encryptedPrivatePEM, publicPEM, nil
}

// DecryptPrivateKey decrypts an encrypted private key using AES-256-GCM.
// The encryptedKey must be in the format produced by EncryptPrivateKey:
// [12-byte nonce][ciphertext][16-byte auth tag].
//
// Parameters:
//   - encryptedKey: encrypted private key bytes from database
//   - kek: 32-byte key encryption key (from LoadKEK or equivalent)
//
// Returns the decrypted private key in PEM format, or an error if decryption
// fails (including authentication failures from tampered data).
//
// Reference: https://www.twilio.com/en-us/blog/developers/community/encrypt-and-decrypt-data-in-go-with-aes-256
func DecryptPrivateKey(encryptedKey []byte, kek []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("error creating aes block cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error setting gcm mode: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedKey) < nonceSize {
		return nil, fmt.Errorf("encrypted key too short: got %d bytes, need at least %d", len(encryptedKey), nonceSize)
	}

	decryptedKey, err := gcm.Open(nil, encryptedKey[:nonceSize], encryptedKey[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data: %w", err)
	}

	return decryptedKey, nil
}

// LoadKEK loads and validates the Key Encryption Key (KEK) from the
// SHEN_KEY_ENCRYPTION_KEY environment variable.
//
// Returns:
//   - kek: 32-byte key encryption key decoded from base64
//   - error: error if the environment variable is not set, cannot be decoded,
//     or is not exactly 32 bytes
//
// The KEK must be set in the SHEN_KEY_ENCRYPTION_KEY environment variable
// as a base64-encoded 32-byte key.
func LoadKEK() ([]byte, error) {
	skek, ok := os.LookupEnv("SHEN_KEY_ENCRYPTION_KEY")
	if !ok {
		return nil, fmt.Errorf("Could not load SHEN_KEY_ENCRYPTION_KEY")
	}

	kek, err := base64.StdEncoding.DecodeString(skek)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode SHEN_KEY_ENCRYPTION_KEY: %w", err)
	}

	if len(kek) != 32 {
		return nil, fmt.Errorf("KEK was not 32 bytes, got %d bytes", len(kek))
	}

	return kek, nil
}

func encryptPrivateKey(privatePEM []byte, kek []byte) ([]byte, error) {
	// ref: https://www.twilio.com/en-us/blog/developers/community/encrypt-and-decrypt-data-in-go-with-aes-256

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("error creating aes block cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error setting gcm mode: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error generating the nonce: %w", err)
	}

	encryptedKey := gcm.Seal(nonce, nonce, privatePEM, nil)

	return encryptedKey, nil
}

func generateKID() string {
	return ulid.Make().String()
}

func generateRSAKeyPair() ([]byte, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	mPrivateKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("Error marshalling RSA private key: %w", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: mPrivateKey,
	})

	mPublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, "", fmt.Errorf("Error marshalling RSA public key: %w", err)
	}

	publicPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: mPublicKey,
	}))

	return privatePEM, publicPEM, nil
}
