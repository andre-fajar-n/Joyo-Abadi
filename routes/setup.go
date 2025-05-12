package routes

import (
	"joyo-abadi/controllers"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	// app.Get("/", controllers.ShowHome)

	// Auth routes
	app.Get("/login", controllers.RenderLoginPage)
	app.Post("/login", controllers.Login(db))

	app.Get("/register", controllers.RenderRegisterPage)
	app.Post("/register", controllers.Register(db))
}
