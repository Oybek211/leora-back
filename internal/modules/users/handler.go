package users

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
	"github.com/leora/leora-server/internal/modules/auth"
)

// Handler exposes user profile endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs a new handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	opts, err := collectListOptions(c)
	if err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}

	users, total, err := h.service.ListUsers(c.Context(), opts)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	meta := response.Meta{
		Page:       opts.Page,
		Limit:      opts.Limit,
		Total:      total,
		TotalPages: calculateTotalPages(total, opts.Limit),
	}
	return response.Success(c, users, &meta)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	user, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.UserNotFound) {
			return response.Failure(c, appErrors.UserNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, user, nil)
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals(auth.ContextUserIDKey).(string)
	if !ok || userID == "" {
		return response.Failure(c, appErrors.InvalidToken)
	}
	user, err := h.service.GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, appErrors.UserNotFound) {
			return response.Failure(c, appErrors.UserNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, user, nil)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals(auth.ContextUserIDKey).(string)
	if !ok || userID == "" {
		return response.Failure(c, appErrors.InvalidToken)
	}
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	profile, err := h.service.UpdateProfile(c.Context(), userID, payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, profile, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload auth.RegisterPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	user, err := h.service.CreateUser(c.Context(), payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, user, nil)
}

func (h *Handler) UpdateByID(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	user, err := h.service.UpdateUser(c.Context(), userID, payload)
	if err != nil {
		if errors.Is(err, appErrors.UserNotFound) {
			return response.Failure(c, appErrors.UserNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, user, nil)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	if err := h.service.DeleteUser(c.Context(), userID); err != nil {
		if errors.Is(err, appErrors.UserNotFound) {
			return response.Failure(c, appErrors.UserNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": userID, "status": "deleted"}, nil)
}

func collectListOptions(c *fiber.Ctx) (ListOptions, error) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return ListOptions{}, errors.New("page must be a positive integer")
	}
	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit < 1 {
		return ListOptions{}, errors.New("limit must be a positive integer")
	}

	var role auth.Role
	if raw := strings.TrimSpace(c.Query("role")); raw != "" {
		parsed, err := auth.ParseRole(raw)
		if err != nil {
			return ListOptions{}, err
		}
		role = parsed
	}

	sortBy := mapSortBy(c.Query("sortBy"))
	sortOrder := strings.ToLower(c.Query("sortOrder", "desc"))
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return ListOptions{
		Role:      role,
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		Limit:     limit,
	}, nil
}

func mapSortBy(value string) string {
	switch strings.ToLower(value) {
	case "last_active_at", "lastlogin", "lastloginat":
		return "last_login_at"
	case "created_at", "createdat":
		return "created_at"
	default:
		return "created_at"
	}
}

func calculateTotalPages(total, limit int) int {
	if limit == 0 {
		return 0
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return pages
}
