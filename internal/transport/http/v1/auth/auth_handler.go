package auth

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/auth"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// AuthHandler contains endpoints defined in the authentication section.
type AuthHandler struct{}

// RegisterRoutes wires auth-related handlers.
func RegisterRoutes(router fiber.Router) {
    h := &AuthHandler{}
    authGroup := router.Group("/auth")

    authGroup.Post("/register", h.register)
    authGroup.Post("/login", h.login)
    authGroup.Post("/refresh", h.refresh)
    authGroup.Post("/forgot-password", h.forgotPassword)
    authGroup.Post("/reset-password", h.resetPassword)
    authGroup.Post("/logout", h.logout)
    authGroup.Get("/me", h.me)

    users := router.Group("/users")
    users.Patch("/me", h.updateMe)
}

type registerRequest struct {
    Email           string `json:"email"`
    FullName        string `json:"fullName"`
    Password        string `json:"password"`
    ConfirmPassword string `json:"confirmPassword"`
    Region          string `json:"region"`
    Currency        string `json:"currency"`
}

func (h *AuthHandler) register(c *fiber.Ctx) error {
    var req registerRequest
    if err := c.BodyParser(&req); err != nil {
        return response.JSONError(c, appErrors.InvalidUserData)
    }

    user := auth.User{
        ID:              uuid.NewString(),
        Email:           req.Email,
        FullName:        req.FullName,
        Region:          auth.Region(req.Region),
        PrimaryCurrency: req.Currency,
        IsEmailVerified: false,
        IsPhoneVerified: false,
        CreatedAt:       time.Now().UTC().Format(time.RFC3339),
        UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
    }

    return response.JSONSuccess(c, fiber.Map{
        "user":         user,
        "accessToken":  "access-token",
        "refreshToken": "refresh-token",
        "expiresIn":    3600,
    }, nil)
}

func (h *AuthHandler) login(c *fiber.Ctx) error {
    type loginRequest struct {
        EmailOrUsername string `json:"emailOrUsername"`
        Password        string `json:"password"`
        RememberMe      bool   `json:"rememberMe"`
    }

    var req loginRequest
    if err := c.BodyParser(&req); err != nil {
        return response.JSONError(c, appErrors.InvalidUserData)
    }

    user := auth.User{
        ID:              uuid.NewString(),
        Email:           req.EmailOrUsername,
        FullName:        "Leora User",
        PrimaryCurrency: "USD",
        IsEmailVerified: true,
        IsPhoneVerified: true,
        CreatedAt:       time.Now().UTC().Format(time.RFC3339),
        UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
    }

    return response.JSONSuccess(c, fiber.Map{
        "user":         user,
        "accessToken":  "access-token",
        "refreshToken": "refresh-token",
        "expiresIn":    3600,
    }, nil)
}

func (h *AuthHandler) refresh(c *fiber.Ctx) error {
    type refreshRequest struct {
        RefreshToken string `json:"refreshToken"`
    }

    var req refreshRequest
    if err := c.BodyParser(&req); err != nil {
        return response.JSONError(c, appErrors.InvalidToken)
    }

    return response.JSONSuccess(c, fiber.Map{
        "accessToken":  "access-token",
        "refreshToken": "refresh-token",
        "expiresIn":    3600,
    }, nil)
}

func (h *AuthHandler) forgotPassword(c *fiber.Ctx) error {
    type forgotRequest struct {
        Email string `json:"email"`
    }

    var req forgotRequest
    if err := c.BodyParser(&req); err != nil {
        return response.JSONError(c, appErrors.InvalidUserData)
    }

    return response.JSONSuccess(c, fiber.Map{"message": "OTP sent to email"}, nil)
}

func (h *AuthHandler) resetPassword(c *fiber.Ctx) error {
    type resetRequest struct {
        Email       string `json:"email"`
        OTP         string `json:"otp"`
        NewPassword string `json:"newPassword"`
    }

    var req resetRequest
    if err := c.BodyParser(&req); err != nil {
        return response.JSONError(c, appErrors.InvalidUserData)
    }

    return response.JSONSuccess(c, fiber.Map{"message": "password reset"}, nil)
}

func (h *AuthHandler) logout(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"message": "logged out"}, nil)
}

func (h *AuthHandler) me(c *fiber.Ctx) error {
    user := auth.User{
        ID:              uuid.NewString(),
        Email:           "me@leora.app",
        FullName:        "Leora User",
        PrimaryCurrency: "USD",
        IsEmailVerified: true,
        IsPhoneVerified: true,
        CreatedAt:       time.Now().UTC().Format(time.RFC3339),
        UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
    }

    return response.JSONSuccess(c, fiber.Map{"user": user}, nil)
}

func (h *AuthHandler) updateMe(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidUserData)
    }

    return response.JSONSuccess(c, fiber.Map{"user": payload}, nil)
}
