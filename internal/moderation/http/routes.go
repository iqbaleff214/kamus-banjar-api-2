package moderationhttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
)

func RegisterRoutes(app *fiber.App, h *Handler) {
	admin := app.Group("/api/v2/admin", auth.RequireAuth(), auth.RequireRole("admin"))

	// ─── Moderation queues ────────────────────────────────────────────────────
	mod := admin.Group("/moderation")
	mod.Get("/queue", h.GetModerationQueue)
	mod.Get("/flags", h.GetFlaggedComments)
	mod.Get("/stats", h.GetModerationStats)

	// ─── User management ──────────────────────────────────────────────────────
	users := admin.Group("/users")
	users.Get("/", h.ListUsers)
	users.Get("/:id", h.GetUser)
	users.Patch("/:id/ban", h.BanUser)
	users.Patch("/:id/unban", h.UnbanUser)
	users.Patch("/:id/role", h.ChangeUserRole)
}
