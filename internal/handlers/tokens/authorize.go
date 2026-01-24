package tokens

import (
	"net/http"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Authorize validates a PAT (Personal Access Token) and returns a short-lived JWT.
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
func (h *Handler) Authorize(c echo.Context) error {
	token := c.Get("pat").(db.ShenToken)
	ctx := c.Request().Context()

	// Get user's roles for this application
	roles, err := h.Queries.GetUserApplicationRoles(ctx, db.GetUserApplicationRolesParams{
		UserID:        token.UserID,
		ApplicationID: token.ApplicationID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to get user roles"))
	}
	if len(roles) == 0 {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("user has no roles for this application"))
	}

	// Get user's active groups
	groups, err := h.Queries.GetUserGroups(ctx, token.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to get user groups"))
	}

	// Extract role and group names for the JWT
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	groupNames := make([]string, len(groups))
	for i, g := range groups {
		groupNames[i] = g.Name
	}

	user, err := h.Queries.GetUserByID(ctx, token.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to find user"))
	}

	app, err := h.Queries.GetApplicationByID(ctx, token.ApplicationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to find application"))
	}
	if !app.Active {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("application is not active"))
	}

	now := time.Now().UTC()
	claims := ShenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "shen",
			Subject:   user.Username,
			Audience:  jwt.ClaimStrings{app.Name},
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Roles:  roleNames,
		Groups: groupNames,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := t.SignedString(h.privateKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to sign token"))
	}

	return c.JSON(http.StatusOK, AuthorizationResponse{Token: signedToken})
}

type ShenClaims struct {
	jwt.RegisteredClaims
	Roles  []string `json:"roles"`
	Groups []string `json:"groups"`
}
