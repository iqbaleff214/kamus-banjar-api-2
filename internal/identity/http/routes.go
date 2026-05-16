package identityhttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
)

func RegisterRoutes(app *fiber.App, h *Handler) {
	v2 := app.Group("/api/v2")

	a := v2.Group("/auth")
	a.Post("/register", h.Register)
	a.Post("/login", h.Login)
	a.Post("/refresh", h.RefreshToken)
	a.Post("/logout", h.Logout)
	a.Post("/verify-email", h.VerifyEmail)
	a.Post("/forgot-password", h.ForgotPassword)
	a.Post("/reset-password", h.ResetPassword)

	me := v2.Group("/me", auth.RequireAuth())
	me.Get("/", h.Me)
	me.Put("/", h.UpdateProfile)
	me.Put("/password", h.ChangePassword)
}
