package notifications

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/notifications")
	group.Get("", handler.List)
	group.Post("", handler.Create)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Patch("/:id", handler.Patch)
	group.Delete("/:id", handler.Delete)
}
