package middleware

import (
	"joyo-abadi/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// APILogger logs detailed information for API requests
func APILogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only log API routes
		if !strings.HasPrefix(c.Path(), "/api") {
			return c.Next()
		}

		start := time.Now()

		// Get request body
		reqBody := string(c.Body())

		// Store the original response body writer
		originalResponseBody := c.Response().Body()

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get response body
		respBody := string(c.Response().Body())

		// Create log fields
		fields := logrus.Fields{
			"status":       c.Response().StatusCode(),
			"duration_ms":  float64(duration.Nanoseconds()) / 1e6,
			"ip":           c.IP(),
			"method":       c.Method(),
			"path":         c.Path(),
			"request_id":   c.Get("X-Request-ID"),
			"request_body": reqBody,
			"response":     respBody,
		}

		// Log based on status code
		if c.Response().StatusCode() >= 500 {
			utils.Log.WithFields(fields).Error("API server error")
		} else if c.Response().StatusCode() >= 400 {
			utils.Log.WithFields(fields).Warn("API client error")
		} else {
			utils.Log.WithFields(fields).Info("API request completed")
		}

		// Restore the original response body
		c.Response().SetBody(originalResponseBody)

		return err
	}
}
