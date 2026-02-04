package dashboard

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/dashboard")
	group.Get("/summary", handler.Summary)
	group.Get("/widgets", handler.Widgets)
	group.Get("/calendar", handler.Calendar)
}
