package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims captures the metadata stored inside access tokens.
// jwt/v5 automatically validates exp, nbf, iat via RegisteredClaims
// so no custom Valid() override is needed.
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Role        Role     `json:"role"`
	Permissions []string `json:"permissions"`
}
