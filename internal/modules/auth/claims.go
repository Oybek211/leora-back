package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// TokenClaims captures the metadata stored inside access tokens.
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Role        Role     `json:"role"`
	Permissions []string `json:"permissions"`
}

// Valid enforces registered claim boundaries such as expiration.
func (c *TokenClaims) Valid() error {
	if c == nil {
		return appErrors.InvalidToken
	}
	now := time.Now()
	if c.NotBefore != nil && now.Before(c.NotBefore.Time) {
		return appErrors.InvalidToken
	}
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Time) {
		return appErrors.TokenExpired
	}
	return nil
}
