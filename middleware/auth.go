package middleware

import (
	"joyo-abadi/utils"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth middleware checks if user is authenticated
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if user is authenticated
		userID, isAuthenticated := utils.GetUserID(c)

		// Set local variables for templates
		utils.SetLocalsUserID(c, userID)
		if !isAuthenticated {
			sess, _ := utils.Store.Get(c)
			sess.Set("error_login", "Your session has expired. Please login again.")
			sess.Save()
			utils.Log.Warn("Unauthorized access attempt")
			return c.Redirect("/login")
		}

		return c.Next()
	}
}
