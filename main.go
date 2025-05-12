package main

import (
	"fmt"
	"joyo-abadi/models"
	"joyo-abadi/routes"
	"joyo-abadi/utils"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	utils.InitLogger()
	utils.Log.Info("Starting joyo-abadi application...")

	// Set up Fiber app with HTML template engine
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"), // Use fiber template engine
	})

	// Set up PostgreSQL connection
	db := initDatabase()

	routes.Setup(app, db)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	utils.Log.Infof("Server is running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func initDatabase() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		utils.Log.Warn("Could not load .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: &utils.GormLogger{},
	})
	if err != nil {
		utils.Log.WithError(err).Fatal("Failed to connect to database")
	}

	err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Transaction{})
	if err != nil {
		utils.Log.WithError(err).Fatal("Failed to auto-migrate database")
	}

	utils.Log.Info("Database connected and migrated successfully")
	return db
}
