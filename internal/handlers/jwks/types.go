package jwks

// JWK represents a JSON Web Key (RFC 7517)
type JWK struct {
	Kty string `json:"kty"` // Key Type (e.g., "RSA")
	Use string `json:"use"` // Public Key Use (e.g., "sig" for signature)
	Kid string `json:"kid"` // Key ID
	N   string `json:"n"`   // RSA Modulus (base64url encoded)
	E   string `json:"e"`   // RSA Exponent (base64url encoded)
}

// JWKS represents a JSON Web Key Set (RFC 7517)
type JWKS struct {
	Keys []JWK `json:"keys"`
}
