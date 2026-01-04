package achievements

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    domainAchievements "github.com/leora/leora-server/internal/domain/achievements"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler exposes achievement endpoints.
type Handler struct{}

// RegisterRoutes registers achievement routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    group := router.Group("/achievements")

    group.Get("", h.list)
    group.Get(":id", h.get)
    group.Get("/categories", h.categories)
    group.Post(":id/claim", h.claim)

    users := router.Group("/users")
    users.Get("/me/level", h.level)
    users.Get("/me/achievements", h.userAchievements)
}

func (h *Handler) list(c *fiber.Ctx) error {
    data := fiber.Map{
        "achievements": []domainAchievements.Achievement{sampleAchievement()},
        "userProgress": fiber.Map{"totalUnlocked": 5, "totalAchievements": 50},
    }
    return response.JSONSuccess(c, data, nil)
}

func (h *Handler) get(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleAchievement(), nil)
}

func (h *Handler) categories(c *fiber.Ctx) error {
    categories := []string{"finance", "tasks", "habits", "focus"}
    return response.JSONSuccess(c, categories, nil)
}

func (h *Handler) claim(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "claimed": true}, nil)
}

func (h *Handler) level(c *fiber.Ctx) error {
    level := domainAchievements.UserLevel{
        ID:        uuid.NewString(),
        UserID:    uuid.NewString(),
        Level:     12,
        CurrentXP: 350,
        TotalXP:   5850,
        Title:     "Achiever",
        CreatedAt: time.Now().UTC().Format(time.RFC3339),
        UpdatedAt: time.Now().UTC().Format(time.RFC3339),
    }
    return response.JSONSuccess(c, level, nil)
}

func (h *Handler) userAchievements(c *fiber.Ctx) error {
    entries := []domainAchievements.UserAchievement{
        {ID: uuid.NewString(), UserID: uuid.NewString(), AchievementID: uuid.NewString(), Progress: 100},
    }
    return response.JSONSuccess(c, entries, nil)
}

func sampleAchievement() domainAchievements.Achievement {
    return domainAchievements.Achievement{
        ID:          uuid.NewString(),
        Key:         "tasks_completed_100",
        Name:        "Century Finisher",
        Description: "Complete 100 tasks",
        Category:    "tasks",
        Icon:        "task",
        Color:       "#FFD700",
        XPReward:    100,
        Requirement: map[string]interface{}{"type": "tasks_completed", "count": 100},
        IsSecret:    false,
        Tier:        "gold",
        Order:       1,
    }
}
