package bootstrap

import (
	"context"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/keys"
)

// GenerateInitialKeys generates the initial JWT signing keys for bootstrap
// and saves them to the database. It loads the KEK from environment and
// generates a new RSA key pair. Panics if any error occurs during key
// generation, encryption, or database insertion.
//
// The generated key will be active for both signing and verification.
func GenerateInitialKeys(ctx context.Context, queries *db.Queries) {
	numKeys, err := queries.CountJWTKeys(ctx)
	if err != nil {
		panic(err)
	}
	if numKeys > 0 {
		return
	}

	kek, err := keys.LoadKEK()
	if err != nil {
		panic(err)
	}

	kid, encryptedPrivatePEM, publicPEM, err := keys.GenerateAndEncryptJWTKey(kek)
	if err != nil {
		panic(err)
	}

	_, err = queries.CreateJWTKey(ctx, db.CreateJWTKeyParams{
		Kid:                   kid,
		EncryptedPrivateKey:   encryptedPrivatePEM,
		PublicKey:             publicPEM,
		ActiveForSigning:      true,
		ActiveForVerification: true,
	})
	if err != nil {
		panic(err)
	}
}
