package controllers

import (
	"joyo-abadi/utils"

	"github.com/gofiber/fiber/v2"
)

// Dashboard renders the dashboard page
func Dashboard() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := utils.GetUserID(c)
		if !ok {
			return c.Redirect("/login")
		}

		utils.Log.WithField("user_id", userID).Info("User accessed dashboard")
		return c.Render("dashboard", fiber.Map{
			"Title": "Dashboard",
			// "IsAuthenticated": c.Locals("IsAuthenticated"),
		})
	}
}
