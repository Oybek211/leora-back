package search

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes search endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Query(c *fiber.Ctx) error {
	term := c.Query("q", "")
	data, err := h.service.Query(c.Context(), term)
	if err != nil {
		return response.Failure(c, appErrors.SearchError)
	}
	return response.Success(c, data, nil)
}
