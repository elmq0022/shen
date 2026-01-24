package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthorizationResponse struct {
	Token string `json:"token"`
}

type CreatePATResponse struct {
	Name string    `json:"name"`
	PAT  string    `json:"pat"`
	Exp  time.Time `json:"exp"`
}

type ShenClaims struct {
	jwt.RegisteredClaims
	Roles  []string `json:"roles"`
	Groups []string `json:"groups"`
}
