package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

// RequestID sets X-Request-ID on every response.
// Preserves an existing client-supplied value; generates a UUID otherwise.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(headerRequestID)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(headerRequestID, id)
		return c.Next()
	}
}
