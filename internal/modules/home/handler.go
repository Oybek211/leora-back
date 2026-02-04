package home

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes home endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	date := c.Query("date") // accepts ?date=2026-01-17
	data, err := h.service.Summary(c.Context(), date)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) Widgets(c *fiber.Ctx) error {
	date := c.Query("date") // accepts ?date=2026-01-17
	data, err := h.service.Summary(c.Context(), date)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) Calendar(c *fiber.Ctx) error {
	return response.Success(c, map[string]interface{}{}, nil)
}
