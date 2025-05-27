package controllers

import (
	"joyo-abadi/utils"

	"github.com/gofiber/fiber/v2"
)

// Home renders the home page
func Home() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := utils.GetLocalUserID(c)

		utils.Log.WithField("user_id", userID).Info("User accessed home")
		return c.Render("pages/home", fiber.Map{
			"Title":           "Home",
			"IsAuthenticated": true,
		}, "base")
	}
}
