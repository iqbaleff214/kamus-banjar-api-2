package moderationhttp

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	commands "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/application/commands"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
	pg "github.com/iqbaleff214/kamus-banjar-api-2/pkg/pagination"
)

type Handler struct {
	svc *commands.ModerationService
}

func NewHandler(svc *commands.ModerationService) *Handler {
	return &Handler{svc: svc}
}

// ─── Moderation queue ─────────────────────────────────────────────────────────

func (h *Handler) GetModerationQueue(c *fiber.Ctx) error {
	p, errResp := pg.Parse(c)
	if errResp != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp)
	}

	contribs, total, err := h.svc.GetModerationQueue(c.Context(), nil, p.Page, p.PerPage)
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}

	out := make([]contributionResponse, len(contribs))
	for i, contrib := range contribs {
		out[i] = toContributionResponse(contrib)
	}
	return httperr.Paginated(c, out, &httperr.PaginationMeta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: (total + p.PerPage - 1) / p.PerPage,
	})
}

func (h *Handler) GetFlaggedComments(c *fiber.Ctx) error {
	p, errResp := pg.Parse(c)
	if errResp != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp)
	}

	comments, total, err := h.svc.GetFlaggedComments(c.Context(), p.Page, p.PerPage)
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}

	out := make([]commentResponse, len(comments))
	for i, cm := range comments {
		out[i] = toCommentResponse(cm)
	}
	return httperr.Paginated(c, out, &httperr.PaginationMeta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: (total + p.PerPage - 1) / p.PerPage,
	})
}

func (h *Handler) GetModerationStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetStats(c.Context())
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.OK(c, toStatsResponse(stats))
}

// ─── User management ──────────────────────────────────────────────────────────

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	p, errResp := pg.Parse(c)
	if errResp != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp)
	}

	role := c.Query("role")
	query := c.Query("q")
	var isActive *bool
	if raw := c.Query("is_active"); raw == "true" {
		v := true
		isActive = &v
	} else if raw == "false" {
		v := false
		isActive = &v
	}

	users, total, err := h.svc.ListUsers(c.Context(), role, query, isActive, p.Page, p.PerPage)
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}

	out := make([]userResponse, len(users))
	for i, u := range users {
		out[i] = toUserResponse(u)
	}
	return httperr.Paginated(c, out, &httperr.PaginationMeta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: (total + p.PerPage - 1) / p.PerPage,
	})
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, httperr.NotFound("user not found"))
	}

	u, err := h.svc.GetUser(c.Context(), id)
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("user not found"))
	}
	return httperr.OK(c, toUserResponse(u))
}

func (h *Handler) BanUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	adminID, _ := uuid.Parse(claims.UserID)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("user not found"))
	}

	var req banUserRequest
	_ = c.BodyParser(&req)

	u, err := h.svc.BanUser(c.Context(), adminID, targetID, req.Reason)
	if err != nil {
		return mapModerationError(c, err)
	}
	return httperr.OK(c, toUserResponse(u))
}

func (h *Handler) UnbanUser(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	adminID, _ := uuid.Parse(claims.UserID)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("user not found"))
	}

	u, err := h.svc.UnbanUser(c.Context(), adminID, targetID)
	if err != nil {
		return mapModerationError(c, err)
	}
	return httperr.OK(c, toUserResponse(u))
}

func (h *Handler) ChangeUserRole(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	adminID, _ := uuid.Parse(claims.UserID)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("user not found"))
	}

	var req changeRoleRequest
	if err := c.BodyParser(&req); err != nil || req.Role == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("role is required", httperr.ErrorDetail{Field: "role", Message: "required"}))
	}

	role := identitydomain.Role(req.Role)
	u, err := h.svc.ChangeUserRole(c.Context(), adminID, targetID, role)
	if err != nil {
		return mapModerationError(c, err)
	}
	return httperr.OK(c, toUserResponse(u))
}

// ─── Error mapper ─────────────────────────────────────────────────────────────

func mapModerationError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, commands.ErrUserNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("user not found"))
	case errors.Is(err, commands.ErrContributionNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("contribution not found"))
	case errors.Is(err, commands.ErrContributionConflict):
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict("contribution is not pending"))
	case errors.Is(err, commands.ErrNoteRequired):
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("reviewer note is required", httperr.ErrorDetail{Field: "note", Message: "required"}))
	case errors.Is(err, commands.ErrInvalidRole):
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("invalid role", httperr.ErrorDetail{Field: "role", Message: "must be 'user' or 'admin'"}))
	case errors.Is(err, moderationdomain.ErrCannotBanAdmin):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden("cannot ban another admin"))
	case errors.Is(err, moderationdomain.ErrCannotSelfDemote):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden("admin cannot change their own role"))
	default:
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
}
