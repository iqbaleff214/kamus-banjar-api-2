package dictionaryhttp

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

func RegisterRoutes(app *fiber.App, h *Handler, counter ratelimit.Counter) {
	v2 := app.Group("/api/v2")

	// Public dictionary routes — TryAuth so authenticated users get higher limit
	wordsRL := ratelimit.AdaptiveLimiter(counter, func(c *fiber.Ctx) (string, int) {
		claims := auth.GetClaims(c)
		if claims == nil {
			return "ratelimit:words:ip:" + c.IP(), 60
		}
		return "ratelimit:words:user:" + claims.UserID, 120
	}, time.Minute)

	words := v2.Group("/words", auth.TryAuth(), wordsRL)
	words.Get("/", h.ListWords)
	words.Get("/search", h.ListWords)
	words.Get("/:id", h.GetWord)
	words.Get("/:id/definitions", h.GetDefinitions)
	words.Get("/:id/examples", h.GetExamples)
	words.Get("/:id/related", h.GetRelatedWords)

	// Admin word management (auth + admin role required)
	admin := v2.Group("/admin/words", auth.RequireAuth(), auth.RequireRole("admin"))
	admin.Post("/", h.AdminCreateWord)
	admin.Patch("/:id", h.AdminUpdateWord)
	admin.Delete("/:id", h.AdminDeleteWord)
}
