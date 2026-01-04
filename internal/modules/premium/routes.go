package premium

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/subscriptions")
	group.Get("", handler.ListSubscriptions)
	group.Post("", handler.CreateSubscription)
	group.Get("/:id", handler.GetSubscription)
	group.Put("/:id", handler.UpdateSubscription)
	group.Patch("/:id", handler.PatchSubscription)
	group.Delete("/:id", handler.DeleteSubscription)
	group.Get("/me", handler.Me)

	plans := group.Group("/plans")
	plans.Get("", handler.ListPlans)
	plans.Post("", handler.CreatePlan)
	plans.Get("/:id", handler.GetPlan)
	plans.Put("/:id", handler.UpdatePlan)
	plans.Patch("/:id", handler.PatchPlan)
	plans.Delete("/:id", handler.DeletePlan)
}
