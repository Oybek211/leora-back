package habits

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/habits")
	group.Get("", handler.List)
	group.Post("", handler.Create)
	group.Post("/bulk-delete", handler.BulkDelete)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Patch("/:id", handler.Patch)
	group.Delete("/:id", handler.Delete)
	group.Post("/:id/complete", handler.Complete)
	group.Post("/:id/completions/toggle", handler.ToggleCompletion)
	group.Get("/:id/history", handler.GetHistory)
	group.Get("/:id/stats", handler.GetStats)

	// Finance integration
	group.Post("/:id/evaluate-finance", handler.EvaluateFinance)
	group.Post("/evaluate-all-finance", handler.EvaluateAllFinance)
}
