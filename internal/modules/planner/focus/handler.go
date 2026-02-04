package focus

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes focus endpoints.
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
	list, err := h.service.List(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	status := c.Query("status")
	taskID := c.Query("taskId")
	goalID := c.Query("goalId")
	filtered := make([]*Session, 0, len(list))
	for _, session := range list {
		if status != "" && session.Status != status {
			continue
		}
		if taskID != "" {
			if session.TaskID == nil || *session.TaskID != taskID {
				continue
			}
		}
		if goalID != "" {
			if session.GoalID == nil || *session.GoalID != goalID {
				continue
			}
		}
		filtered = append(filtered, session)
	}
	list = filtered
	start, end := utils.SliceBounds(len(list), page, limit)
	paged := list[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(list), TotalPages: utils.TotalPages(len(list), limit)})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload Session
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	session, err := h.service.Create(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Session
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	updated, err := h.service.Update(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
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
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
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
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	stats, err := h.service.GetStats(c.Context())
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, stats, nil)
}

func (h *Handler) Pause(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.service.Pause(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}

func (h *Handler) Resume(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.service.Resume(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}

func (h *Handler) Complete(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload struct {
		ActualMinutes int     `json:"actualMinutes"`
		Notes         *string `json:"notes"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidPlannerData)
	}
	session, err := h.service.Complete(c.Context(), id, payload.ActualMinutes, payload.Notes)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}

func (h *Handler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.service.Cancel(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}

func (h *Handler) Interrupt(c *fiber.Ctx) error {
	id := c.Params("id")
	session, err := h.service.Interrupt(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.FocusSessionNotFound) {
			return response.Failure(c, appErrors.FocusSessionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, session, nil)
}
