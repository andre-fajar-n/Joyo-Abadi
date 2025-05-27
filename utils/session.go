package utils

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var Store *session.Store

func InitSession() {
	// Determine if we're in production (Railway sets RAILWAY_ENVIRONMENT)
	isProduction := os.Getenv("RAILWAY_ENVIRONMENT") == "production" || os.Getenv("ENV") == "production"

	Store = session.New(session.Config{
		Expiration:     time.Hour * 24,
		CookieSecure:   isProduction, // Only secure cookies in production with HTTPS
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

func GetDataByKey(c *fiber.Ctx, key string) (interface{}, error) {
	sess, err := Store.Get(c)
	if err != nil {
		Log.Info("Failed to get session in GetDataByKey")
		return nil, err
	}

	return sess.Get(key), nil
}

func SetDataByKey(c *fiber.Ctx, key string, value interface{}) error {
	sess, err := Store.Get(c)
	if err != nil {
		Log.Info("Failed to get session in SetDataByKey")
		return err
	}

	sess.Set(key, value)

	return sess.Save()
}

func DestroyByKey(c *fiber.Ctx, key string) error {
	sess, err := Store.Get(c)
	if err != nil {
		Log.Info("Failed to get session in DestroyByKey")
		return err
	}

	sess.Delete(key)

	return sess.Save()
}
