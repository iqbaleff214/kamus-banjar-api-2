package aihttp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

func RegisterRoutes(app *fiber.App, h *Handler, counter ratelimit.Counter) {
	v2 := app.Group("/api/v2")

	ai := v2.Group("/ai", auth.RequireAuth())
	ai.Post("/translate",
		ratelimit.Limiter(counter, func(c *fiber.Ctx) string {
			claims := auth.GetClaims(c)
			if claims == nil {
				return "ratelimit:ai_translate:ip:" + c.IP()
			}
			return "ratelimit:ai_translate:" + claims.UserID
		}, 30, time.Hour),
		h.Translate,
	)
}
