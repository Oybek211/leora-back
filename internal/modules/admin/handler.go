package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
	"github.com/leora/leora-server/internal/modules/auth"
)

// Handler exposes admin-only endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs a handler for admin routes.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.service.ListUsers(c.Context())
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, users, nil)
}

func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	userID := c.Params("id")
	var payload updateRoleRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	role, err := auth.ParseRole(payload.Role)
	if err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}

	user, err := h.service.UpdateUserRole(c.Context(), userID, role)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.UserNotFound)
	}
	return response.Success(c, user, nil)
}
