package aihttp

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
)

type Handler struct {
	svc *commands.TranslateService
}

func NewHandler(svc *commands.TranslateService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Translate(c *fiber.Ctx) error {
	var req translateRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("text is required", httperr.ErrorDetail{Field: "text", Message: "required"}))
	}

	result, err := h.svc.TranslateText(c.Context(), req.Text, req.Context)
	if err != nil {
		switch {
		case errors.Is(err, commands.ErrTextTooLong), errors.Is(err, commands.ErrEmptyText):
			return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation(err.Error()))
		case errors.Is(err, domain.ErrAIUnavailable):
			return httperr.Send(c, fiber.StatusServiceUnavailable, httperr.AIUnavailable("AI service is temporarily unavailable"))
		default:
			return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal server error"))
		}
	}

	return httperr.OK(c, toTranslationResponse(result))
}
