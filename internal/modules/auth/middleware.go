package auth

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const (
	ContextUserIDKey          = "auth_user_id"
	ContextUserRoleKey        = "auth_user_role"
	ContextUserPermissionsKey = "auth_user_permissions"
)

// Middleware exposes handlers that gate routes based on JWT content.
type Middleware struct {
	service *Service
}

// NewMiddleware creates authentication middleware tied to the auth service.
func NewMiddleware(service *Service) *Middleware {
	return &Middleware{service: service}
}

// RequireAuth validates tokens and populates context with user metadata.
func (m *Middleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		trimmed := strings.TrimSpace(authHeader)
		hasHeader := trimmed != ""
		hasBearer := strings.HasPrefix(strings.ToLower(trimmed), "bearer ")
		log.Printf("[auth] RequireAuth header present=%v bearer=%v", hasHeader, hasBearer)

		token, err := ExtractBearerToken(authHeader)
		if err != nil {
			return response.Failure(c, appErrors.InvalidToken)
		}

		claims, err := m.service.ValidateAccessToken(token)
		if err != nil {
			tail := token
			if len(tail) > 8 {
				tail = "..." + tail[len(tail)-8:]
			}
			log.Printf("[auth] RequireAuth failed for %s %s (token=%s): %v", c.Method(), c.Path(), tail, err)
			if typed, ok := err.(*appErrors.Error); ok {
				return response.Failure(c, typed)
			}
			return response.Failure(c, appErrors.InvalidToken)
		}

		// Store in Fiber Locals for middleware access
		c.Locals(ContextUserIDKey, claims.UserID)
		c.Locals(ContextUserRoleKey, claims.Role)
		c.Locals(ContextUserPermissionsKey, claims.Permissions)

		// Store in Go context for repository access
		ctx := c.Context()
		ctx.SetUserValue("user_id", claims.UserID)
		ctx.SetUserValue("role", string(claims.Role))
		ctx.SetUserValue("permissions", claims.Permissions)

		return c.Next()
	}
}

// RequirePermission ensures the user has the requested permission.
func (m *Middleware) RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleVal, ok := c.Locals(ContextUserRoleKey).(Role)
		if !ok {
			return response.Failure(c, appErrors.InvalidToken)
		}
		if roleVal.Level() >= RoleAdmin.Level() {
			return c.Next()
		}

		perms, _ := c.Locals(ContextUserPermissionsKey).([]string)
		if hasPermission(perms, permission) {
			return c.Next()
		}
		return response.Failure(c, appErrors.PermissionDenied)
	}
}

// RequireRoleAtLeast enforces the minimum role hierarchy.
func (m *Middleware) RequireRoleAtLeast(role Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		current, ok := c.Locals(ContextUserRoleKey).(Role)
		if !ok {
			return response.Failure(c, appErrors.InvalidToken)
		}
		if current.Level() < role.Level() {
			return response.Failure(c, appErrors.PermissionDenied)
		}
		return c.Next()
	}
}

// ExtractBearerToken strips the bearer prefix from Authorization headers.
func ExtractBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", errors.New("authorization header missing")
	}
	lower := strings.ToLower(header)
	if strings.HasPrefix(lower, "bearer ") {
		return strings.TrimSpace(header[len("bearer "):]), nil
	}
	return "", errors.New("authorization header must be a bearer token")
}

func hasPermission(perms []string, required string) bool {
	for _, perm := range perms {
		if perm == "*" || perm == required {
			return true
		}
	}
	return false
}
