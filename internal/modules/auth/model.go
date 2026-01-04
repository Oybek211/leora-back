package auth

import (
	"fmt"
)

// Role describes the authorization level of a user.
type Role string

const (
	RoleUser           Role = "user"
	RoleModeratorAdmin Role = "moderator_admin"
	RoleAdmin          Role = "admin"
	RoleSuperAdmin     Role = "super_admin"
)

var (
	roleLevels = map[Role]int{
		RoleUser:           1,
		RoleModeratorAdmin: 2,
		RoleAdmin:          3,
		RoleSuperAdmin:     4,
	}
	rolePermissions = map[Role][]string{
		RoleUser: {
			"planner:read",
			"planner:write",
			"finance:read",
			"notifications:read",
			"widgets:read",
		},
		RoleModeratorAdmin: {
			"planner:read",
			"planner:write",
			"finance:read",
			"users:view",
			"notifications:manage",
		},
		RoleAdmin: {
			"planner:read",
			"planner:write",
			"finance:manage",
			"users:manage",
			"notifications:manage",
		},
		RoleSuperAdmin: {
			"*",
			"system:configure",
		},
	}
)

// PermissionsForRole returns a defensive copy of permissions assigned to the role.
func PermissionsForRole(role Role) []string {
	perms, ok := rolePermissions[role]
	if !ok {
		perms = rolePermissions[RoleUser]
	}
	copied := make([]string, len(perms))
	copy(copied, perms)
	return copied
}

// Level returns an ordering value that can be used to compare roles.
func (r Role) Level() int {
	if level, ok := roleLevels[r]; ok {
		return level
	}
	return 0
}

// ParseRole validates that the supplied string is a known role.
func ParseRole(value string) (Role, error) {
	role := Role(value)
	if _, ok := roleLevels[role]; !ok {
		return "", fmt.Errorf("unknown role %q", value)
	}
	return role, nil
}

// User mirrors the user entity defined in docs.
type User struct {
	ID              string   `json:"id"`
	Email           string   `json:"email"`
	FullName        string   `json:"fullName"`
	Region          string   `json:"region"`
	PrimaryCurrency string   `json:"primaryCurrency"`
	Role            Role     `json:"role"`
	Status          string   `json:"status"`
	Permissions     []string `json:"permissions"`
	PasswordHash    string   `json:"-"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
	LastLoginAt     string   `json:"lastLoginAt,omitempty"`
}

// Tokens captures JWT pair information.
type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}
