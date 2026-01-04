package subscriptions

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    domainSub "github.com/leora/leora-server/internal/domain/subscriptions"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes subscription endpoints.
type Handler struct{}

// RegisterRoutes wires subscription endpoints.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/subscriptions")
    group.Get("/me", h.me)
    group.Get("/plans", h.plans)
    group.Post("/checkout", h.checkout)
    group.Post("/verify-receipt", h.verify)
    group.Post("/cancel", h.cancel)
    group.Post("/restore", h.restore)
    group.Get("/history", h.history)
    group.Post("/webhook/stripe", h.webhook)
    group.Post("/webhook/apple", h.webhook)
    group.Post("/webhook/google", h.webhook)
}

func (h *Handler) me(c *fiber.Ctx) error {
    subscription := domainSub.Subscription{
        ID:               uuid.NewString(),
        Tier:             "premium",
        Status:           "active",
        CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339),
        CancelAtPeriodEnd: false,
    }

    responsePayload := fiber.Map{
        "subscription": subscription,
        "features": fiber.Map{
            "ai_daily_insights": fiber.Map{"limit": -1, "used": 5},
            "ai_questions":      fiber.Map{"limit": -1, "used": 12},
            "accounts":          fiber.Map{"limit": -1, "used": 4},
        },
        "plan": fiber.Map{"name": "Premium Yillik", "interval": "yearly"},
    }

    return response.JSONSuccess(c, responsePayload, nil)
}

func (h *Handler) plans(c *fiber.Ctx) error {
    plans := []domainSub.Plan{
        {ID: "plan_monthly", Name: "Premium Oy", Tier: "premium", Interval: "monthly", Prices: map[string]float64{"USD": 4.99, "UZS": 59900}, Discount: 0, Popular: false},
        {ID: "plan_yearly", Name: "Premium Yillik", Tier: "premium", Interval: "yearly", Prices: map[string]float64{"USD": 39.99, "UZS": 479900}, Discount: 33, Popular: true},
    }
    return response.JSONSuccess(c, plans, nil)
}

func (h *Handler) checkout(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"checkoutUrl": "https://checkout.stripe.com/xyz", "sessionId": "cs_xxx"}, nil)
}

func (h *Handler) verify(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"status": "verified", "productId": "com.leora.premium.yearly"}, nil)
}

func (h *Handler) cancel(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"message": "subscription will cancel at period end"}, nil)
}

func (h *Handler) restore(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"message": "subscription restored"}, nil)
}

func (h *Handler) history(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []fiber.Map{{"id": uuid.NewString(), "status": "active", "tier": "premium"}}, nil)
}

func (h *Handler) webhook(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"status": "processed"}, nil)
}
