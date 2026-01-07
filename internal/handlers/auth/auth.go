package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) Handler {
	return Handler{
		queries: queries,
	}
}

func (h *Handler) Login(c echo.Context) error {

	var lr LoginRequest
	if err := c.Bind(&lr); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	user, err := h.queries.GetUserByUsername(c.Request().Context(), lr.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid credentials"))
	}

	if !user.Active {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("Forbidden"))
	}

	role, err := h.queries.GetRoleByName(c.Request().Context(), "service") // there's probably a better way
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Internal server error"))
	}
	if user.Role == role.ID {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("Forbidden"))
	}

	authorized, err := crypto.CheckPassword(lr.Password, user.HashedPassword.String)
	if err != nil || !authorized {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid credentials"))
	}

	token, err := crypto.GenerateSessionToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to generate token"))
	}

	hashedToken := crypto.HashToken(token)

	// TODO: configure token expiry
	_, err = h.queries.CreateSession(c.Request().Context(), db.CreateSessionParams{
		HashedToken: hashedToken,
		UserID:      user.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(30 * 24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to create session"))
	}

	return c.JSON(http.StatusOK, LoginResponse{SessionToken: token})
}

func (h *Handler) Logout(c echo.Context) error {
	var lr LogoutRequest
	if err := c.Bind(&lr); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	hashedToken := crypto.HashToken(lr.SessionToken)

	// Revoke the session by hashed token
	err := h.queries.RevokeSessionByHashedToken(c.Request().Context(), hashedToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to revoke session"))
	}

	return c.JSON(http.StatusOK, LogoutResponse{Message: "Successfully logged out"})
}

// GetJWKS returns the JSON Web Key Set (JWKS) containing all active verification keys.
// This endpoint is used by applications to fetch Shen's public keys for JWT verification.
//
// Endpoint: GET /.well-known/jwks.json
//
// Response format follows RFC 7517 (JSON Web Key)
func (h *Handler) GetJWKS(c echo.Context) error {

	keys, err := h.queries.ListActiveVerificationKeys(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to retrieve keys"))
	}

	jwks := JWKS{
		Keys: make([]JWK, 0, len(keys)),
	}

	for _, key := range keys {
		jwk, err := pemToJWK(key.Kid, key.PublicKey)
		if err != nil {
			// Log error but continue with other keys
			continue
		}
		jwks.Keys = append(jwks.Keys, jwk)
	}

	return c.JSON(http.StatusOK, jwks)
}

// pemToJWK converts a PEM-encoded RSA public key to JWK format
func pemToJWK(kid, pemPublicKey string) (JWK, error) {
	block, _ := pem.Decode([]byte(pemPublicKey))
	if block == nil {
		return JWK{}, fmt.Errorf("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return JWK{}, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return JWK{}, fmt.Errorf("public key is not RSA")
	}

	// Convert RSA modulus (N) and exponent (E) to base64url encoding
	// Per RFC 7518 Section 6.3: N and E are base64url-encoded
	nBytes := rsaPubKey.N.Bytes()
	eBytes := big.NewInt(int64(rsaPubKey.E)).Bytes()

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}, nil
}
