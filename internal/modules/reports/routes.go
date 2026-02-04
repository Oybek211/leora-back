package reports

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/reports")

	finance := group.Group("/finance")
	finance.Get("/summary", handler.FinanceSummary)
	finance.Get("/categories", handler.FinanceCategories)
	finance.Get("/cashflow", handler.FinanceCashflow)
	finance.Get("/debts", handler.FinanceDebts)

	planner := group.Group("/planner")
	planner.Get("/productivity", handler.PlannerProductivity)

	insights := group.Group("/insights")
	insights.Get("/daily", handler.InsightsDaily)
	insights.Get("/period", handler.InsightsPeriod)
}
