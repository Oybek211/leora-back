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
