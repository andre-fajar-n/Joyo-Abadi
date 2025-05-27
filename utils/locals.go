package utils

import "github.com/gofiber/fiber/v2"

func SetLocalsUserID(c *fiber.Ctx, userID uint) {
	c.Locals("userID", userID)
}

func GetLocalUserID(c *fiber.Ctx) uint {
	userID := c.Locals("userID")
	if userID == nil {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}
