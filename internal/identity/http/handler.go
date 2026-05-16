package identityhttp

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/application/commands"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
)

type Handler struct {
	svc *commands.Service
}

func NewHandler(svc *commands.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("name, email and password are required",
			httperr.ErrorDetail{Field: "name", Message: "required"},
			httperr.ErrorDetail{Field: "email", Message: "required"},
			httperr.ErrorDetail{Field: "password", Message: "required"},
		))
	}
	user, err := h.svc.RegisterUser(c.Context(), req.Name, req.Email, req.Password, req.PasswordConfirmation)
	if err != nil {
		return mapDomainError(c, err)
	}
	return httperr.Created(c, toUserResponse(user))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	if req.Email == "" || req.Password == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("email and password are required"))
	}
	access, refresh, err := h.svc.LoginUser(c.Context(), req.Email, req.Password)
	if err != nil {
		return mapDomainError(c, err)
	}
	return httperr.OK(c, tokenPairResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    900, // 15 minutes
	})
}

func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("refresh_token is required"))
	}
	access, refresh, err := h.svc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return mapDomainError(c, err)
	}
	return httperr.OK(c, tokenPairResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    900,
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req logoutRequest
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("refresh_token is required"))
	}
	_ = h.svc.LogoutUser(c.Context(), req.RefreshToken)
	return httperr.NoContent(c)
}

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	var req verifyEmailRequest
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("token is required"))
	}
	if err := h.svc.VerifyEmail(c.Context(), req.Token); err != nil {
		return mapDomainError(c, err)
	}
	return httperr.NoContent(c)
}

func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req forgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	_ = h.svc.ForgotPassword(c.Context(), req.Email)
	return httperr.NoContent(c)
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req resetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	if err := h.svc.ResetPassword(c.Context(), req.Token, req.Password, req.PasswordConfirmation); err != nil {
		return mapDomainError(c, err)
	}
	return httperr.NoContent(c)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	if claims == nil {
		return httperr.Send(c, fiber.StatusUnauthorized, httperr.Unauthorized("authentication required"))
	}
	// We only have claims in context; return what we have.
	// Full user fetch can be added if profile data is needed beyond JWT claims.
	return httperr.OK(c, fiber.Map{
		"id":   claims.UserID,
		"role": claims.Role,
	})
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	var req updateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	user, err := h.svc.UpdateProfile(c.Context(), claims.UserID, req.Name)
	if err != nil {
		return mapDomainError(c, err)
	}
	return httperr.OK(c, toUserResponse(user))
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	claims := auth.GetClaims(c)
	var req changePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation("invalid request body"))
	}
	if err := h.svc.ChangePassword(c.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		return mapDomainError(c, err)
	}
	return httperr.NoContent(c)
}

func mapDomainError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailConflict):
		return httperr.Send(c, fiber.StatusConflict, httperr.Conflict(err.Error()))
	case errors.Is(err, domain.ErrWrongPassword),
		errors.Is(err, domain.ErrUnverifiedEmail),
		errors.Is(err, domain.ErrUserNotFound):
		return httperr.Send(c, fiber.StatusUnauthorized, httperr.Unauthorized(err.Error()))
	case errors.Is(err, domain.ErrUserBanned):
		return httperr.Send(c, fiber.StatusForbidden, httperr.Forbidden(err.Error()))
	case errors.Is(err, domain.ErrEmptyName),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, commands.ErrPasswordMismatch):
		return httperr.Send(c, fiber.StatusUnprocessableEntity, httperr.Validation(err.Error()))
	default:
		return httperr.Send(c, fiber.StatusInternalServerError, httperr.Internal("internal server error"))
	}
}
