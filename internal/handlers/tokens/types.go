package tokens

import (
	"time"
)

type AuthorizationResponse struct {
	Token string `json:"token"`
}

type CreatePATResponse struct {
	Name string    `json:"name"`
	PAT  string    `json:"pat"`
	Exp  time.Time `json:"exp"`
}
