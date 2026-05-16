package dictionaryhttp

import (
	"errors"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/queries"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/pagination"
)

type Handler struct {
	qs  *queries.WordQueryService
	cmd *commands.WordCommandService
}

func NewHandler(qs *queries.WordQueryService, cmd *commands.WordCommandService) *Handler {
	return &Handler{qs: qs, cmd: cmd}
}

// ─── Public ───────────────────────────────────────────────────────────────────

func (h *Handler) ListWords(c *fiber.Ctx) error {
	p, errResp := pagination.Parse(c)
	if errResp != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, errResp)
	}

	filter := buildFilter(c)
	q := c.Query("q")

	var result *queries.ListResult
	var err error
	if q != "" {
		result, err = h.qs.SearchWords(c.Context(), q, filter, p.Page, p.PerPage)
	} else {
		result, err = h.qs.ListWords(c.Context(), filter, p.Page, p.PerPage)
	}
	if err != nil {
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal(err.Error()))
	}

	summaries := make([]wordSummaryResponse, 0, len(result.Words))
	for _, w := range result.Words {
		summaries = append(summaries, toWordSummaryResponse(w))
	}
	return httperr.Paginated(c, summaries, &result.Meta)
}

func (h *Handler) GetWord(c *fiber.Ctx) error {
	id := c.Params("id")
	word, err := h.qs.GetWord(c.Context(), id)
	if err != nil {
		return mapWordError(c, err)
	}
	return httperr.OK(c, toWordResponse(word))
}

func (h *Handler) GetDefinitions(c *fiber.Ctx) error {
	id := c.Params("id")
	word, err := h.qs.GetWord(c.Context(), id)
	if err != nil {
		return mapWordError(c, err)
	}
	defs := make([]definitionResponse, 0, len(word.Definitions))
	for _, d := range word.Definitions {
		defs = append(defs, definitionResponse{
			ID: d.ID, Meaning: d.Meaning, SortOrder: d.SortOrder,
			Source: string(d.Source), Upvotes: d.Upvotes, Downvotes: d.Downvotes, NetScore: d.NetScore(),
		})
	}
	sort.Slice(defs, func(i, j int) bool {
		if defs[i].NetScore != defs[j].NetScore {
			return defs[i].NetScore > defs[j].NetScore
		}
		return defs[i].SortOrder < defs[j].SortOrder
	})
	return httperr.OK(c, defs)
}

func (h *Handler) GetExamples(c *fiber.Ctx) error {
	id := c.Params("id")
	word, err := h.qs.GetWord(c.Context(), id)
	if err != nil {
		return mapWordError(c, err)
	}
	exs := make([]exampleResponse, 0, len(word.Examples))
	for _, e := range word.Examples {
		exs = append(exs, exampleResponse{
			ID: e.ID, BanjarSentence: e.BanjarSentence,
			IndonesianTranslation: e.IndonesianTranslation, Source: string(e.Source),
		})
	}
	return httperr.OK(c, exs)
}

func (h *Handler) GetRelatedWords(c *fiber.Ctx) error {
	id := c.Params("id")
	word, err := h.qs.GetWord(c.Context(), id)
	if err != nil {
		return mapWordError(c, err)
	}
	related := make([]wordSummaryResponse, 0)
	for _, relID := range word.RelatedWords {
		rw, err := h.qs.GetWord(c.Context(), relID.String())
		if err != nil {
			continue
		}
		related = append(related, toWordSummaryResponse(rw))
	}
	return httperr.OK(c, related)
}

// ─── Admin ────────────────────────────────────────────────────────────────────

func (h *Handler) AdminCreateWord(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	var req wordInputRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	if req.Banjar == "" || req.WordClass == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity,
			httperr.Validation("banjar and word_class are required",
				httperr.ErrorDetail{Field: "banjar", Message: "required"},
				httperr.ErrorDetail{Field: "word_class", Message: "required"},
			))
	}

	input := toWordInput(req)
	word, err := h.cmd.CreateWord(c.Context(), claims.UserID, input)
	if err != nil {
		return mapWordError(c, err)
	}
	return httperr.Created(c, toWordResponse(word))
}

func (h *Handler) AdminUpdateWord(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	wordID := c.Params("id")
	var req wordInputRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	input := toWordInput(req)
	word, err := h.cmd.UpdateWord(c.Context(), claims.UserID, wordID, input)
	if err != nil {
		return mapWordError(c, err)
	}
	return httperr.OK(c, toWordResponse(word))
}

func (h *Handler) AdminDeleteWord(c *fiber.Ctx) error {
	wordID := c.Params("id")
	if err := h.cmd.SoftDeleteWord(c.Context(), wordID); err != nil {
		return mapWordError(c, err)
	}
	return httperr.NoContent(c)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func buildFilter(c *fiber.Ctx) domain.WordFilter {
	f := domain.WordFilter{
		Sort: c.Query("sort", "alphabetical"),
	}
	if wc := c.Query("word_class"); wc != "" {
		wct := domain.WordClass(wc)
		f.WordClass = &wct
	}
	if src := c.Query("source"); src != "" {
		s := domain.Source(src)
		f.Source = &s
	}
	switch c.Query("is_root") {
	case "true":
		t := true
		f.IsRoot = &t
	case "false":
		t := false
		f.IsRoot = &t
	}
	return f
}

func toWordInput(req wordInputRequest) commands.WordInput {
	defs := make([]commands.DefinitionInput, 0, len(req.Definitions))
	for _, d := range req.Definitions {
		defs = append(defs, commands.DefinitionInput{Meaning: d.Meaning, SortOrder: d.SortOrder})
	}
	exs := make([]commands.ExampleInput, 0, len(req.Examples))
	for _, e := range req.Examples {
		exs = append(exs, commands.ExampleInput{BanjarSentence: e.BanjarSentence, IndonesianTranslation: e.IndonesianTranslation})
	}
	dialect := req.Dialect
	if dialect == "" {
		dialect = string(domain.DialectHulu)
	}
	isRoot := req.IsRoot
	if req.RootWordID == nil {
		isRoot = true
	}
	var rootID *uuid.UUID
	if req.RootWordID != nil {
		r := *req.RootWordID
		rootID = &r
		isRoot = false
	}
	return commands.WordInput{
		Banjar:            req.Banjar,
		BanjarSyllabified: req.BanjarSyllabified,
		WordClass:         req.WordClass,
		Dialect:           dialect,
		HomonymNumber:     req.HomonymNumber,
		IsRoot:            isRoot,
		RootWordID:        rootID,
		Definitions:       defs,
		Examples:          exs,
		SourceReference:   req.SourceReference,
	}
}

func mapWordError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrWordNotFound):
		return httperr.Send(c, fiber.StatusNotFound, httperr.NotFound(err.Error()))
	case errors.Is(err, domain.ErrWordConflict):
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict(err.Error()))
	case errors.Is(err, domain.ErrEmptyBanjar),
		errors.Is(err, domain.ErrInvalidWordClass),
		errors.Is(err, domain.ErrInvalidDialect),
		errors.Is(err, domain.ErrMeaningTooLong):
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation(err.Error()))
	default:
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal server error"))
	}
}
