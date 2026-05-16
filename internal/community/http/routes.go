package communityhttp

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

func RegisterRoutes(app *fiber.App, h *Handler, counter ratelimit.Counter) {
	v2 := app.Group("/api/v2")

	// ─── Contributions (auth required) ────────────────────────────────────────
	contrib := v2.Group("/contributions", auth.RequireAuth())
	contrib.Get("/", h.ListContributions)
	contrib.Post("/",
		ratelimit.Limiter(counter, func(c *fiber.Ctx) string {
			claims := auth.GetClaims(c)
			if claims == nil {
				return "ratelimit:contributions:ip:" + c.IP()
			}
			return "ratelimit:contributions:" + claims.UserID
		}, 10, time.Hour),
		h.SubmitContribution,
	)
	contrib.Get("/:id", h.GetContribution)
	contrib.Patch("/:id/withdraw", h.WithdrawContribution)
	contrib.Patch("/:id/approve", auth.RequireRole("admin"), h.ApproveContribution)
	contrib.Patch("/:id/reject", auth.RequireRole("admin"), h.RejectContribution)

	// ─── Votes on words (auth required for writes) ────────────────────────────
	wordVotes := v2.Group("/words/:id/votes")
	wordVotes.Post("/", auth.RequireAuth(), h.CastWordVote)
	wordVotes.Delete("/", auth.RequireAuth(), h.RemoveWordVote)

	// ─── Votes on definitions (auth required for writes) ──────────────────────
	defVotes := v2.Group("/definitions/:id/votes")
	defVotes.Post("/", auth.RequireAuth(), h.CastDefinitionVote)
	defVotes.Delete("/", auth.RequireAuth(), h.RemoveDefinitionVote)

	// ─── Bookmarks (auth required) ────────────────────────────────────────────
	bm := v2.Group("/bookmarks", auth.RequireAuth())
	bm.Get("/", h.ListBookmarks)
	bm.Post("/", h.AddBookmark)
	bm.Delete("/:word_id", h.RemoveBookmark)

	// ─── Comments (reads public, writes auth required) ────────────────────────
	wordComments := v2.Group("/words/:id/comments")
	wordComments.Get("/", h.ListComments)
	wordComments.Post("/", auth.RequireAuth(), h.PostComment)

	comments := v2.Group("/comments", auth.RequireAuth())
	comments.Patch("/:id", h.EditComment)
	comments.Delete("/:id", h.DeleteComment)
	comments.Post("/:id/flag", h.FlagComment)
}
