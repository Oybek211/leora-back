package sync

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/sync"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler manages sync endpoints.
type Handler struct{}

// RegisterRoutes registers sync endpoints.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/sync")
    group.Post("/request", h.request)
    group.Get("/status", h.status)
}

func (h *Handler) request(c *fiber.Ctx) error {
    event := sync.SyncEvent{
        ID:        uuid.NewString(),
        UserID:    uuid.NewString(),
        Status:    sync.SyncStatusRunning,
        StartedAt: time.Now().UTC().Format(time.RFC3339),
    }
    return response.JSONSuccess(c, event, nil)
}

func (h *Handler) status(c *fiber.Ctx) error {
    event := sync.SyncEvent{
        ID:          uuid.NewString(),
        UserID:      uuid.NewString(),
        Status:      sync.SyncStatusSuccess,
        StartedAt:   time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
        CompletedAt: time.Now().UTC().Format(time.RFC3339),
    }
    return response.JSONSuccess(c, event, nil)
}
