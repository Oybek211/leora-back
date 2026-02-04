package reports

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes report endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) FinanceSummary(c *fiber.Ctx) error {
	fromDate := c.Query("from")
	toDate := c.Query("to")
	data, err := h.service.FinanceSummary(c.Context(), fromDate, toDate)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) FinanceCategories(c *fiber.Ctx) error {
	fromDate := c.Query("from")
	toDate := c.Query("to")
	data, err := h.service.FinanceCategories(c.Context(), fromDate, toDate)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) FinanceCashflow(c *fiber.Ctx) error {
	fromDate := c.Query("from")
	toDate := c.Query("to")
	granularity := c.Query("granularity")
	data, err := h.service.Cashflow(c.Context(), fromDate, toDate, granularity)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) FinanceDebts(c *fiber.Ctx) error {
	data, err := h.service.DebtReport(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) PlannerProductivity(c *fiber.Ctx) error {
	fromDate := c.Query("from")
	toDate := c.Query("to")
	data, err := h.service.Productivity(c.Context(), fromDate, toDate)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) InsightsDaily(c *fiber.Ctx) error {
	data, err := h.service.InsightsContext(c.Context(), c.Query("from"), c.Query("to"))
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) InsightsPeriod(c *fiber.Ctx) error {
	data, err := h.service.InsightsContext(c.Context(), c.Query("from"), c.Query("to"))
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}
