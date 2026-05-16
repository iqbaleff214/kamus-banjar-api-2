package httperr

import (
	"github.com/gofiber/fiber/v2"
)

// Error codes matching OAS components/schemas/ErrorBody
const (
	CodeValidation  = "VALIDATION_ERROR"
	CodeUnauth      = "UNAUTHORIZED"
	CodeForbidden   = "FORBIDDEN"
	CodeNotFound    = "NOT_FOUND"
	CodeConflict    = "CONFLICT"
	CodeRateLimited = "RATE_LIMITED"
	CodeAIUnavail   = "AI_UNAVAILABLE"
	CodeInternal    = "INTERNAL_ERROR"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorResponse is the JSON envelope for all error responses.
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   errorBody `json:"error"`
}

// PaginationMeta holds pagination metadata for list responses.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// SuccessResponse is the JSON envelope for successful responses.
type SuccessResponse struct {
	Success bool            `json:"success"`
	Data    any             `json:"data"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

func newError(code, message string, details []ErrorDetail) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: errorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func Validation(message string, details ...ErrorDetail) *ErrorResponse {
	return newError(CodeValidation, message, details)
}

func Unauthorized(message string) *ErrorResponse {
	return newError(CodeUnauth, message, nil)
}

func Forbidden(message string) *ErrorResponse {
	return newError(CodeForbidden, message, nil)
}

func NotFound(message string) *ErrorResponse {
	return newError(CodeNotFound, message, nil)
}

func Conflict(message string) *ErrorResponse {
	return newError(CodeConflict, message, nil)
}

func RateLimited(message string) *ErrorResponse {
	return newError(CodeRateLimited, message, nil)
}

func AIUnavailable(message string) *ErrorResponse {
	return newError(CodeAIUnavail, message, nil)
}

func Internal(message string) *ErrorResponse {
	return newError(CodeInternal, message, nil)
}

// Send writes the error response to the Fiber context.
func Send(c *fiber.Ctx, status int, resp *ErrorResponse) error {
	return c.Status(status).JSON(resp)
}

// OK writes a success response without pagination.
func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(&SuccessResponse{Success: true, Data: data})
}

// Created writes a 201 success response.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(&SuccessResponse{Success: true, Data: data})
}

// Paginated writes a success response with pagination metadata.
func Paginated(c *fiber.Ctx, data any, meta *PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(&SuccessResponse{Success: true, Data: data, Meta: meta})
}

// NoContent writes a 204 response.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
