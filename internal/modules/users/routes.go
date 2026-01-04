package users

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/modules/auth"
)

// RegisterRoutes wires user endpoints.
func RegisterRoutes(router fiber.Router, handler *Handler, middleware *auth.Middleware) {
	group := router.Group("/users")
	group.Get("/me", handler.GetMe)
	group.Patch("/me", handler.Update)

	adminGroup := group.Group("")
	adminGroup.Use(middleware.RequirePermission("users:view"))
	adminGroup.Get("", handler.List)
	adminGroup.Get("/:id", handler.GetByID)
	adminGroup.Post("", handler.Create)
	adminGroup.Put("/:id", handler.UpdateByID)
	adminGroup.Patch("/:id", handler.UpdateByID)
	adminGroup.Delete("/:id", handler.Delete)
}
