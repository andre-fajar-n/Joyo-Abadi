package controllers

import (
	"joyo-abadi/models"
	"joyo-abadi/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RenderLoginPage renders the login page
func RenderLoginPage(c *fiber.Ctx) error {
	return c.Render("login", nil)
}

// RenderRegisterPage renders the register page
func RenderRegisterPage(c *fiber.Ctx) error {
	return c.Render("register", nil)
}

// Login handles user login
func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginData models.User
		if err := c.BodyParser(&loginData); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
		}

		var user models.User
		if err := db.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("User not found")
		}

		// Compare hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid password")
		}

		// Success
		return c.Status(fiber.StatusOK).JSON(user)
	}
}

// Register handles user registration
func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var registerData models.User
		if err := c.BodyParser(&registerData); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse register request body")
			return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
		}

		// Hash the password before saving
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerData.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.Log.WithError(err).Error("Password hashing failed")
			return c.Status(fiber.StatusInternalServerError).SendString("Error hashing password")
		}
		registerData.Password = string(hashedPassword)

		// Save user to database
		if err := db.Create(&registerData).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to create user in database")
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
		}

		// Success
		utils.Log.WithField("email", registerData.Email).Info("User registered successfully")
		return c.Status(fiber.StatusCreated).JSON(registerData)
	}
}
