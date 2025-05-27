package utils

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInitLoggerSimple(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := Log.Out
	Log.SetOutput(&buf)
	defer Log.SetOutput(originalOutput)

	t.Run("Default Log Level", func(t *testing.T) {
		os.Unsetenv("LOG_LEVEL")
		InitLogger()
		assert.Equal(t, logrus.InfoLevel, Log.GetLevel())
	})

	t.Run("Debug Log Level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "debug")
		defer os.Unsetenv("LOG_LEVEL")

		InitLogger()
		assert.Equal(t, logrus.DebugLevel, Log.GetLevel())
	})

	t.Run("Info Log Level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "info")
		defer os.Unsetenv("LOG_LEVEL")

		InitLogger()
		assert.Equal(t, logrus.InfoLevel, Log.GetLevel())
	})

	t.Run("Warn Log Level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "warn")
		defer os.Unsetenv("LOG_LEVEL")

		InitLogger()
		assert.Equal(t, logrus.WarnLevel, Log.GetLevel())
	})

	t.Run("Error Log Level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "error")
		defer os.Unsetenv("LOG_LEVEL")

		InitLogger()
		assert.Equal(t, logrus.ErrorLevel, Log.GetLevel())
	})

	t.Run("Invalid Log Level Defaults to Info", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "invalid")
		defer os.Unsetenv("LOG_LEVEL")

		InitLogger()
		assert.Equal(t, logrus.InfoLevel, Log.GetLevel())
	})
}

func TestFiberLoggerSimple(t *testing.T) {
	// Create test app with logger middleware
	app := fiber.New()
	app.Use(FiberLogger())

	// Capture log output
	var buf bytes.Buffer
	originalOutput := Log.Out
	Log.SetOutput(&buf)
	defer Log.SetOutput(originalOutput)

	t.Run("Successful Request Logging", func(t *testing.T) {
		app.Get("/test-success", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/test-success", nil)
		req.Header.Set("User-Agent", "test-agent")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "Request completed")
		assert.Contains(t, logOutput, "GET")
		assert.Contains(t, logOutput, "/test-success")
		assert.Contains(t, logOutput, "test-agent")
	})

	t.Run("Client Error Logging", func(t *testing.T) {
		buf.Reset() // Clear previous logs

		app.Get("/test-404", func(c *fiber.Ctx) error {
			return c.Status(404).SendString("Not Found")
		})

		req, _ := http.NewRequest("GET", "/test-404", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "Client error")
		assert.Contains(t, logOutput, "404")
	})

	t.Run("Server Error Logging", func(t *testing.T) {
		buf.Reset() // Clear previous logs

		app.Get("/test-500", func(c *fiber.Ctx) error {
			return c.Status(500).SendString("Internal Server Error")
		})

		req, _ := http.NewRequest("GET", "/test-500", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "Server error")
		assert.Contains(t, logOutput, "500")
	})
}

func TestLoggerOutputSimple(t *testing.T) {
	t.Run("Log Messages", func(t *testing.T) {
		// Create a new buffer for this test
		var buf bytes.Buffer

		// Create a new logger instance for testing
		testLogger := logrus.New()
		testLogger.SetOutput(&buf)
		testLogger.SetLevel(logrus.InfoLevel)

		testLogger.Info("Test info message")
		testLogger.Warn("Test warning message")
		testLogger.Error("Test error message")

		logOutput := buf.String()
		assert.Contains(t, logOutput, "Test info message")
		assert.Contains(t, logOutput, "Test warning message")
		assert.Contains(t, logOutput, "Test error message")
	})

	t.Run("Log Fields", func(t *testing.T) {
		// Create a new buffer for this test
		var buf bytes.Buffer

		// Create a new logger instance for testing
		testLogger := logrus.New()
		testLogger.SetOutput(&buf)
		testLogger.SetLevel(logrus.InfoLevel)

		testLogger.WithFields(logrus.Fields{
			"user_id": 123,
			"action":  "login",
		}).Info("User action")

		logOutput := buf.String()
		assert.Contains(t, logOutput, "User action")
		assert.Contains(t, logOutput, "user_id=123")
		assert.Contains(t, logOutput, "action=login")
	})

	t.Run("Logger Initialization", func(t *testing.T) {
		// Test that our global logger is properly initialized
		InitLogger()
		assert.NotNil(t, Log)
		assert.Equal(t, logrus.InfoLevel, Log.GetLevel())
	})
}
