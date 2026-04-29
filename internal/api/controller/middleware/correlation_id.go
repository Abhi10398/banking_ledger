package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awesomeProject/internal/constants"
)

const CorrelationIDHeader = "X-Correlation-ID"

// CorrelationIDMiddleware reads X-Correlation-ID from the request header.
// If absent a new UUID is generated. The value is written back into the
// response header and stored in the request context so logger.Get(ctx)
// automatically includes it in every log entry.
func CorrelationIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		corrID := c.Get(CorrelationIDHeader)
		if corrID == "" {
			corrID = uuid.New().String()
		}
		c.Set(CorrelationIDHeader, corrID)
		ctx := context.WithValue(c.UserContext(), constants.CorrelationId, corrID)
		c.SetUserContext(ctx)
		return c.Next()
	}
}
