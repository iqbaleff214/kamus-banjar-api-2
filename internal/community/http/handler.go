package communityhttp

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/community/application/commands"
	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
	pg "github.com/iqbaleff214/kamus-banjar-api-2/pkg/pagination"
)

type Handler struct {
	contributions *commands.ContributionService
	votes         *commands.VoteService
	bookmarks     *commands.BookmarkService
	comments      *commands.CommentService
}

func NewHandler(
	contributions *commands.ContributionService,
	votes *commands.VoteService,
	bookmarks *commands.BookmarkService,
	comments *commands.CommentService,
) *Handler {
	return &Handler{
		contributions: contributions,
		votes:         votes,
		bookmarks:     bookmarks,
		comments:      comments,
	}
}

// ─── Contributions ────────────────────────────────────────────────────────────

func (h *Handler) SubmitContribution(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid user id"))
	}

	var req submitContributionRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	if req.Type == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("type is required", httperr.ErrorDetail{Field: "type", Message: "required"}))
	}
	if req.Payload == nil {
		req.Payload = make(map[string]any)
	}

	var targetWordID *uuid.UUID
	if req.TargetWordID != nil && *req.TargetWordID != "" {
		id, parseErr := uuid.Parse(*req.TargetWordID)
		if parseErr != nil {
			return httperr.Send(c, fiber.StatusUnprocessableEntity,
				httperr.Validation("invalid target_word_id", httperr.ErrorDetail{Field: "target_word_id", Message: "must be a valid UUID"}))
		}
		targetWordID = &id
	}

	contrib, err := h.contributions.SubmitContribution(c.Context(), userID, commands.ContributionInput{
		Type:         domain.ContributionType(req.Type),
		TargetWordID: targetWordID,
		Payload:      req.Payload,
	})
	if err != nil {
		return mapContributionError(c, err)
	}
	return httperr.Created(c, toContributionResponse(contrib))
}

func (h *Handler) ListContributions(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	callerID, _ := uuid.Parse(claims.UserID)
	p, errResp := pg.Parse(c)
	if errResp != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp)
	}
	page, perPage := p.Page, p.PerPage

	filter := domain.ContributionFilter{}
	if s := c.Query("status"); s != "" {
		st := domain.ContributionStatus(s)
		filter.Status = &st
	}
	if t := c.Query("type"); t != "" {
		ct := domain.ContributionType(t)
		filter.Type = &ct
	}

	contribs, total, err := h.contributions.ListContributions(c.Context(), callerID, claims.Role, filter, page, perPage)
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}

	data := make([]contributionResponse, 0, len(contribs))
	for _, con := range contribs {
		data = append(data, toContributionResponse(con))
	}
	return httperr.Paginated(c, data, &httperr.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	})
}

func (h *Handler) GetContribution(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	callerID, _ := uuid.Parse(claims.UserID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("contribution not found"))
	}

	contrib, err := h.contributions.GetContribution(c.Context(), callerID, claims.Role, id)
	if err != nil {
		return mapContributionError(c, err)
	}
	return httperr.OK(c, toContributionResponse(contrib))
}

func (h *Handler) WithdrawContribution(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	callerID, _ := uuid.Parse(claims.UserID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("contribution not found"))
	}

	contrib, err := h.contributions.WithdrawContribution(c.Context(), callerID, id)
	if err != nil {
		return mapContributionError(c, err)
	}
	return httperr.OK(c, toContributionResponse(contrib))
}

func (h *Handler) ApproveContribution(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("contribution not found"))
	}
	claims := auth.GetClaims(c)
	reviewerID, _ := uuid.Parse(claims.UserID)

	var req approveContributionRequest
	_ = c.BodyParser(&req)

	contrib, fetchErr := h.contributions.GetContribution(c.Context(), reviewerID, "admin", id)
	if fetchErr != nil {
		return mapContributionError(c, fetchErr)
	}
	if approveErr := contrib.Approve(reviewerID, req.Note); approveErr != nil {
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict(approveErr.Error()))
	}
	if updateErr := h.contributions.Update(c.Context(), contrib); updateErr != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.OK(c, toContributionResponse(contrib))
}

func (h *Handler) RejectContribution(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("contribution not found"))
	}
	claims := auth.GetClaims(c)
	reviewerID, _ := uuid.Parse(claims.UserID)

	var req rejectContributionRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Note) == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("note is required", httperr.ErrorDetail{Field: "note", Message: "required"}))
	}

	contrib, fetchErr := h.contributions.GetContribution(c.Context(), reviewerID, "admin", id)
	if fetchErr != nil {
		return mapContributionError(c, fetchErr)
	}
	if rejectErr := contrib.Reject(reviewerID, req.Note); rejectErr != nil {
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict(rejectErr.Error()))
	}
	if updateErr := h.contributions.Update(c.Context(), contrib); updateErr != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.OK(c, toContributionResponse(contrib))
}

// ─── Votes ────────────────────────────────────────────────────────────────────

func (h *Handler) CastWordVote(c *fiber.Ctx) error {
	return h.castVote(c, domain.VoteTargetWord)
}

func (h *Handler) RemoveWordVote(c *fiber.Ctx) error {
	return h.removeVote(c, domain.VoteTargetWord)
}

func (h *Handler) CastDefinitionVote(c *fiber.Ctx) error {
	return h.castVote(c, domain.VoteTargetDefinition)
}

func (h *Handler) RemoveDefinitionVote(c *fiber.Ctx) error {
	return h.removeVote(c, domain.VoteTargetDefinition)
}

func (h *Handler) castVote(c *fiber.Ctx, targetType domain.VoteTargetType) error {
	claims := auth.GetClaims(c)
	userID, _ := uuid.Parse(claims.UserID)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("target not found"))
	}

	var req castVoteRequest
	if err := c.BodyParser(&req); err != nil || (req.Value != "up" && req.Value != "down") {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("value must be 'up' or 'down'", httperr.ErrorDetail{Field: "value", Message: "must be 'up' or 'down'"}))
	}

	v, err := h.votes.CastVote(c.Context(), userID, targetType, targetID, domain.VoteValue(req.Value))
	if err != nil {
		if errors.Is(err, commands.ErrVoteTargetNotFound) {
			return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("target not found"))
		}
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": toVoteResponse(v)})
}

func (h *Handler) removeVote(c *fiber.Ctx, targetType domain.VoteTargetType) error {
	claims := auth.GetClaims(c)
	userID, _ := uuid.Parse(claims.UserID)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("target not found"))
	}

	if err := h.votes.RemoveVote(c.Context(), userID, targetType, targetID); err != nil {
		if errors.Is(err, commands.ErrVoteTargetNotFound) {
			return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("vote not found"))
		}
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.NoContent(c)
}

// ─── Bookmarks ────────────────────────────────────────────────────────────────

func (h *Handler) ListBookmarks(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	userID, _ := uuid.Parse(claims.UserID)
	p2, errResp2 := pg.Parse(c)
	if errResp2 != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp2)
	}
	page, perPage := p2.Page, p2.PerPage

	bookmarks, total, err := h.bookmarks.ListBookmarks(c.Context(), userID, page, perPage)
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}

	data := make([]bookmarkResponse, 0, len(bookmarks))
	for _, b := range bookmarks {
		data = append(data, toBookmarkResponse(b))
	}
	return httperr.Paginated(c, data, &httperr.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	})
}

func (h *Handler) AddBookmark(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	userID, _ := uuid.Parse(claims.UserID)

	var req addBookmarkRequest
	if err := c.BodyParser(&req); err != nil || req.WordID == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("word_id is required", httperr.ErrorDetail{Field: "word_id", Message: "required"}))
	}
	wordID, err := uuid.Parse(req.WordID)
	if err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("invalid word_id", httperr.ErrorDetail{Field: "word_id", Message: "must be a valid UUID"}))
	}

	b, err := h.bookmarks.AddBookmark(c.Context(), userID, wordID)
	if err != nil {
		if errors.Is(err, domain.ErrBookmarkConflict) {
			return httperr.Send(c, fiber.StatusConflict, httperr.Conflict("bookmark already exists"))
		}
		if errors.Is(err, commands.ErrBookmarkWordNotFound) {
			return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("word not found"))
		}
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.Created(c, toBookmarkResponse(b))
}

func (h *Handler) RemoveBookmark(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	userID, _ := uuid.Parse(claims.UserID)
	wordID, err := uuid.Parse(c.Params("word_id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("bookmark not found"))
	}

	if err := h.bookmarks.RemoveBookmark(c.Context(), userID, wordID); err != nil {
		if errors.Is(err, domain.ErrBookmarkNotFound) {
			return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("bookmark not found"))
		}
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
	return httperr.NoContent(c)
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func (h *Handler) ListComments(c *fiber.Ctx) error {
	wordID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("word not found"))
	}
	p3, errResp3 := pg.Parse(c)
	if errResp3 != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp3)
	}
	page, perPage := p3.Page, p3.PerPage

	comments, total, err := h.comments.ListComments(c.Context(), wordID, page, perPage)
	if err != nil {
		if errors.Is(err, commands.ErrCommentWordNotFound) {
			return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("word not found"))
		}
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}

	data := make([]commentResponse, 0, len(comments))
	for _, com := range comments {
		data = append(data, toCommentResponse(com))
	}
	return httperr.Paginated(c, data, &httperr.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	})
}

func (h *Handler) PostComment(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	userID, _ := uuid.Parse(claims.UserID)
	wordID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("word not found"))
	}

	var req postCommentRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("body is required", httperr.ErrorDetail{Field: "body", Message: "required"}))
	}

	comment, err := h.comments.PostComment(c.Context(), userID, wordID, req.Body)
	if err != nil {
		return mapCommentError(c, err)
	}
	return httperr.Created(c, toCommentResponse(comment))
}

func (h *Handler) EditComment(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	callerID, _ := uuid.Parse(claims.UserID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("comment not found"))
	}

	var req editCommentRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("body is required", httperr.ErrorDetail{Field: "body", Message: "required"}))
	}

	comment, err := h.comments.EditComment(c.Context(), callerID, id, req.Body)
	if err != nil {
		return mapCommentError(c, err)
	}
	return httperr.OK(c, toCommentResponse(comment))
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	callerID, _ := uuid.Parse(claims.UserID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("comment not found"))
	}

	if err := h.comments.DeleteComment(c.Context(), callerID, claims.Role, id); err != nil {
		return mapCommentError(c, err)
	}
	return httperr.NoContent(c)
}

func (h *Handler) FlagComment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("comment not found"))
	}

	comment, err := h.comments.FlagComment(c.Context(), id)
	if err != nil {
		return mapCommentError(c, err)
	}
	return httperr.OK(c, toCommentResponse(comment))
}

// ─── Error mappers ────────────────────────────────────────────────────────────

func mapContributionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrContributionNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("contribution not found"))
	case errors.Is(err, commands.ErrEmailNotVerified):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden(err.Error()))
	case errors.Is(err, commands.ErrTargetWordNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("target word not found"))
	case errors.Is(err, commands.ErrContributionForbidden):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden(err.Error()))
	case errors.Is(err, domain.ErrInvalidTransition):
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict(err.Error()))
	case errors.Is(err, domain.ErrForbidden):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden(err.Error()))
	case errors.Is(err, domain.ErrTargetRequired):
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation(err.Error(), httperr.ErrorDetail{Field: "target_word_id", Message: "required"}))
	default:
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
}

func mapCommentError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrCommentNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("comment not found"))
	case errors.Is(err, domain.ErrForbidden):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden(err.Error()))
	case errors.Is(err, domain.ErrBodyTooLong), errors.Is(err, domain.ErrEmptyBody):
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation(err.Error()))
	case errors.Is(err, commands.ErrCommentWordNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound("word not found"))
	default:
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal error"))
	}
}
