package integrations

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    domainIntegrations "github.com/leora/leora-server/internal/domain/integrations"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes integration endpoints.
type Handler struct{}

// RegisterRoutes registers integration routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/integrations")

    group.Get("", h.list)
    group.Get(":provider/connect", h.connect)
    group.Post(":provider/callback", h.callback)
    group.Delete(":id", h.disconnect)
    group.Post(":id/sync", h.sync)
    group.Get(":id/logs", h.logs)
}

func (h *Handler) list(c *fiber.Ctx) error {
    integrations := []domainIntegrations.Integration{
        {ID: uuid.NewString(), UserID: uuid.NewString(), Provider: "google_calendar", Category: "calendar", Status: domainIntegrations.IntegrationStatusConnected},
    }
    return response.JSONSuccess(c, integrations, nil)
}

func (h *Handler) connect(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"authUrl": "https://accounts.google.com/o/oauth2/v2/auth", "state": uuid.NewString()}, nil)
}

func (h *Handler) callback(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"status": "connected"}, nil)
}

func (h *Handler) disconnect(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "disconnected"}, nil)
}

func (h *Handler) sync(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "sync_started"}, nil)
}

func (h *Handler) logs(c *fiber.Ctx) error {
    logs := []domainIntegrations.SyncLog{
        {ID: uuid.NewString(), IntegrationID: c.Params("id"), Direction: "pull", Status: "success", ItemsSynced: 5, StartedAt: time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339), CompletedAt: time.Now().UTC().Format(time.RFC3339)},
    }
    return response.JSONSuccess(c, logs, nil)
}
