package auth

import (
	"context"
	"errors"
	"strings"
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
	repo        Repository
	tokenStore  TokenStore
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
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
		return nil, appErrors.InvalidToken
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
		return nil, appErrors.InvalidToken
	}

	parsed, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, appErrors.InvalidToken
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, appErrors.TokenExpired
		}
		return nil, appErrors.InvalidToken
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok {
		return nil, appErrors.InvalidToken
	}
	if err := claims.Valid(); err != nil {
		if errors.Is(err, appErrors.TokenExpired) {
			return nil, appErrors.TokenExpired
		}
		return nil, appErrors.InvalidToken
	}
	return claims, nil
}

func (s *Service) issueTokens(user *User) (*Tokens, error) {
	if user.Permissions == nil {
		user.Permissions = PermissionsForRole(user.Role)
	}

	now := time.Now().UTC()
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
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

func (s *Service) sanitize(user *User) *User {
	if user == nil {
		return nil
	}
	copy := *user
	copy.PasswordHash = ""
	return &copy
}
