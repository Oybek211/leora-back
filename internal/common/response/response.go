package response

import (
	"github.com/gofiber/fiber/v2"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Meta defines pagination metadata.
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"totalPages,omitempty"`
}

// ErrorDetail contains field-specific errors.
// Success returns a success payload.
func Success(c *fiber.Ctx, data interface{}, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    meta,
		"error":   nil,
	})
}

// Failure returns a standardized error.
func Failure(c *fiber.Ctx, err *appErrors.Error) error {
	status := appErrors.StatusFromType(err.Type)
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    err.Code,
			"type":    err.Type,
			"message": err.Message,
		},
	})
}
