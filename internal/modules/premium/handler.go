package premium

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
	"github.com/leora/leora-server/internal/modules/auth"
)

// Handler exposes premium/subscription endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals(auth.ContextUserIDKey).(string)
	if !ok || userID == "" {
		return response.Failure(c, appErrors.InvalidToken)
	}
	data, err := h.service.GetSubscription(c.Context(), userID)
	if err != nil {
		if errors.Is(err, appErrors.SubscriptionNotFound) {
			return response.Failure(c, appErrors.SubscriptionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, data, nil)
}

func (h *Handler) ListSubscriptions(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	data, err := h.service.ListSubscriptions(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetSubscription(c *fiber.Ctx) error {
	id := c.Params("id")
	sub, err := h.service.GetSubscription(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.SubscriptionNotFound) {
			return response.Failure(c, appErrors.SubscriptionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, sub, nil)
}

func (h *Handler) CreateSubscription(c *fiber.Ctx) error {
	var payload Subscription
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	created, err := h.service.CreateSubscription(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdateSubscription(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Subscription
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	updated, err := h.service.UpdateSubscription(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.SubscriptionNotFound) {
			return response.Failure(c, appErrors.SubscriptionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) PatchSubscription(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	updated, err := h.service.PatchSubscription(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.SubscriptionNotFound) {
			return response.Failure(c, appErrors.SubscriptionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeleteSubscription(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteSubscription(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.SubscriptionNotFound) {
			return response.Failure(c, appErrors.SubscriptionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) ListPlans(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	data, err := h.service.ListPlans(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetPlan(c *fiber.Ctx) error {
	id := c.Params("id")
	plan, err := h.service.GetPlan(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.PlanNotFound) {
			return response.Failure(c, appErrors.PlanNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, plan, nil)
}

func (h *Handler) CreatePlan(c *fiber.Ctx) error {
	var payload Plan
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	created, err := h.service.CreatePlan(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdatePlan(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Plan
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	updated, err := h.service.UpdatePlan(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.PlanNotFound) {
			return response.Failure(c, appErrors.PlanNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) PatchPlan(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidSubscriptionData)
	}
	updated, err := h.service.PatchPlan(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.PlanNotFound) {
			return response.Failure(c, appErrors.PlanNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeletePlan(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeletePlan(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.PlanNotFound) {
			return response.Failure(c, appErrors.PlanNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}
