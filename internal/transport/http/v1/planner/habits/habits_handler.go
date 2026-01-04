package habits

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/common"
    "github.com/leora/leora-server/internal/domain/planner/habits"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes habit endpoints.
type Handler struct{}

// RegisterRoutes adds habit routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/habits")

    group.Get("", h.list)
    group.Get(":id", h.get)
    group.Post("", h.create)
    group.Patch(":id", h.update)
    group.Delete(":id", h.delete)
    group.Post(":id/complete", h.complete)
    group.Get(":id/history", h.history)
    group.Get(":id/stats", h.stats)
}

func (h *Handler) list(c *fiber.Ctx) error {
    meta := &response.Meta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}
    return response.JSONSuccess(c, []habits.Habit{sampleHabit()}, meta)
}

func (h *Handler) get(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleHabit(), nil)
}

func (h *Handler) create(c *fiber.Ctx) error {
    var payload habits.Habit
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

func (h *Handler) complete(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"habitId": c.Params("id"), "status": "done"}, nil)
}

func (h *Handler) history(c *fiber.Ctx) error {
    rows := []habits.HabitCompletionEntry{
        {
            ID:        uuid.NewString(),
            HabitID:   c.Params("id"),
            DateKey:   time.Now().UTC().Format("2006-01-02"),
            Status:    "done",
            Value:     1,
            CreatedAt: time.Now().UTC().Format(time.RFC3339),
        },
    }
    return response.JSONSuccess(c, rows, nil)
}

func (h *Handler) stats(c *fiber.Ctx) error {
    stats := fiber.Map{
        "streakCurrent": 10,
        "streakBest":    15,
        "completionRate": 92.5,
    }
    return response.JSONSuccess(c, stats, nil)
}

func sampleHabit() habits.Habit {
    now := time.Now().UTC().Format(time.RFC3339)
    return habits.Habit{
        BaseEntity: common.BaseEntity{
            ID:         uuid.NewString(),
            UserID:     uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt:  now,
            UpdatedAt:  now,
        },
        Title:             "Morning run",
        Type:              habits.HabitTypeHealth,
        Status:            "active",
        Frequency:         habits.HabitFrequencyDaily,
        CompletionMode:    habits.CompletionModeNumeric,
        CountingType:      habits.CountingTypeCreate,
        Difficulty:        habits.DifficultyMedium,
        Priority:          "medium",
        ReminderEnabled:   true,
        StreakCurrent:     3,
        StreakBest:        8,
        CompletionRate30d: 82.1,
    }
}
