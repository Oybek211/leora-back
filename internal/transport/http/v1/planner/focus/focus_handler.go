package focus

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/common"
    "github.com/leora/leora-server/internal/domain/planner/focus"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes focus session endpoints.
type Handler struct{}

// RegisterRoutes adds focus session routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/focus-sessions")

    group.Get("", h.list)
    group.Get(":id", h.get)
    group.Post("/start", h.start)
    group.Patch(":id/pause", h.pause)
    group.Patch(":id/resume", h.resume)
    group.Post(":id/complete", h.complete)
    group.Post(":id/cancel", h.cancel)
    group.Post(":id/interrupt", h.interrupt)
    group.Get("/stats", h.stats)
}

func (h *Handler) list(c *fiber.Ctx) error {
    meta := &response.Meta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}
    return response.JSONSuccess(c, []focus.FocusSession{sampleSession()}, meta)
}

func (h *Handler) get(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleSession(), nil)
}

func (h *Handler) start(c *fiber.Ctx) error {
    type startRequest struct {
        TaskID        string `json:"taskId"`
        PlannedMinutes int   `json:"plannedMinutes"`
    }

    var req startRequest
    if err := c.BodyParser(&req); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }

    now := time.Now().UTC().Format(time.RFC3339)
    session := focus.FocusSession{
        BaseEntity: common.BaseEntity{
            ID:        uuid.NewString(),
            UserID:    uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusPending,
            CreatedAt: now,
            UpdatedAt: now,
        },
        TaskID:         req.TaskID,
        PlannedMinutes: req.PlannedMinutes,
        ActualMinutes:  0,
        Status:         focus.FocusStatusInProgress,
        StartedAt:      now,
    }

    return response.JSONSuccess(c, session, nil)
}

func (h *Handler) pause(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "paused"}, nil)
}

func (h *Handler) resume(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "in_progress"}, nil)
}

func (h *Handler) complete(c *fiber.Ctx) error {
    var payload struct {
        ActualMinutes int    `json:"actualMinutes"`
        Notes         string `json:"notes"`
    }
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "completed", "actualMinutes": payload.ActualMinutes, "notes": payload.Notes}, nil)
}

func (h *Handler) cancel(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "canceled"}, nil)
}

func (h *Handler) interrupt(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "interruptions": 1}, nil)
}

func (h *Handler) stats(c *fiber.Ctx) error {
    stats := fiber.Map{
        "today": fiber.Map{"totalMinutes": 120, "sessionsCount": 5, "completedCount": 4},
        "thisWeek": fiber.Map{"totalMinutes": 840, "sessionsCount": 35},
        "thisMonth": fiber.Map{"totalMinutes": 3200, "avgPerDay": 106.7},
    }
    return response.JSONSuccess(c, stats, nil)
}

func sampleSession() focus.FocusSession {
    now := time.Now().UTC().Format(time.RFC3339)
    return focus.FocusSession{
        BaseEntity: common.BaseEntity{
            ID:         uuid.NewString(),
            UserID:     uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt:  now,
            UpdatedAt:  now,
        },
        PlannedMinutes:   25,
        ActualMinutes:    23,
        Status:           focus.FocusStatusCompleted,
        StartedAt:        now,
        EndedAt:          now,
        InterruptionsCount: 0,
    }
}
