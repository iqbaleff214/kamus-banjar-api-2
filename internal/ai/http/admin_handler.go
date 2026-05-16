package aihttp

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	aidomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
)

type AdminHandler struct {
	svc *commands.EnrichmentService
}

func NewAdminHandler(svc *commands.EnrichmentService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) EnrichDefinition(c *fiber.Ctx) error {
	adminID, wordID, err := adminAndWord(c)
	if err != nil {
		return err
	}
	req, svcErr := h.svc.TriggerDefinitionEnrichment(c.Context(), adminID, wordID)
	if svcErr != nil {
		return mapEnrichmentError(c, svcErr)
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"success": true, "data": toAIRequestResponse(req)})
}

func (h *AdminHandler) SuggestExample(c *fiber.Ctx) error {
	adminID, wordID, err := adminAndWord(c)
	if err != nil {
		return err
	}
	req, svcErr := h.svc.TriggerExampleSuggestion(c.Context(), adminID, wordID)
	if svcErr != nil {
		return mapEnrichmentError(c, svcErr)
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"success": true, "data": toAIRequestResponse(req)})
}

func (h *AdminHandler) SuggestRelated(c *fiber.Ctx) error {
	adminID, wordID, err := adminAndWord(c)
	if err != nil {
		return err
	}
	req, svcErr := h.svc.TriggerRelatedWordSuggestion(c.Context(), adminID, wordID)
	if svcErr != nil {
		return mapEnrichmentError(c, svcErr)
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"success": true, "data": toAIRequestResponse(req)})
}

func (h *AdminHandler) QualityCheck(c *fiber.Ctx) error {
	adminID, err := adminUUID(c)
	if err != nil {
		return httperr.Send(c, fiber.StatusUnauthorized, httperr.Forbidden("invalid token"))
	}
	contribID, err := uuid.Parse(c.Params("contribution_id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, httperr.Validation("invalid contribution id"))
	}
	req, svcErr := h.svc.TriggerQualityCheck(c.Context(), adminID, contribID)
	if svcErr != nil {
		return mapEnrichmentError(c, svcErr)
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"success": true, "data": toAIRequestResponse(req)})
}

func (h *AdminHandler) ListAIRequests(c *fiber.Ctx) error {
	page, perPage := paginate(c)

	wordIDStr := c.Query("word_id")
	if wordIDStr != "" {
		wordID, err := uuid.Parse(wordIDStr)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, httperr.Validation("invalid word_id"))
		}
		reqs, total, err := h.svc.ListByWord(c.Context(), wordID, page, perPage)
		if err != nil {
			return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
		}
		return httperr.Paginated(c, toAIRequestResponses(reqs), &httperr.PaginationMeta{
			Page: page, PerPage: perPage, Total: total,
		})
	}

	reqs, total, err := h.svc.ListPendingReview(c.Context(), page, perPage)
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.Paginated(c, toAIRequestResponses(reqs), &httperr.PaginationMeta{
		Page: page, PerPage: perPage, Total: total,
	})
}

func (h *AdminHandler) ApproveAIRequest(c *fiber.Ctx) error {
	adminID, err := adminUUID(c)
	if err != nil {
		return httperr.Send(c, fiber.StatusUnauthorized, httperr.Forbidden("invalid token"))
	}
	reqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, httperr.Validation("invalid request id"))
	}
	req, svcErr := h.svc.ApproveAIRequest(c.Context(), adminID, reqID)
	if svcErr != nil {
		return mapEnrichmentError(c, svcErr)
	}
	return httperr.OK(c, toAIRequestResponse(req))
}

func (h *AdminHandler) RejectAIRequest(c *fiber.Ctx) error {
	adminID, err := adminUUID(c)
	if err != nil {
		return httperr.Send(c, fiber.StatusUnauthorized, httperr.Forbidden("invalid token"))
	}
	reqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, httperr.Validation("invalid request id"))
	}
	req, svcErr := h.svc.RejectAIRequest(c.Context(), adminID, reqID)
	if svcErr != nil {
		return mapEnrichmentError(c, svcErr)
	}
	return httperr.OK(c, toAIRequestResponse(req))
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type aiRequestResponse struct {
	ID                   string         `json:"id"`
	Type                 string         `json:"type"`
	TargetWordID         *string        `json:"target_word_id,omitempty"`
	TargetContributionID *string        `json:"target_contribution_id,omitempty"`
	RequestedBy          string         `json:"requested_by"`
	Model                string         `json:"model"`
	Status               string         `json:"status"`
	ReviewStatus         string         `json:"review_status"`
	ParsedOutput         map[string]any `json:"parsed_output,omitempty"`
	ReviewedBy           *string        `json:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time     `json:"reviewed_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
}

func toAIRequestResponse(r *aidomain.AIRequest) aiRequestResponse {
	resp := aiRequestResponse{
		ID:           r.ID.String(),
		Type:         string(r.Type),
		RequestedBy:  r.RequestedBy.String(),
		Model:        r.Model,
		Status:       string(r.Status),
		ReviewStatus: string(r.ReviewStatus),
		ParsedOutput: r.ParsedOutput,
		ReviewedAt:   r.ReviewedAt,
		CreatedAt:    r.CreatedAt,
	}
	if r.TargetWordID != nil {
		s := r.TargetWordID.String()
		resp.TargetWordID = &s
	}
	if r.TargetContributionID != nil {
		s := r.TargetContributionID.String()
		resp.TargetContributionID = &s
	}
	if r.ReviewedBy != nil {
		s := r.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	return resp
}

func toAIRequestResponses(rs []*aidomain.AIRequest) []aiRequestResponse {
	out := make([]aiRequestResponse, len(rs))
	for i, r := range rs {
		out[i] = toAIRequestResponse(r)
	}
	return out
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func adminAndWord(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	adminID, err := adminUUID(c)
	if err != nil {
		_ = httperr.Send(c, fiber.StatusUnauthorized, httperr.Forbidden("invalid token"))
		return uuid.Nil, uuid.Nil, fiber.ErrUnauthorized
	}
	wordID, err := uuid.Parse(c.Params("word_id"))
	if err != nil {
		_ = httperr.Send(c, fiber.StatusBadRequest, httperr.Validation("invalid word id"))
		return uuid.Nil, uuid.Nil, fiber.ErrBadRequest
	}
	return adminID, wordID, nil
}

func adminUUID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := auth.GetClaims(c)
	if claims == nil {
		return uuid.Nil, errors.New("no claims")
	}
	return uuid.Parse(claims.UserID)
}

func paginate(c *fiber.Ctx) (int, int) {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

func mapEnrichmentError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, commands.ErrWordNotFound),
		errors.Is(err, commands.ErrContributionNotFound),
		errors.Is(err, commands.ErrAIRequestNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound(err.Error()))
	case errors.Is(err, commands.ErrAIRequestConflict):
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict(err.Error()))
	default:
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
}
