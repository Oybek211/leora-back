package settings

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/settings"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes settings endpoints.
type Handler struct{}

// RegisterRoutes registers settings-related routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/settings")

    group.Get("", h.get)
    group.Patch("", h.update)
    group.Patch("/notifications", h.updateNotifications)
    group.Patch("/security", h.updateSecurity)
    group.Patch("/ai", h.updateAI)
    group.Patch("/focus", h.updateFocus)
    group.Post("/reset", h.reset)
}

func (h *Handler) get(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleSettings(), nil)
}

func (h *Handler) update(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidUserData)
    }
    payload["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) updateNotifications(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"notifications": "updated"}, nil)
}

func (h *Handler) updateSecurity(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"security": "updated"}, nil)
}

func (h *Handler) updateAI(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"ai": "updated"}, nil)
}

func (h *Handler) updateFocus(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"focus": "updated"}, nil)
}

func (h *Handler) reset(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"message": "settings reset"}, nil)
}

func sampleSettings() settings.UserSettings {
    now := time.Now().UTC().Format(time.RFC3339)
    return settings.UserSettings{
        ID:       uuid.NewString(),
        UserID:   uuid.NewString(),
        Theme:    "light",
        Language: "uz",
        Notifications: map[string]interface{}{
            "enabled": true,
        },
        Security: map[string]interface{}{
            "lockEnabled": true,
        },
        AI: map[string]interface{}{
            "helpLevel": "medium",
        },
        Focus: map[string]interface{}{
            "defaultDuration": 25,
        },
        Privacy: map[string]interface{}{
            "anonymousAnalytics": true,
        },
        CreatedAt: now,
        UpdatedAt: now,
    }
}
