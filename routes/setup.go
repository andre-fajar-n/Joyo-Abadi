package routes

import (
	"joyo-abadi/controllers"
	"joyo-abadi/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	// Health check endpoint for Railway
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "joyo-abadi",
		})
	})

	// Public routes
	app.Get("/login", controllers.RenderLoginPage)
	app.Get("/register", controllers.RenderRegisterPage)

	// Apply rate limiting to authentication endpoints
	authLimiter := middleware.RateLimiter()
	app.Post("/login", authLimiter, controllers.Login(db))
	app.Post("/register", authLimiter, controllers.Register(db))

	app.Get("/logout", controllers.Logout())

	// Protected routes
	app.Use(middleware.RequireAuth())

	app.Get("/", controllers.Home())

	// Product routes
	app.Get("/products", controllers.ListProducts(db))
	app.Get("/products/create", controllers.RenderCreateProduct())
	app.Post("/products/create", controllers.CreateProduct(db))
	app.Get("/products/:id", controllers.GetProductDetail(db))
	app.Get("/products/edit/:id", controllers.RenderEditProduct(db))
	app.Post("/products/edit/:id", controllers.UpdateProduct(db))
	app.Delete("/products/delete/:id", controllers.DeleteProduct(db))
}
