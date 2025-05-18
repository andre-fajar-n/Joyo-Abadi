package routes

import (
	"joyo-abadi/controllers"
	"joyo-abadi/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	// Apply template data middleware to all routes
	// app.Use(middleware.TemplateData())

	// Public routes
	app.Get("/login", controllers.RenderLoginPage)
	app.Get("/register", controllers.RenderRegisterPage)

	// Apply rate limiting to authentication endpoints
	authLimiter := middleware.RateLimiter()
	app.Post("/login", authLimiter, controllers.Login(db))
	app.Post("/register", authLimiter, controllers.Register(db))

	app.Get("/logout", controllers.Logout())

	// Protected routes
	app.Use("/dashboard", middleware.RequireAuth())
	app.Get("/dashboard", controllers.Dashboard())

	// Add more protected routes here
}
