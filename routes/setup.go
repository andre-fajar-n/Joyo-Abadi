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

	// setup controllers
	authController := controllers.NewAuth(db)
	productController := controllers.NewProduct(db)
	homeController := controllers.NewHome()

	// Public routes
	app.Get("/login", authController.RenderLoginPage)
	app.Get("/register", authController.RenderRegisterPage)

	// Apply rate limiting to authentication endpoints
	authLimiter := middleware.RateLimiter()
	app.Post("/login", authLimiter, authController.Login)
	app.Post("/register", authLimiter, authController.Register)

	app.Get("/logout", authController.Logout)

	// Protected routes
	app.Use(middleware.RequireAuth())

	app.Get("/", homeController.Home)

	// Product routes
	app.Get("/products", productController.ListProducts)
	app.Get("/products/create", productController.RenderCreateProduct)
	app.Post("/products/create", productController.CreateProduct)
	app.Get("/products/:id", productController.GetProductDetail)
	app.Get("/products/edit/:id", productController.RenderEditProduct)
	app.Post("/products/edit/:id", productController.UpdateProduct)
	app.Delete("/products/delete/:id", productController.DeleteProduct)
}
