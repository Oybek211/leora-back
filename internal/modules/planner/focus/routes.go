package focus

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	sessions := router.Group("/focus-sessions")
	sessions.Get("", handler.List)
	sessions.Post("/start", handler.Create)
	sessions.Get("/stats", handler.GetStats)
	sessions.Get("/:id", handler.GetByID)
	sessions.Put("/:id", handler.Update)
	sessions.Patch("/:id", handler.Patch)
	sessions.Delete("/:id", handler.Delete)
	sessions.Patch("/:id/pause", handler.Pause)
	sessions.Patch("/:id/resume", handler.Resume)
	sessions.Post("/:id/complete", handler.Complete)
	sessions.Post("/:id/cancel", handler.Cancel)
	sessions.Post("/:id/interrupt", handler.Interrupt)
}
