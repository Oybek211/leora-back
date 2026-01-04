package ai

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    domainAI "github.com/leora/leora-server/internal/domain/ai"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes AI quota/usage endpoints.
type Handler struct{}

// RegisterRoutes registers Ai-related routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/ai")
    group.Get("/quota", h.quota)
    group.Get("/usage/history", h.usageHistory)
    group.Get("/usage/stats", h.usageStats)
}

func (h *Handler) quota(c *fiber.Ctx) error {
    quota := domainAI.AIQuota{
        ID:         uuid.NewString(),
        UserID:     uuid.NewString(),
        Channel:    "daily",
        PeriodStart: time.Now().UTC().Format(time.RFC3339),
        PeriodEnd:   time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
        Limit:      2,
        Used:       1,
        CreatedAt:  time.Now().UTC().Format(time.RFC3339),
        UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
    }
    return response.JSONSuccess(c, quota, nil)
}

func (h *Handler) usageHistory(c *fiber.Ctx) error {
    rows := []domainAI.AIUsage{
        {ID: uuid.NewString(), UserID: uuid.NewString(), Channel: "daily", RequestType: "insight", TokensUsed: 120, ResponseTime: 320, Success: true, CreatedAt: time.Now().UTC().Format(time.RFC3339)},
    }
    return response.JSONSuccess(c, rows, nil)
}

func (h *Handler) usageStats(c *fiber.Ctx) error {
    stats := fiber.Map{
        "thisMonth": fiber.Map{"totalRequests": 150, "byChannel": fiber.Map{"daily": 30, "qa": 80, "voice": 40}, "avgResponseTime": 450},
        "allTime": fiber.Map{"totalRequests": 1200},
    }
    return response.JSONSuccess(c, stats, nil)
}
