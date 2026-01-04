package users

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appErrors "github.com/leora/leora-server/internal/errors"
	"github.com/leora/leora-server/internal/modules/auth"
	"github.com/redis/go-redis/v9"
)

// Service orchestrates user profile logic.
type Service struct {
	authService *auth.Service
	repo        auth.Repository
	cache       *redis.Client
}

// ListOptions exposes filters that power GET /users.
type ListOptions struct {
	Role      auth.Role
	Status    string
	Search    string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

type cachedUsers struct {
	Items []*auth.User `json:"items"`
	Total int          `json:"total"`
}

const userListCacheTTL = 30 * time.Second

// NewService constructs a new service.
func NewService(authService *auth.Service, repo auth.Repository, cache *redis.Client) *Service {
	return &Service{authService: authService, repo: repo, cache: cache}
}

// ListUsers returns filtered users with pagination metadata.
func (s *Service) ListUsers(ctx context.Context, opts ListOptions) ([]*auth.User, int, error) {
	cacheKey := fmt.Sprintf("users:list:%s:%s:%s:%d:%d:%s:%s",
		opts.Role, opts.Status, opts.Search, opts.Page, opts.Limit, opts.SortBy, strings.ToLower(opts.SortOrder))
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			var snapshot cachedUsers
			if err := json.Unmarshal([]byte(cached), &snapshot); err == nil {
				return snapshot.Items, snapshot.Total, nil
			}
		}
	}

	listOpts := auth.ListUsersOptions{
		Role:      opts.Role,
		Status:    opts.Status,
		Search:    opts.Search,
		SortBy:    opts.SortBy,
		SortOrder: opts.SortOrder,
		Page:      opts.Page,
		Limit:     opts.Limit,
	}

	users, total, err := s.repo.ListUsers(ctx, listOpts)
	if err != nil {
		return nil, 0, err
	}

	if s.cache != nil {
		data, err := json.Marshal(cachedUsers{Items: users, Total: total})
		if err == nil {
			s.cache.Set(ctx, cacheKey, data, userListCacheTTL)
		}
	}

	return users, total, nil
}

// GetByID fetches a single user.
func (s *Service) GetByID(ctx context.Context, userID string) (*auth.User, error) {
	return s.repo.FindByID(ctx, userID)
}

// CreateUser registers a new user through the auth service.
func (s *Service) CreateUser(ctx context.Context, payload auth.RegisterPayload) (*auth.User, error) {
	if s.authService == nil {
		return nil, appErrors.InternalServerError
	}
	user, _, err := s.authService.Register(ctx, payload)
	return user, err
}

// UpdateUser applies editable profile fields for a target user.
func (s *Service) UpdateUser(ctx context.Context, userID string, fields map[string]interface{}) (*auth.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if v, ok := fields["fullName"].(string); ok && strings.TrimSpace(v) != "" {
		user.FullName = strings.TrimSpace(v)
	}
	if v, ok := fields["region"].(string); ok && strings.TrimSpace(v) != "" {
		user.Region = strings.TrimSpace(v)
	}
	if v, ok := fields["primaryCurrency"].(string); ok && strings.TrimSpace(v) != "" {
		user.PrimaryCurrency = strings.TrimSpace(v)
	}
	if v, ok := fields["status"].(string); ok && strings.TrimSpace(v) != "" {
		user.Status = strings.TrimSpace(v)
	}

	user.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser soft-deletes a user.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.DeleteUser(ctx, userID)
}

// GetProfile returns the current user's profile.
func (s *Service) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapUserToProfile(user), nil
}

// UpdateProfile applies editable profile fields.
func (s *Service) UpdateProfile(ctx context.Context, userID string, fields map[string]interface{}) (*Profile, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if v, ok := fields["fullName"].(string); ok && strings.TrimSpace(v) != "" {
		user.FullName = strings.TrimSpace(v)
	}
	if v, ok := fields["region"].(string); ok && strings.TrimSpace(v) != "" {
		user.Region = strings.TrimSpace(v)
	}
	if v, ok := fields["primaryCurrency"].(string); ok && strings.TrimSpace(v) != "" {
		user.PrimaryCurrency = strings.TrimSpace(v)
	}
	user.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return mapUserToProfile(user), nil
}

func mapUserToProfile(user *auth.User) *Profile {
	if user == nil {
		return nil
	}
	return &Profile{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		Region:          user.Region,
		PrimaryCurrency: user.PrimaryCurrency,
	}
}
