package utils

import "github.com/gofiber/fiber/v2"

func SetLocalsUserID(c *fiber.Ctx, userID uint) {
	c.Locals("userID", userID)
}

func GetLocalUserID(c *fiber.Ctx) uint {
	return c.Locals("userID").(uint)
}
