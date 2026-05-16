package dictionaryhttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
)

func RegisterRoutes(app *fiber.App, h *Handler) {
	v2 := app.Group("/api/v2")

	// Public dictionary routes (no auth required)
	words := v2.Group("/words")
	words.Get("/", h.ListWords)
	words.Get("/search", h.ListWords) // alias — uses ?q= param
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
