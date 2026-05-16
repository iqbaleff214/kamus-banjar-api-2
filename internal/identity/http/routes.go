package identityhttp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

func RegisterRoutes(app *fiber.App, h *Handler, counter ratelimit.Counter) {
	v2 := app.Group("/api/v2")

	loginRL := ratelimit.Limiter(counter, func(c *fiber.Ctx) string {
		return "ratelimit:login:ip:" + c.IP()
	}, 5, time.Minute)

	a := v2.Group("/auth")
	a.Post("/register", h.Register)
	a.Post("/login", loginRL, h.Login)
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
