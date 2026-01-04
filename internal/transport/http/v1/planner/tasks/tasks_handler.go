package tasks

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/common"
    "github.com/leora/leora-server/internal/domain/planner/tasks"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes task-related endpoints.
type Handler struct{}

// RegisterRoutes wires task routes under the planner module.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/tasks")

    group.Get("", h.list)
    group.Get(":id", h.get)
    group.Post("", h.create)
    group.Patch(":id", h.update)
    group.Delete(":id", h.delete)
    group.Post(":id/complete", h.complete)
    group.Post(":id/reopen", h.reopen)
    group.Patch(":id/checklist/:itemId", h.updateChecklist)
}

func (h *Handler) list(c *fiber.Ctx) error {
    t := sampleTask()
    meta := &response.Meta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}
    return response.JSONSuccess(c, []tasks.Task{t}, meta)
}

func (h *Handler) get(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleTask(), nil)
}

func (h *Handler) create(c *fiber.Ctx) error {
    var payload tasks.Task
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }

    payload.ID = uuid.NewString()
    payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
    payload.UpdatedAt = payload.CreatedAt
    payload.FocusTotalMinutes = 0

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

func (h *Handler) complete(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "completed"}, nil)
}

func (h *Handler) reopen(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "planned"}, nil)
}

func (h *Handler) updateChecklist(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidPlannerData)
    }
    return response.JSONSuccess(c, fiber.Map{"taskId": c.Params("id"), "itemId": c.Params("itemId"), "item": payload}, nil)
}

func sampleTask() tasks.Task {
    now := time.Now().UTC().Format(time.RFC3339)
    return tasks.Task{
        BaseEntity: common.BaseEntity{
            ID:          uuid.NewString(),
            UserID:      uuid.NewString(),
            ShowStatus:  common.ShowStatusActive,
            SyncStatus:  common.SyncStatusSynced,
            CreatedAt:   now,
            UpdatedAt:   now,
        },
        Title:              "Read strategy doc",
        Status:             tasks.TaskStatusInbox,
        Priority:           tasks.TaskPriorityMedium,
        FocusTotalMinutes:  0,
    }
}
