package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/localization"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Meta describes pagination metadata.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// JSONSuccess builds a uniform success response.
func JSONSuccess(c *fiber.Ctx, data interface{}, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    meta,
		"error":   nil,
	})
}

// JSONError builds a uniform error response.
func JSONError(c *fiber.Ctx, err *appErrors.Error) error {
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
			"type":    err.Type,
			"message": message,
		},
		"meta": nil,
	})
}
