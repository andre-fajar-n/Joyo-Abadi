package utils

import (
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitSessionSimple(t *testing.T) {
	t.Run("Development Environment", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("RAILWAY_ENVIRONMENT")
		os.Unsetenv("ENV")

		InitSession()
		assert.NotNil(t, Store)
	})

	t.Run("Production Environment - Railway", func(t *testing.T) {
		os.Setenv("RAILWAY_ENVIRONMENT", "production")
		defer os.Unsetenv("RAILWAY_ENVIRONMENT")

		InitSession()
		assert.NotNil(t, Store)
	})

	t.Run("Production Environment - ENV", func(t *testing.T) {
		os.Setenv("ENV", "production")
		defer os.Unsetenv("ENV")

		InitSession()
		assert.NotNil(t, Store)
	})
}

func TestSessionOperationsSimple(t *testing.T) {
	// Initialize session store
	InitSession()

	// Create a test Fiber app
	app := fiber.New()

	t.Run("Set and Get Data by Key", func(t *testing.T) {
		app.Get("/test", func(c *fiber.Ctx) error {
			// Set data
			err := SetDataByKey(c, "test_key", "test_value")
			assert.NoError(t, err)

			// Get data
			value, err := GetDataByKey(c, "test_key")
			assert.NoError(t, err)
			assert.Equal(t, "test_value", value)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Get UserID from Session", func(t *testing.T) {
		app.Get("/user-test", func(c *fiber.Ctx) error {
			// Set user ID
			err := SetDataByKey(c, "userID", uint(123))
			assert.NoError(t, err)

			// Get user ID
			userID, exists := GetUserID(c)
			assert.True(t, exists)
			assert.Equal(t, uint(123), userID)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/user-test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Get Non-existent UserID", func(t *testing.T) {
		app.Get("/no-user", func(c *fiber.Ctx) error {
			userID, exists := GetUserID(c)
			assert.False(t, exists)
			assert.Equal(t, uint(0), userID)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/no-user", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Destroy Data by Key", func(t *testing.T) {
		app.Get("/destroy-test", func(c *fiber.Ctx) error {
			// Set data
			err := SetDataByKey(c, "destroy_key", "destroy_value")
			assert.NoError(t, err)

			// Verify data exists
			value, err := GetDataByKey(c, "destroy_key")
			assert.NoError(t, err)
			assert.Equal(t, "destroy_value", value)

			// Destroy data
			err = DestroyByKey(c, "destroy_key")
			assert.NoError(t, err)

			// Verify data is gone
			value, err = GetDataByKey(c, "destroy_key")
			assert.NoError(t, err)
			assert.Nil(t, value)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/destroy-test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
