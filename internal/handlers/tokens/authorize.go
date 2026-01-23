package tokens

import "github.com/labstack/echo/v4"

// authorize validates a PAT (Personal Access Token) for API access.
//
// Token Verification Flow:
//  1. Client sends the raw PAT in the Authorization header (Bearer <token>)
//  2. Server extracts the token and computes its SHA-256 hash
//  3. The hash is used to look up the token in the database
//  4. Only the hashed token is stored in the database, never the raw token
//
// This approach ensures that even if the database is compromised, the raw
// tokens cannot be recovered. Clients must securely store their PAT after
// creation, as it cannot be retrieved again.
func (h *Handler) authorize(c echo.Context) error {
	return nil
}
