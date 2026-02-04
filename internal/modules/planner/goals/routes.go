package goals

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/goals")
	group.Get("", handler.List)
	group.Post("", handler.Create)
	group.Post("/bulk-delete", handler.BulkDelete)
	group.Get("/:id", handler.GetByID)
	group.Get("/:id/stats", handler.GetStats)
	group.Get("/:id/tasks", handler.GetTasks)
	group.Get("/:id/habits", handler.GetHabits)
	group.Put("/:id", handler.Update)
	group.Patch("/:id", handler.Patch)
	group.Delete("/:id", handler.Delete)

	// Finance integration endpoints
	group.Post("/:id/link-budget", handler.LinkBudget)
	group.Delete("/:id/unlink-budget", handler.UnlinkBudget)
	group.Post("/:id/link-debt", handler.LinkDebt)
	group.Delete("/:id/unlink-debt", handler.UnlinkDebt)
	group.Get("/:id/finance-progress", handler.GetFinanceProgress)
}
