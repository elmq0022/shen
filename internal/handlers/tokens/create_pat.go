package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreatePAT(c echo.Context) error {
	name := strings.ToLower(c.Param("name"))
	appName := c.Param("application")
	user := c.Get("user").(db.ShenUser)

	app, err := h.Queries.GetApplicationByName(c.Request().Context(), appName)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("application not found"))
	}

	roles, err := h.Queries.GetUserApplicationRoles(c.Request().Context(), db.GetUserApplicationRolesParams{
		UserID:        user.ID,
		ApplicationID: app.ID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to check user authorization"))
	}
	if len(roles) == 0 {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("user not authorized for this application"))
	}

	_, err = h.Queries.GetTokenByUserApplicationName(c.Request().Context(), db.GetTokenByUserApplicationNameParams{
		UserID:        user.ID,
		ApplicationID: app.ID,
		Name:          name,
	})
	if err == nil {
		return c.JSON(http.StatusConflict, handlers.NewErrorResponse("token name already exists for this application"))
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to check for existing token"))
	}

	pat, err := GeneratePAT()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to generate token"))
	}

	hashedPAT := crypto.HashToken(pat)

	var exp time.Time
	expString := c.QueryParam("exp")
	if expString != "" {
		exp, err = time.Parse(time.RFC3339, expString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid expiration format, expected ISO 8601"))
		}
		if exp.After(time.Now().AddDate(0, 6, 0)) {
			return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("expiration cannot exceed 6 months"))
		}
	} else {
		exp = time.Now().AddDate(0, 0, 30).UTC()
	}

	_, err = h.Queries.CreateToken(c.Request().Context(), db.CreateTokenParams{
		Name:          name,
		HashedToken:   hashedPAT,
		UserID:        user.ID,
		ApplicationID: app.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  exp,
			Valid: true,
		},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusConflict, handlers.NewErrorResponse("token name already exists for this application"))
		}
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to create token"))
	}

	return c.JSON(http.StatusOK, CreatePATResponse{
		Name: name,
		PAT:  pat,
		Exp:  exp,
	})
}

func GeneratePAT() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(bytes)
	return "shen_pat_" + token, nil
}
