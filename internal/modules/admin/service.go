package admin

import (
	"context"
	"time"

	"github.com/leora/leora-server/internal/modules/auth"
)

// Service handles administrative operations around users.
type Service struct {
	repo auth.Repository
}

// NewService creates a new admin service wired to the auth repository.
func NewService(repo auth.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context) ([]*auth.User, error) {
	users, _, err := s.repo.ListUsers(ctx, auth.ListUsersOptions{
		Page:  1,
		Limit: 100,
	})
	return users, err
}

func (s *Service) UpdateUserRole(ctx context.Context, userID string, role auth.Role) (*auth.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Role = role
	user.Permissions = auth.PermissionsForRole(role)
	user.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
