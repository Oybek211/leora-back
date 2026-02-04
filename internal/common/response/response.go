package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/localization"
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
	code := err.Slug
	if code == "" {
		code = err.Type
	}
	lang := localization.ResolveLanguage(c)
	message := localization.TranslateError(c.Context(), code, err.Message, lang)
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"data":    nil,
		"error": fiber.Map{
			"code":    code,
			"legacyCode": err.Code,
			"type":    normalizeErrorType(err.Type),
			"message": message,
			"details": err.Details,
		},
		"meta": nil,
	})
}

func SuccessWithStatus(c *fiber.Ctx, status int, data interface{}, meta *Meta) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    meta,
		"error":   nil,
	})
}

func normalizeErrorType(errType string) string {
	switch errType {
	case "VALIDATION":
		return "VALIDATION_ERROR"
	case "INTERNAL":
		return "INTERNAL_ERROR"
	default:
		return errType
	}
}
