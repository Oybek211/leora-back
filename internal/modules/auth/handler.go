package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes auth endpoints via Fiber.
type Handler struct {
	service *Service
}

// NewHandler creates a handler for auth routes.
func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

type registerRequest struct {
	Email           string `json:"email"`
	FullName        string `json:"fullName"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Region          string `json:"region"`
	Currency        string `json:"currency"`
}

type loginRequest struct {
	EmailOrUsername string `json:"emailOrUsername"`
	Password        string `json:"password"`
	RememberMe      bool   `json:"rememberMe"`
}

type refreshRequest struct {
	RefreshToken      string `json:"refreshToken"`
	RefreshTokenSnake string `json:"refresh_token"`
}

type resetRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var payload registerRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}

	user, tokens, err := h.service.Register(c.Context(), RegisterPayload{
		Email:           payload.Email,
		FullName:        payload.FullName,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
		Region:          payload.Region,
		Currency:        payload.Currency,
	})
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	return response.Success(c, fiber.Map{"user": user, "accessToken": tokens.AccessToken, "refreshToken": tokens.RefreshToken, "expiresIn": tokens.ExpiresIn}, nil)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var payload loginRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}

	user, tokens, err := h.service.Login(c.Context(), payload.EmailOrUsername, payload.Password)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InvalidCredentials)
	}

	return response.Success(c, fiber.Map{"user": user, "accessToken": tokens.AccessToken, "refreshToken": tokens.RefreshToken, "expiresIn": tokens.ExpiresIn}, nil)
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var payload refreshRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidRefreshToken)
	}

	refreshToken := strings.TrimSpace(payload.RefreshToken)
	if refreshToken == "" {
		refreshToken = strings.TrimSpace(payload.RefreshTokenSnake)
	}
	if refreshToken == "" {
		return response.Failure(c, appErrors.InvalidRefreshToken)
	}

	tokens, err := h.service.Refresh(c.Context(), refreshToken)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InvalidRefreshToken)
	}

	return response.Success(c, fiber.Map{"accessToken": tokens.AccessToken, "refreshToken": tokens.RefreshToken, "expiresIn": tokens.ExpiresIn}, nil)
}

func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var payload struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	if err := h.service.ForgotPassword(c.Context(), payload.Email); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.UserNotFound)
	}
	return response.Success(c, fiber.Map{"message": "OTP sent to email"}, nil)
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var payload resetRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidUserData)
	}
	if err := h.service.ResetPassword(c.Context(), payload.Email, payload.OTP, payload.NewPassword); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.UserNotFound)
	}
	return response.Success(c, fiber.Map{"message": "password reset"}, nil)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	token, err := ExtractBearerToken(c.Get("Authorization"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidToken)
	}
	userID, ok := c.Locals(ContextUserIDKey).(string)
	if !ok || userID == "" {
		return response.Failure(c, appErrors.InvalidToken)
	}
	if err := h.service.Logout(c.Context(), token, userID); err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"message": "logged out"}, nil)
}

type googleLoginRequest struct {
	IDToken  string `json:"idToken"`
	Region   string `json:"region"`
	Currency string `json:"currency"`
}

func (h *Handler) GoogleLogin(c *fiber.Ctx) error {
	var payload googleLoginRequest
	if err := c.BodyParser(&payload); err != nil || strings.TrimSpace(payload.IDToken) == "" {
		return response.Failure(c, appErrors.InvalidGoogleToken)
	}

	user, tokens, err := h.service.GoogleLogin(c.Context(), GoogleLoginPayload{
		IDToken:  payload.IDToken,
		Region:   payload.Region,
		Currency: payload.Currency,
	})
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	return response.Success(c, fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	}, nil)
}

type appleLoginRequest struct {
	IdentityToken string `json:"identityToken"`
	Email         string `json:"email"`
	FullName      string `json:"fullName"`
	Region        string `json:"region"`
	Currency      string `json:"currency"`
}

func (h *Handler) AppleLogin(c *fiber.Ctx) error {
	var payload appleLoginRequest
	if err := c.BodyParser(&payload); err != nil || strings.TrimSpace(payload.IdentityToken) == "" {
		return response.Failure(c, appErrors.InvalidAppleToken)
	}

	user, tokens, err := h.service.AppleLogin(c.Context(), AppleLoginPayload{
		IdentityToken: payload.IdentityToken,
		Email:         payload.Email,
		FullName:      payload.FullName,
		Region:        payload.Region,
		Currency:      payload.Currency,
	})
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	return response.Success(c, fiber.Map{
		"user":         user,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	}, nil)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals(ContextUserIDKey).(string)
	if !ok || userID == "" {
		return response.Failure(c, appErrors.InvalidToken)
	}
	user, err := h.service.Profile(c.Context(), userID)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.UserNotFound)
	}
	return response.Success(c, fiber.Map{"user": user}, nil)
}
