package pagination

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Params struct {
	Page    int
	PerPage int
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p Params) Limit() int {
	return p.PerPage
}

func TotalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := total / perPage
	if total%perPage != 0 {
		pages++
	}
	return pages
}

// Parse reads page and per_page query params from a Fiber context.
func Parse(c *fiber.Ctx) (Params, *httperr.ErrorResponse) {
	page, err := parsePositiveInt(c.Query("page", "1"), "page")
	if err != nil {
		return Params{}, err
	}
	perPage, err := parsePositiveInt(c.Query("per_page", "20"), "per_page")
	if err != nil {
		return Params{}, err
	}
	if page < 1 {
		return Params{}, httperr.Validation("page must be >= 1",
			httperr.ErrorDetail{Field: "page", Message: "must be >= 1"})
	}
	if perPage < 1 {
		return Params{}, httperr.Validation("per_page must be >= 1",
			httperr.ErrorDetail{Field: "per_page", Message: "must be >= 1"})
	}
	if perPage > MaxPerPage {
		return Params{}, httperr.Validation("per_page exceeds maximum of 100",
			httperr.ErrorDetail{Field: "per_page", Message: "must be <= 100"})
	}
	return Params{Page: page, PerPage: perPage}, nil
}

func parsePositiveInt(s, field string) (int, *httperr.ErrorResponse) {
	v, convErr := strconv.Atoi(s)
	if convErr != nil {
		return 0, httperr.Validation("invalid "+field,
			httperr.ErrorDetail{Field: field, Message: "must be a positive integer"})
	}
	return v, nil
}
