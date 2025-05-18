package middleware

import (
    "joyo-abadi/utils"
    
    "github.com/gofiber/fiber/v2"
)

// TemplateData adds common data to all templates
func TemplateData() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Check if user is authenticated
        userID, isAuthenticated := utils.GetUserID(c)
        
        // Set local variables for templates
        c.Locals("IsAuthenticated", isAuthenticated)
        if isAuthenticated {
            c.Locals("UserID", userID)
        }
        
        return c.Next()
    }
}