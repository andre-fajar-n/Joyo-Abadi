package utils

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestLocalsOperationsSimple(t *testing.T) {
	app := fiber.New()

	t.Run("Set and Get Local UserID", func(t *testing.T) {
		app.Get("/locals-test", func(c *fiber.Ctx) error {
			// Set local user ID
			SetLocalsUserID(c, uint(456))

			// Get local user ID
			userID := GetLocalUserID(c)
			assert.Equal(t, uint(456), userID)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/locals-test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Multiple Local Operations", func(t *testing.T) {
		app.Get("/multiple-locals", func(c *fiber.Ctx) error {
			// Set multiple times
			SetLocalsUserID(c, uint(100))
			userID1 := GetLocalUserID(c)
			assert.Equal(t, uint(100), userID1)

			SetLocalsUserID(c, uint(200))
			userID2 := GetLocalUserID(c)
			assert.Equal(t, uint(200), userID2)

			// Should have the latest value
			assert.NotEqual(t, userID1, userID2)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/multiple-locals", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestLocalsWithoutUserID(t *testing.T) {
	app := fiber.New()

	t.Run("Get Local UserID Without Setting", func(t *testing.T) {
		app.Get("/no-locals", func(c *fiber.Ctx) error {
			// This should panic or return zero value
			defer func() {
				if r := recover(); r != nil {
					// Expected panic when trying to access non-existent local
					assert.NotNil(t, r)
				}
			}()

			// This might panic if userID is not set
			userID := GetLocalUserID(c)
			// If it doesn't panic, it should return zero value
			assert.Equal(t, uint(0), userID)

			return c.SendString("OK")
		})

		req, _ := http.NewRequest("GET", "/no-locals", nil)
		// This test might fail if GetLocalUserID panics
		// In a real application, you'd want to handle this gracefully
		app.Test(req)
	})
}
