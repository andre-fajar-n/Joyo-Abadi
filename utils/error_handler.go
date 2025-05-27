package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// ErrorHandler is a custom error handler for Fiber
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code is 500
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Log the error
	Log.WithFields(logrus.Fields{
		"status": code,
		"path":   c.Path(),
		"method": c.Method(),
		"ip":     c.IP(),
	}).WithError(err).Error("Request error")

	// Determine the response format based on Accept header
	accept := c.Get("Accept")

	// Return JSON for API requests
	if c.XHR() || accept == "application/json" {
		return c.Status(code).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// For HTML requests, render an error page
	switch code {
	case fiber.StatusNotFound:
		return c.Status(code).Render("pages/404", fiber.Map{
			"Title": "Page Not Found",
		}, "base")
	case fiber.StatusUnauthorized:
		return c.Status(code).Render("pages/401", fiber.Map{
			"Title": "Unauthorized",
		}, "base")
	default:
		return c.Status(code).Render("pages/error", fiber.Map{
			"Title":        "Error",
			"ErrorCode":    code,
			"ErrorMessage": err.Error(),
		}, "base")
	}
}
