package controllers

import (
	"joyo-abadi/utils"

	"github.com/gofiber/fiber/v2"
)

func NewHome() IHome {
	return &home{}
}

type home struct{}

// Home renders the home page
func (h *home) Home(c *fiber.Ctx) error {
	userID := utils.GetLocalUserID(c)

	utils.Log.WithField("user_id", userID).Info("User accessed home")
	return c.Render("pages/home", fiber.Map{
		"Title":           "Home",
		"IsAuthenticated": true,
	}, "base")
}
