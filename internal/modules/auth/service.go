package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	appErrors "github.com/leora/leora-server/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// RegisterPayload captures the fields supplied by the client during registration.
type RegisterPayload struct {
	Email           string
	FullName        string
	Password        string
	ConfirmPassword string
	Region          string
	Currency        string
}

// Service orchestrates auth flows.
type Service struct {
	repo             Repository
	tokenStore       TokenStore
	jwtSecret        string
	accessTTL        time.Duration
	refreshTTL       time.Duration
	googleClientID   string
	appleBundleID    string
}

// NewService initializes the auth service with required dependencies.
func NewService(repo Repository, store TokenStore, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		tokenStore: store,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// SetGoogleClientID configures the expected audience for Google ID tokens.
func (s *Service) SetGoogleClientID(clientID string) {
	s.googleClientID = clientID
}

// SetAppleBundleID configures the expected audience for Apple ID tokens.
func (s *Service) SetAppleBundleID(bundleID string) {
	s.appleBundleID = bundleID
}

// Register registers a new user and returns tokens.
func (s *Service) Register(ctx context.Context, payload RegisterPayload) (*User, *Tokens, error) {
	if strings.TrimSpace(payload.Password) == "" {
		return nil, nil, appErrors.InvalidUserData
	}
	if payload.Password != payload.ConfirmPassword {
		return nil, nil, appErrors.InvalidUserData
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	now := time.Now().UTC().Format(time.RFC3339)
	user := &User{
		ID:              uuid.NewString(),
		Email:           normalizeEmail(payload.Email),
		FullName:        strings.TrimSpace(payload.FullName),
		Region:          strings.TrimSpace(payload.Region),
		PrimaryCurrency: strings.TrimSpace(payload.Currency),
		Role:            RoleUser,
		Status:          "active",
		PasswordHash:    string(hashed),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	user.Permissions = PermissionsForRole(user.Role)

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(user)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	return s.sanitize(user), tokens, nil
}

// Login verifies credentials and returns the user with tokens.
func (s *Service) Login(ctx context.Context, target, password string) (*User, *Tokens, error) {
	user, err := s.repo.FindByEmail(ctx, normalizeEmail(target))
	if err != nil {
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, appErrors.InvalidCredentials
	}

	now := time.Now().UTC().Format(time.RFC3339)
	user.LastLoginAt = now
	user.UpdatedAt = now
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(user)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	return s.sanitize(user), tokens, nil
}

// Refresh returns a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	userID, err := s.tokenStore.ConsumeRefreshToken(refreshToken)
	if err != nil {
		switch err {
		case ErrRefreshTokenExpired:
			return nil, appErrors.RefreshTokenExpired
		case ErrRefreshTokenNotFound:
			return nil, appErrors.InvalidRefreshToken
		default:
			return nil, appErrors.InvalidRefreshToken
		}
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(user)
}

// ForgotPassword returns acknowledgement after validating the email exists.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	_, err := s.repo.FindByEmail(ctx, normalizeEmail(email))
	return err
}

// ResetPassword simulates a password reset.
func (s *Service) ResetPassword(ctx context.Context, email, otp, password string) error {
	if strings.TrimSpace(password) == "" {
		return appErrors.InvalidUserData
	}
	_, err := s.repo.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		return err
	}
	return nil
}

// Logout invalidates both access and refresh tokens.
func (s *Service) Logout(ctx context.Context, token, userID string) error {
	s.tokenStore.BlacklistAccessToken(token, time.Now().Add(s.accessTTL))
	s.tokenStore.RevokeRefreshTokensForUser(userID)
	return nil
}

// Profile returns the currently logged in user.
func (s *Service) Profile(ctx context.Context, userID string) (*User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.sanitize(user), nil
}

// ValidateAccessToken ensures the token is valid and not revoked.
func (s *Service) ValidateAccessToken(token string) (*TokenClaims, error) {
	if s.tokenStore.IsAccessTokenBlacklisted(token) {
		log.Printf("[auth] access token blacklisted (len=%d)", len(token))
		return nil, appErrors.InvalidToken
	}

	parsed, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[auth] token signing method invalid: %v", t.Header["alg"])
			return nil, appErrors.InvalidToken
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Printf("[auth] access token expired: %v", err)
			return nil, appErrors.TokenExpired
		}
		log.Printf("[auth] access token parse error: %v", err)
		return nil, appErrors.InvalidToken
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok || !parsed.Valid {
		log.Printf("[auth] token claims type invalid or token not valid")
		return nil, appErrors.InvalidToken
	}
	return claims, nil
}

// ExtractUserIDFromToken parses the JWT and returns the userID without
// checking the blacklist or expiry. Used by logout so that already-blacklisted
// or expired tokens can still be processed.
func (s *Service) ExtractUserIDFromToken(token string) (string, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	parsed, _, err := parser.ParseUnverified(token, &TokenClaims{})
	if err != nil {
		return "", err
	}
	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok || claims.UserID == "" {
		return "", errors.New("invalid token claims")
	}
	return claims.UserID, nil
}

func (s *Service) issueTokens(user *User) (*Tokens, error) {
	if user.Permissions == nil {
		user.Permissions = PermissionsForRole(user.Role)
	}

	now := time.Now().UTC()
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:      user.ID,
		Role:        user.Role,
		Permissions: user.Permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, appErrors.InternalServerError
	}

	refreshToken := uuid.NewString()
	s.tokenStore.SaveRefreshToken(refreshToken, refreshRecord{
		userID:    user.ID,
		expiresAt: now.Add(s.refreshTTL),
	})

	return &Tokens{
		AccessToken:  signed,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTTL.Seconds()),
	}, nil
}

// GoogleLoginPayload carries the ID token sent by the iOS client.
type GoogleLoginPayload struct {
	IDToken  string
	Region   string
	Currency string
	Mode     string // "register" or "login" (empty = login)
}

// GoogleLogin verifies a Google ID token and returns or creates the matching user.
func (s *Service) GoogleLogin(ctx context.Context, payload GoogleLoginPayload) (*User, *Tokens, error) {
	claims, err := verifyGoogleIDToken(payload.IDToken, s.googleClientID)
	if err != nil {
		return nil, nil, appErrors.InvalidGoogleToken
	}

	sub := claims.Subject
	email := claims.Email
	name := claims.Name

	// 1. Check if we already have this Google identity linked.
	identity, err := s.repo.FindIdentity(ctx, "google", sub)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	var user *User

	if identity != nil {
		if payload.Mode == "register" {
			return nil, nil, appErrors.UserAlreadyExists
		}
		// Returning user — fetch from DB.
		user, err = s.repo.FindByID(ctx, identity.UserID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// New Google sign-in — check if email already exists.
		existing, _ := s.repo.FindByEmail(ctx, normalizeEmail(email))
		if existing != nil {
			if payload.Mode == "register" {
				return nil, nil, appErrors.UserAlreadyExists
			}
			user = existing
		} else {
			if payload.Mode == "login" {
				return nil, nil, appErrors.UserNotFound
			}
			// Create a brand-new user (no password).
			now := time.Now().UTC().Format(time.RFC3339)
			region := strings.TrimSpace(payload.Region)
			if region == "" {
				region = "us"
			}
			currency := strings.TrimSpace(payload.Currency)
			if currency == "" {
				currency = "USD"
			}
			user = &User{
				ID:              uuid.NewString(),
				Email:           normalizeEmail(email),
				FullName:        name,
				Region:          region,
				PrimaryCurrency: currency,
				Role:            RoleUser,
				Status:          "active",
				PasswordHash:    "", // social-only account
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			user.Permissions = PermissionsForRole(user.Role)
			if err := s.repo.CreateUser(ctx, user); err != nil {
				return nil, nil, err
			}
		}

		// Link the Google identity to the user.
		newIdentity := &AuthIdentity{
			ID:         uuid.NewString(),
			UserID:     user.ID,
			Provider:   "google",
			ProviderID: sub,
			Email:      email,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		}
		if err := s.repo.CreateIdentity(ctx, newIdentity); err != nil {
			return nil, nil, appErrors.InternalServerError
		}
	}

	// Update last login.
	now := time.Now().UTC().Format(time.RFC3339)
	user.LastLoginAt = now
	user.UpdatedAt = now
	_ = s.repo.UpdateUser(ctx, user)

	tokens, err := s.issueTokens(user)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	return s.sanitize(user), tokens, nil
}

// ── Google ID-token verification ──────────────────────────────────────

type googleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Google's public JWKS endpoint.
const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

var (
	googleKeysMu    sync.RWMutex
	googleKeysCache map[string]*rsa.PublicKey
	googleKeysTTL   time.Time
)

func verifyGoogleIDToken(idToken, expectedAudience string) (*googleClaims, error) {
	keys, err := fetchGoogleKeys()
	if err != nil {
		return nil, fmt.Errorf("fetch google keys: %w", err)
	}

	token, err := jwt.ParseWithClaims(idToken, &googleClaims{}, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid header")
		}
		key, exists := keys[kid]
		if !exists {
			return nil, fmt.Errorf("unknown kid %q", kid)
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse id token: %w", err)
	}

	claims, ok := token.Claims.(*googleClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Audience check.
	if expectedAudience != "" {
		aud, _ := claims.GetAudience()
		found := false
		for _, a := range aud {
			if a == expectedAudience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("audience mismatch")
		}
	}

	// Issuer check.
	iss, _ := claims.GetIssuer()
	if iss != "accounts.google.com" && iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("issuer mismatch: %s", iss)
	}

	return claims, nil
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
}

func fetchGoogleKeys() (map[string]*rsa.PublicKey, error) {
	googleKeysMu.RLock()
	if googleKeysCache != nil && time.Now().Before(googleKeysTTL) {
		defer googleKeysMu.RUnlock()
		return googleKeysCache, nil
	}
	googleKeysMu.RUnlock()

	resp, err := http.Get(googleCertsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jwks jwksResponse
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := int(new(big.Int).SetBytes(eBytes).Int64())
		keys[k.Kid] = &rsa.PublicKey{N: n, E: e}
	}

	googleKeysMu.Lock()
	googleKeysCache = keys
	googleKeysTTL = time.Now().Add(1 * time.Hour)
	googleKeysMu.Unlock()

	return keys, nil
}

// ── Apple Sign-In ─────────────────────────────────────────────────────

// AppleLoginPayload carries the identity token from the iOS client.
type AppleLoginPayload struct {
	IdentityToken string
	Email         string // only provided on first sign-in
	FullName      string // only provided on first sign-in
	Region        string
	Currency      string
	Mode          string // "register" or "login" (empty = login)
}

// AppleLogin verifies an Apple identity token and returns or creates the matching user.
func (s *Service) AppleLogin(ctx context.Context, payload AppleLoginPayload) (*User, *Tokens, error) {
	claims, err := verifyAppleIDToken(payload.IdentityToken, s.appleBundleID)
	if err != nil {
		return nil, nil, appErrors.InvalidAppleToken
	}

	sub := claims.Subject
	email := claims.Email
	if email == "" {
		email = payload.Email
	}
	name := strings.TrimSpace(payload.FullName)

	// 1. Check if we already have this Apple identity linked.
	identity, err := s.repo.FindIdentity(ctx, "apple", sub)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	var user *User

	if identity != nil {
		if payload.Mode == "register" {
			return nil, nil, appErrors.UserAlreadyExists
		}
		user, err = s.repo.FindByID(ctx, identity.UserID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// New Apple sign-in — check if email already exists.
		if email != "" {
			existing, _ := s.repo.FindByEmail(ctx, normalizeEmail(email))
			if existing != nil {
				if payload.Mode == "register" {
					return nil, nil, appErrors.UserAlreadyExists
				}
				user = existing
			}
		}

		if user == nil {
			if payload.Mode == "login" {
				return nil, nil, appErrors.UserNotFound
			}
			now := time.Now().UTC().Format(time.RFC3339)
			region := strings.TrimSpace(payload.Region)
			if region == "" {
				region = "us"
			}
			currency := strings.TrimSpace(payload.Currency)
			if currency == "" {
				currency = "USD"
			}
			if name == "" {
				name = "Apple User"
			}
			user = &User{
				ID:              uuid.NewString(),
				Email:           normalizeEmail(email),
				FullName:        name,
				Region:          region,
				PrimaryCurrency: currency,
				Role:            RoleUser,
				Status:          "active",
				PasswordHash:    "",
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			user.Permissions = PermissionsForRole(user.Role)
			if err := s.repo.CreateUser(ctx, user); err != nil {
				return nil, nil, err
			}
		}

		newIdentity := &AuthIdentity{
			ID:         uuid.NewString(),
			UserID:     user.ID,
			Provider:   "apple",
			ProviderID: sub,
			Email:      email,
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		}
		if err := s.repo.CreateIdentity(ctx, newIdentity); err != nil {
			return nil, nil, appErrors.InternalServerError
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	user.LastLoginAt = now
	user.UpdatedAt = now
	_ = s.repo.UpdateUser(ctx, user)

	tokens, err := s.issueTokens(user)
	if err != nil {
		return nil, nil, appErrors.InternalServerError
	}

	return s.sanitize(user), tokens, nil
}

// ── Apple ID-token verification ───────────────────────────────────────

type appleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"` // Apple sends bool or string
}

const appleCertsURL = "https://appleid.apple.com/auth/keys"

var (
	appleKeysMu    sync.RWMutex
	appleKeysCache map[string]*rsa.PublicKey
	appleKeysTTL   time.Time
)

func verifyAppleIDToken(idToken, expectedAudience string) (*appleClaims, error) {
	keys, err := fetchAppleKeys()
	if err != nil {
		return nil, fmt.Errorf("fetch apple keys: %w", err)
	}

	token, err := jwt.ParseWithClaims(idToken, &appleClaims{}, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid header")
		}
		key, exists := keys[kid]
		if !exists {
			return nil, fmt.Errorf("unknown kid %q", kid)
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse id token: %w", err)
	}

	claims, ok := token.Claims.(*appleClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Audience check (must match bundle ID).
	if expectedAudience != "" {
		aud, _ := claims.GetAudience()
		found := false
		for _, a := range aud {
			if a == expectedAudience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("audience mismatch")
		}
	}

	// Issuer check.
	iss, _ := claims.GetIssuer()
	if iss != "https://appleid.apple.com" {
		return nil, fmt.Errorf("issuer mismatch: %s", iss)
	}

	return claims, nil
}

func fetchAppleKeys() (map[string]*rsa.PublicKey, error) {
	appleKeysMu.RLock()
	if appleKeysCache != nil && time.Now().Before(appleKeysTTL) {
		defer appleKeysMu.RUnlock()
		return appleKeysCache, nil
	}
	appleKeysMu.RUnlock()

	resp, err := http.Get(appleCertsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jwks jwksResponse
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := int(new(big.Int).SetBytes(eBytes).Int64())
		keys[k.Kid] = &rsa.PublicKey{N: n, E: e}
	}

	appleKeysMu.Lock()
	appleKeysCache = keys
	appleKeysTTL = time.Now().Add(1 * time.Hour)
	appleKeysMu.Unlock()

	return keys, nil
}

func (s *Service) sanitize(user *User) *User {
	if user == nil {
		return nil
	}
	copy := *user
	copy.PasswordHash = ""
	return &copy
}
