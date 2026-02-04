package insights

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
)

// Handler exposes insights endpoints.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Daily(c *fiber.Ctx) error {
	return response.Success(c, defaultAiResponse("daily"), nil)
}

func (h *Handler) Period(c *fiber.Ctx) error {
	return response.Success(c, defaultAiResponse("period"), nil)
}

func (h *Handler) QA(c *fiber.Ctx) error {
	return response.Success(c, defaultAiResponse("qa"), nil)
}

func (h *Handler) Voice(c *fiber.Ctx) error {
	return response.Success(c, defaultAiResponse("voice"), nil)
}

func defaultAiResponse(mode string) map[string]interface{} {
	return map[string]interface{}{
		"narration": "Insights generation is not configured yet.",
		"actions":   []interface{}{},
		"cards":     []interface{}{},
		"metadata": map[string]interface{}{
			"mode":        mode,
			"model":       "stub",
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
		},
	}
}
