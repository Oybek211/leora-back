package habits

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes habit endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	data, err := h.service.List(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	status := c.Query("status")
	habitType := c.Query("habitType")
	goalID := c.Query("goalId")
	filtered := make([]*Habit, 0, len(data))
	for _, habit := range data {
		if status != "" && habit.Status != status {
			continue
		}
		if habitType != "" && habit.HabitType != habitType {
			continue
		}
		if goalID != "" {
			if habit.GoalID == nil || *habit.GoalID != goalID {
				continue
			}
		}
		filtered = append(filtered, habit)
	}
	data = filtered
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	habit, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, habit, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload Habit
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	created, err := h.service.Create(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Habit
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	updated, err := h.service.Update(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) Patch(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	updated, err := h.service.Patch(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) BulkDelete(c *fiber.Ctx) error {
	var payload struct {
		IDs []string `json:"ids"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	if len(payload.IDs) == 0 {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	deleted, err := h.service.BulkDelete(c.Context(), payload.IDs)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"deleted": deleted, "ids": payload.IDs}, nil)
}

func (h *Handler) ToggleCompletion(c *fiber.Ctx) error {
	habitID := c.Params("id")
	var payload struct {
		Date string `json:"date"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	// Default to today if no date provided
	if payload.Date == "" {
		payload.Date = utils.NowUTC()[:10] // YYYY-MM-DD format
	}
	completion, err := h.service.ToggleCompletion(c.Context(), habitID, payload.Date)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, completion, nil)
}

func (h *Handler) Complete(c *fiber.Ctx) error {
	habitID := c.Params("id")
	var payload HabitCompletion
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	payload.HabitID = habitID
	completion, err := h.service.CreateCompletion(c.Context(), &payload)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, completion, nil)
}

func (h *Handler) GetHistory(c *fiber.Ctx) error {
	habitID := c.Params("id")
	completions, err := h.service.GetCompletions(c.Context(), habitID)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, completions, nil)
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	habitID := c.Params("id")
	stats, err := h.service.GetStats(c.Context(), habitID)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, stats, nil)
}

// EvaluateFinance evaluates a habit against finance transactions
func (h *Handler) EvaluateFinance(c *fiber.Ctx) error {
	habitID := c.Params("id")

	var payload struct {
		DateKey string `json:"dateKey"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}

	// Default to today if no date provided
	if payload.DateKey == "" {
		payload.DateKey = utils.NowUTC()[:10] // YYYY-MM-DD format
	}

	result, err := h.service.EvaluateFinance(c.Context(), habitID, payload.DateKey)
	if err != nil {
		if errors.Is(err, appErrors.HabitNotFound) {
			return response.Failure(c, appErrors.HabitNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	return response.Success(c, result, nil)
}

// EvaluateAllFinance evaluates finance rules for all habits for the given date.
func (h *Handler) EvaluateAllFinance(c *fiber.Ctx) error {
	var payload struct {
		DateKey string `json:"dateKey"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	if payload.DateKey == "" {
		payload.DateKey = utils.NowUTC()[:10]
	}
	results, err := h.service.EvaluateAllFinance(c.Context(), payload.DateKey)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, results, nil)
}
