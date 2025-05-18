package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var Store *session.Store

func InitSession() {
	Store = session.New(session.Config{
		Expiration:     time.Hour * 24,
		CookieSecure:   true, // Ensure this is true in production with HTTPS
		CookieHTTPOnly: true,
		CookieSameSite: "Lax", // Helps with CSRF protection
	})
}

// GetUserID retrieves the authenticated user ID from session
func GetUserID(c *fiber.Ctx) (uint, bool) {
	sess, err := Store.Get(c)
	if err != nil {
		return 0, false
	}

	userID, ok := sess.Get("userID").(uint)
	return userID, ok
}
