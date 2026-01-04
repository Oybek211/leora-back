package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/modules/auth"
)

// RegisterRoutes wires admin endpoints.
func RegisterRoutes(router fiber.Router, handler *Handler, middleware *auth.Middleware) {
	group := router.Group("/admin")
	group.Get("/users", middleware.RequireRoleAtLeast(auth.RoleModeratorAdmin), handler.ListUsers)
	group.Patch("/users/:id/role", middleware.RequireRoleAtLeast(auth.RoleSuperAdmin), handler.UpdateUserRole)
}
