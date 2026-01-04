package insights

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/leora/leora-server/internal/domain/insights"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes insight endpoints.
type Handler struct{}

// RegisterRoutes registers insight endpoints.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/insights")
    group.Get("", h.summary)
}

func (h *Handler) summary(c *fiber.Ctx) error {
    insight := insights.InsightSummary{
        TasksCompleted: 120,
        HabitsStreak:   7,
        FocusMinutes:   560,
        SavingsGrowth:  230000,
    }

    trends := []insights.InsightDetail{
        {Period: "week", Value: 3.2},
        {Period: "month", Value: 12.4},
    }

    return response.JSONSuccess(c, fiber.Map{"summary": insight, "trends": trends, "generatedAt": time.Now().UTC().Format(time.RFC3339)}, nil)
}
