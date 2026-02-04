package insights

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/insights")
	group.Post("/daily", handler.Daily)
	group.Post("/period", handler.Period)
	group.Post("/qa", handler.QA)
	group.Post("/voice", handler.Voice)
}
