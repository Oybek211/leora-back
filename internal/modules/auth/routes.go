package auth

import "github.com/gofiber/fiber/v2"

// RegisterRoutes attaches auth routes.
func RegisterRoutes(router fiber.Router, handler *Handler, middleware *Middleware) {
	group := router.Group("/auth")
	group.Post("/register", handler.Register)
	group.Post("/login", handler.Login)
	group.Post("/refresh", handler.Refresh)
	group.Post("/forgot-password", handler.ForgotPassword)
	group.Post("/reset-password", handler.ResetPassword)
	group.Post("/google", handler.GoogleLogin)
	group.Post("/apple", handler.AppleLogin)
	group.Post("/logout", middleware.RequireAuth(), handler.Logout)
	group.Get("/me", middleware.RequireAuth(), handler.Me)
}
