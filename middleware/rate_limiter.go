package middleware

import (
	"joyo-abadi/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter creates a middleware that limits repeated requests to auth endpoints
func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,               // 5 requests
		Expiration: 1 * time.Minute, // per 1 minute
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the rate limit key
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			utils.Log.Warnf("Rate limit reached for IP: %s", c.IP())
			return c.Status(fiber.StatusTooManyRequests).Render("pages/login", fiber.Map{
				"Title": "Login",
				"Error": "Too many login attempts. Please try again later.",
			}, "base")
		},
	})
}
