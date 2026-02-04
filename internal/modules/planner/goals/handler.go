package goals

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes goal endpoints.
type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) List(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	goals, err := h.service.List(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	status := c.Query("status")
	goalType := c.Query("goalType")
	financeMode := c.Query("financeMode")
	filtered := make([]*Goal, 0, len(goals))
	for _, goal := range goals {
		if status != "" && goal.Status != status {
			continue
		}
		if goalType != "" && goal.GoalType != goalType {
			continue
		}
		if financeMode != "" {
			if goal.FinanceMode == nil || *goal.FinanceMode != financeMode {
				continue
			}
		}
		filtered = append(filtered, goal)
	}
	goals = filtered
	start, end := utils.SliceBounds(len(goals), page, limit)
	paged := goals[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(goals), TotalPages: utils.TotalPages(len(goals), limit)})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	goal, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, goal, nil)
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	id := c.Params("id")
	stats, err := h.service.GetStats(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, stats, nil)
}

func (h *Handler) GetTasks(c *fiber.Ctx) error {
	id := c.Params("id")
	tasks, err := h.service.GetTasks(c.Context(), id)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, tasks, nil)
}

func (h *Handler) GetHabits(c *fiber.Ctx) error {
	id := c.Params("id")
	habits, err := h.service.GetHabits(c.Context(), id)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, habits, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload Goal
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
	var payload Goal
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	updated, err := h.service.Update(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
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
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
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
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
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
	deleted, unlinkedBudgets, unlinkedDebts, err := h.service.BulkDelete(c.Context(), payload.IDs)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{
		"deleted":         deleted,
		"ids":             payload.IDs,
		"unlinkedBudgets": unlinkedBudgets,
		"unlinkedDebts":   unlinkedDebts,
	}, nil)
}

// LinkBudget links a budget to a goal (bidirectional)
func (h *Handler) LinkBudget(c *fiber.Ctx) error {
	goalID := c.Params("id")
	var payload struct {
		BudgetID string `json:"budgetId"`
	}
	if err := c.BodyParser(&payload); err != nil || payload.BudgetID == "" {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}

	goal, err := h.service.LinkBudget(c.Context(), goalID, payload.BudgetID)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, goal, nil)
}

// UnlinkBudget removes budget link from a goal
func (h *Handler) UnlinkBudget(c *fiber.Ctx) error {
	goalID := c.Params("id")
	goal, err := h.service.UnlinkBudget(c.Context(), goalID)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, goal, nil)
}

// LinkDebt links a debt to a goal (bidirectional)
func (h *Handler) LinkDebt(c *fiber.Ctx) error {
	goalID := c.Params("id")
	var payload struct {
		DebtID string `json:"debtId"`
	}
	if err := c.BodyParser(&payload); err != nil || payload.DebtID == "" {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}

	goal, err := h.service.LinkDebt(c.Context(), goalID, payload.DebtID)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, goal, nil)
}

// UnlinkDebt removes debt link from a goal
func (h *Handler) UnlinkDebt(c *fiber.Ctx) error {
	goalID := c.Params("id")
	goal, err := h.service.UnlinkDebt(c.Context(), goalID)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, goal, nil)
}

// GetFinanceProgress returns finance-based progress for a goal
func (h *Handler) GetFinanceProgress(c *fiber.Ctx) error {
	goalID := c.Params("id")
	progress, err := h.service.GetFinanceProgress(c.Context(), goalID)
	if err != nil {
		if errors.Is(err, appErrors.GoalNotFound) {
			return response.Failure(c, appErrors.GoalNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, progress, nil)
}
