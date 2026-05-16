package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Init configures the global slog logger.
// JSON handler in production, text handler otherwise.
func Init(appEnv string) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler
	if appEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}

// RequestLogger returns a Fiber middleware that logs each HTTP request.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start).Milliseconds()

		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", latency,
		}
		if reqID := c.Get("X-Request-ID"); reqID != "" {
			attrs = append(attrs, "request_id", reqID)
		}
		if err != nil {
			slog.Error("request error", append(attrs, "error", err.Error())...)
		} else {
			slog.Info("request", attrs...)
		}
		return err
	}
}
