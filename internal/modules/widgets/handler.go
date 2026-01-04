package widgets

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes widget endpoints.
type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) List(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidWidgetData)
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
	widget, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.WidgetNotFound) {
			return response.Failure(c, appErrors.WidgetNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, widget, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload Widget
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidWidgetData)
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
	var payload Widget
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidWidgetData)
	}
	updated, err := h.service.Update(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.WidgetNotFound) {
			return response.Failure(c, appErrors.WidgetNotFound)
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
		return response.Failure(c, appErrors.InvalidWidgetData)
	}
	updated, err := h.service.Patch(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.WidgetNotFound) {
			return response.Failure(c, appErrors.WidgetNotFound)
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
		if errors.Is(err, appErrors.WidgetNotFound) {
			return response.Failure(c, appErrors.WidgetNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}
