package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes dashboard endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	date := c.Query("date")
	summary, err := h.service.Summary(c.Context(), date)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, summary, nil)
}

func (h *Handler) Widgets(c *fiber.Ctx) error {
	date := c.Query("date")
	summary, err := h.service.Summary(c.Context(), date)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, summary, nil)
}

func (h *Handler) Calendar(c *fiber.Ctx) error {
	fromDate := c.Query("from")
	toDate := c.Query("to")
	entries, err := h.service.Calendar(c.Context(), fromDate, toDate)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, entries, nil)
}
