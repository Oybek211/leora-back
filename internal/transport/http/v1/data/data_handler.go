package data

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    domainData "github.com/leora/leora-server/internal/domain/data"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes data management endpoints.
type Handler struct{}

// RegisterRoutes registers data endpoints.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/data")

    group.Post("/backup", h.createBackup)
    group.Get("/backups", h.listBackups)
    group.Get("/backups/:id", h.getBackup)
    group.Post("/backups/:id/restore", h.restoreBackup)
    group.Delete("/backups/:id", h.deleteBackup)
    group.Post("/export", h.createExport)
    group.Get("/exports", h.listExports)
    group.Get("/exports/:id/download", h.downloadExport)
    group.Delete("/account", h.deleteAccount)
    group.Post("/cache/clear", h.clearCache)
}

func (h *Handler) createBackup(c *fiber.Ctx) error {
    payload := domainData.Backup{
        ID:              uuid.NewString(),
        UserID:          uuid.NewString(),
        Type:            "manual",
        Status:          "completed",
        Storage:         "cloud",
        EntitiesIncluded: []string{"tasks", "goals"},
        CreatedAt:       time.Now().UTC().Format(time.RFC3339),
    }
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) listBackups(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []domainData.Backup{h.snapshot()}, nil)
}

func (h *Handler) getBackup(c *fiber.Ctx) error {
    return response.JSONSuccess(c, h.snapshot(), nil)
}

func (h *Handler) restoreBackup(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "restored"}, nil)
}

func (h *Handler) deleteBackup(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "deleted"}, nil)
}

func (h *Handler) createExport(c *fiber.Ctx) error {
    payload := domainData.Export{
        ID:        uuid.NewString(),
        UserID:    uuid.NewString(),
        Format:    "csv",
        Scope:     "finance",
        Status:    "pending",
        CreatedAt: time.Now().UTC().Format(time.RFC3339),
    }
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) listExports(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []domainData.Export{h.snapshotExport()}, nil)
}

func (h *Handler) downloadExport(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "fileUrl": "https://storage.leora.app/export.csv"}, nil)
}

func (h *Handler) deleteAccount(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"message": "account deletion requested"}, nil)
}

func (h *Handler) clearCache(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"message": "cache cleared"}, nil)
}

func (h *Handler) snapshot() domainData.Backup {
    return domainData.Backup{
        ID:              uuid.NewString(),
        UserID:          uuid.NewString(),
        Type:            "manual",
        Status:          "completed",
        Storage:         "cloud",
        EntitiesIncluded: []string{"tasks", "habits", "accounts"},
        EntityCounts:    map[string]int{"tasks": 20, "accounts": 3},
        CreatedAt:       time.Now().UTC().Format(time.RFC3339),
    }
}

func (h *Handler) snapshotExport() domainData.Export {
    return domainData.Export{
        ID:        uuid.NewString(),
        UserID:    uuid.NewString(),
        Format:    "csv",
        Scope:     "all",
        Status:    "completed",
        FileURL:   "https://storage.leora.app/export.csv",
        CreatedAt: time.Now().UTC().Format(time.RFC3339),
    }
}
