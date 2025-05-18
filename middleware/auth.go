package middleware

import (
    "joyo-abadi/utils"
    
    "github.com/gofiber/fiber/v2"
)

// RequireAuth middleware checks if user is authenticated
func RequireAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        _, ok := utils.GetUserID(c)
        if !ok {
            utils.Log.Warn("Unauthorized access attempt")
            return c.Redirect("/login")
        }
        return c.Next()
    }
}