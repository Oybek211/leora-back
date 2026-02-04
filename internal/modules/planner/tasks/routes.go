package tasks

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/tasks")
	group.Get("", handler.List)
	group.Post("", handler.Create)
	group.Post("/bulk-delete", handler.BulkDelete)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Patch("/:id", handler.Patch)
	group.Delete("/:id", handler.Delete)
	group.Post("/:id/complete", handler.Complete)
	group.Post("/:id/reopen", handler.Reopen)
	group.Patch("/:id/checklist/:itemId", handler.UpdateChecklistItem)

	// Finance integration
	group.Post("/check-finance-trigger", handler.CheckFinanceTrigger)
}
