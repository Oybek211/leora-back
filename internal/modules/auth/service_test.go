package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueTokens_UniquePerLogin(t *testing.T) {
	store := NewInMemoryTokenStore()
	service := NewService(nil, store, "test-secret", time.Minute, time.Hour)

	user := &User{
		ID:    "user-1",
		Role:  RoleUser,
		Email: "user@example.com",
	}

	first, err := service.issueTokens(user)
	if err != nil {
		t.Fatalf("issueTokens first: %v", err)
	}
	second, err := service.issueTokens(user)
	if err != nil {
		t.Fatalf("issueTokens second: %v", err)
	}

	if first.AccessToken == second.AccessToken {
		t.Fatalf("expected unique access tokens, got identical tokens")
	}

	firstClaims := &TokenClaims{}
	_, err = jwt.ParseWithClaims(first.AccessToken, firstClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("parse first token: %v", err)
	}

	secondClaims := &TokenClaims{}
	_, err = jwt.ParseWithClaims(second.AccessToken, secondClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("parse second token: %v", err)
	}

	if firstClaims.ID == "" || secondClaims.ID == "" {
		t.Fatalf("expected jti to be set on access tokens")
	}
	if firstClaims.ID == secondClaims.ID {
		t.Fatalf("expected unique jti values, got identical IDs")
	}
}
