package tasks

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes task endpoints.
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

	// Parse filter and sort options from query params
	opts := ListOptions{
		Status:      c.Query("status"),
		ShowStatus:  c.Query("showStatus"),
		Priority:    c.Query("priority"),
		GoalID:      c.Query("goalId"),
		HabitID:     c.Query("habitId"),
		DueDate:     c.Query("dueDate"),
		DueDateFrom: c.Query("dueDateFrom"),
		DueDateTo:   c.Query("dueDateTo"),
		Search:      c.Query("search"),
		SortBy:      c.Query("sortBy"),
		SortOrder:   c.Query("sortOrder"),
	}

	list, err := h.service.List(c.Context(), opts)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(list), page, limit)
	paged := list[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(list), TotalPages: utils.TotalPages(len(list), limit)})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload Task
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

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, task, nil)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Task
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	updated, err := h.service.Update(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
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
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
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
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) Complete(c *fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.service.Complete(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, task, nil)
}

func (h *Handler) Reopen(c *fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.service.Reopen(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, task, nil)
}

func (h *Handler) UpdateChecklistItem(c *fiber.Ctx) error {
	taskID := c.Params("id")
	itemID := c.Params("itemId")

	var payload struct {
		Completed bool `json:"completed"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}

	if err := h.service.UpdateChecklistItem(c.Context(), taskID, itemID, payload.Completed); err != nil {
		if errors.Is(err, appErrors.TaskNotFound) {
			return response.Failure(c, appErrors.TaskNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	return response.Success(c, fiber.Map{"taskId": taskID, "itemId": itemID, "completed": payload.Completed}, nil)
}
