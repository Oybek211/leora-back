package home

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/home")
	group.Get("", handler.Summary)
	group.Get("/widgets", handler.Widgets)
	group.Get("/calendar", handler.Calendar)
}
