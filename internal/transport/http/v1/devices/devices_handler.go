package devices

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/devicesessions"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes device/session endpoints.
type Handler struct{}

// RegisterRoutes registers device and session routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    devices := router.Group("/devices")
    devices.Get("", h.listDevices)
    devices.Delete(":id", h.deleteDevice)
    devices.Post(":id/trust", h.trustDevice)

    sessions := router.Group("/sessions")
    sessions.Get("", h.listSessions)
    sessions.Delete(":id", h.deleteSession)
    sessions.Delete("/all", h.deleteAll)
}

func (h *Handler) listDevices(c *fiber.Ctx) error {
    now := time.Now().UTC().Format(time.RFC3339)
    data := []devicesessions.UserDevice{{
        ID: uuid.NewString(), UserID: uuid.NewString(), DeviceID: "device-1", DeviceType: devicesessions.DeviceTypeIOS, LastActiveAt: now, IsTrusted: true, CreatedAt: now,
    }}
    return response.JSONSuccess(c, data, nil)
}

func (h *Handler) deleteDevice(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "removed"}, nil)
}

func (h *Handler) trustDevice(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "trusted"}, nil)
}

func (h *Handler) listSessions(c *fiber.Ctx) error {
    now := time.Now().UTC().Format(time.RFC3339)
    data := []devicesessions.UserSession{{
        ID: uuid.NewString(), UserID: uuid.NewString(), DeviceID: "device-1", TokenHash: "hash", IsActive: true, ExpiresAt: now, LastUsedAt: now, CreatedAt: now,
    }}
    return response.JSONSuccess(c, data, nil)
}

func (h *Handler) deleteSession(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "revoked"}, nil)
}

func (h *Handler) deleteAll(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"status": "all_sessions_revoked"}, nil)
}
