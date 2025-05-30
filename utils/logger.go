package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

// InitLogger configures the global logger
func InitLogger() {
	// Set output to stdout
	Log.SetOutput(os.Stdout)

	// Configure formatter with colors and full timestamp
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	// Set log level from environment or default to debug for development
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		// Default to DEBUG in development to see SQL queries
		if os.Getenv("ENV") == "production" {
			Log.SetLevel(logrus.InfoLevel)
		} else {
			Log.SetLevel(logrus.DebugLevel)
		}
	}

	Log.Info("Logger initialized")
}

// FiberLogger is a custom middleware for Fiber that logs requests using logrus
func FiberLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Store the request path for logging after request is processed
		path := c.Path()
		method := c.Method()
		ip := c.IP()

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get status code
		status := c.Response().StatusCode()

		// Create log fields
		fields := logrus.Fields{
			"status":     status,
			"duration":   fmt.Sprintf("%.3fms", float64(duration.Nanoseconds())/1e6),
			"ip":         ip,
			"method":     method,
			"path":       path,
			"user_agent": c.Get("User-Agent"),
		}

		// Log based on status code
		if status >= 500 {
			Log.WithFields(fields).Error("Server error")
		} else if status >= 400 {
			Log.WithFields(fields).Warn("Client error")
		} else {
			Log.WithFields(fields).Info("Request completed")
		}

		return err
	}
}
