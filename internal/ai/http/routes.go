package aihttp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

func RegisterRoutes(app *fiber.App, h *Handler, ah *AdminHandler, counter ratelimit.Counter) {
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

	adminAI := v2.Group("/admin/ai", auth.RequireAuth(), auth.RequireRole("admin"))

	triggerLimiter := ratelimit.Limiter(counter, func(c *fiber.Ctx) string {
		claims := auth.GetClaims(c)
		if claims == nil {
			return "ratelimit:admin_ai:ip:" + c.IP()
		}
		return "ratelimit:admin_ai:" + claims.UserID
	}, 50, time.Hour)

	adminAI.Post("/enrich/:word_id", triggerLimiter, ah.EnrichDefinition)
	adminAI.Post("/example/:word_id", triggerLimiter, ah.SuggestExample)
	adminAI.Post("/related/:word_id", triggerLimiter, ah.SuggestRelated)
	adminAI.Post("/check/:contribution_id", triggerLimiter, ah.QualityCheck)

	adminAI.Get("/requests", ah.ListAIRequests)
	adminAI.Patch("/requests/:id/approve", ah.ApproveAIRequest)
	adminAI.Patch("/requests/:id/reject", ah.RejectAIRequest)
}
