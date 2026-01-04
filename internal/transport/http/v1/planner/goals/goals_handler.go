package goals

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/common"
    "github.com/leora/leora-server/internal/domain/planner/goals"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes goal endpoints.
type Handler struct{}

// RegisterRoutes wires the planner goal routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/goals")

    group.Get("", h.list)
    group.Get(":id", h.get)
    group.Post("", h.create)
    group.Patch(":id", h.update)
    group.Delete(":id", h.delete)
    group.Post(":id/check-in", h.checkIn)
    group.Post(":id/complete", h.complete)
    group.Post(":id/reactivate", h.reactivate)
}

func (h *Handler) list(c *fiber.Ctx) error {
    meta := &response.Meta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}
    return response.JSONSuccess(c, []goals.Goal{sampleGoal()}, meta)
}

func (h *Handler) get(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleGoal(), nil)
}

func (h *Handler) create(c *fiber.Ctx) error {
    var payload goals.Goal
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }
    payload.ID = uuid.NewString()
    payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
    payload.UpdatedAt = payload.CreatedAt
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) update(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }
    payload["id"] = c.Params("id")
    payload["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) delete(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "showStatus": "deleted"}, nil)
}

func (h *Handler) checkIn(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }
    return response.JSONSuccess(c, fiber.Map{"goalId": c.Params("id"), "checkIn": payload}, nil)
}

func (h *Handler) complete(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"goalId": c.Params("id"), "status": "completed"}, nil)
}

func (h *Handler) reactivate(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"goalId": c.Params("id"), "status": "active"}, nil)
}

func sampleGoal() goals.Goal {
    now := time.Now().UTC().Format(time.RFC3339)
    return goals.Goal{
        BaseEntity: common.BaseEntity{
            ID:         uuid.NewString(),
            UserID:     uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt:  now,
            UpdatedAt:  now,
        },
        Title:           "Save 1000 USD",
        Type:            goals.GoalTypeFinancial,
        Status:          "active",
        MetricType:      goals.MetricTypeAmount,
        Direction:       goals.DirectionIncrease,
        CurrentValue:    0,
        ProgressPercent: 0,
    }
}
