package middleware

import "github.com/gofiber/fiber/v2"

func PingHealthCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		message := "pong!!"
		return c.JSON(message)
	}
}
