package tokens

import (
	"time"
)

type CreatePATResponse struct {
	Name string    `json:"name"`
	PAT  string    `json:"pat"`
	Exp  time.Time `json:"exp"`
}
