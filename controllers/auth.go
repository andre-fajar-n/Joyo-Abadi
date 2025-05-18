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
	utils.Log.Info("Rendering login page")
	return c.Render("pages/login", fiber.Map{
		"Title": "Login",
		// "IsAuthenticated": c.Locals("IsAuthenticated"),
	}, "base")
}

// RenderRegisterPage renders the register page
func RenderRegisterPage(c *fiber.Ctx) error {
	utils.Log.Info("Rendering register page")
	return c.Render("pages/register", fiber.Map{
		"Title": "Register",
		// "IsAuthenticated": c.Locals("IsAuthenticated"),
	}, "base")
}

// Login handles user login
func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginData models.User
		if err := c.BodyParser(&loginData); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse login request body")
			return c.Render("pages/login", fiber.Map{
				"Title": "Login",
				"Error": "Invalid login data",
			}, "base")
		}

		var user models.User
		if err := db.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
			utils.Log.WithField("email", loginData.Email).Warn("User not found during login")
			// Use generic error message to prevent user enumeration
			return c.Render("pages/login", fiber.Map{
				"Title": "Login",
				"Error": "Invalid email or password",
			}, "base")
		}

		// Compare hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
			utils.Log.WithField("email", user.Email).Warn("Invalid password attempt")
			return c.Render("pages/login", fiber.Map{
				"Title": "Login",
				"Error": "Invalid email or password",
			}, "base")
		}

		sess, err := utils.Store.Get(c)
		if err != nil {
			utils.Log.WithError(err).Error("Failed to create session")
			return c.Status(fiber.StatusInternalServerError).SendString("Session error")
		}

		// Generate a new CSRF token for the session
		sess.Set("userID", user.ID)
		if err := sess.Save(); err != nil {
			utils.Log.WithError(err).Error("Failed to save session")
			return c.Status(fiber.StatusInternalServerError).SendString("Session error")
		}

		utils.Log.WithField("email", loginData.Email).Info("User logged in successfully")
		return c.Redirect("/dashboard")
	}
}

// Register handles user registration
func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var registerData models.User
		if err := c.BodyParser(&registerData); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse register request body")
			return c.Render("pages/register", fiber.Map{
				"Title": "Register",
				"Error": "Invalid registration data",
			}, "base")
		}

		// Hash the password before saving
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerData.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.Log.WithError(err).Error("Password hashing failed")
			return c.Render("pages/register", fiber.Map{
				"Title": "Register",
				"Error": "Registration failed. Please try again.",
			}, "base")
		}
		registerData.Password = string(hashedPassword)

		// Save user to database
		if err := db.Create(&registerData).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to create user in database")
			return c.Render("pages/register", fiber.Map{
				"Title": "Register",
				"Error": "Email already exists or registration failed.",
			}, "base")
		}

		// Success - redirect to login
		utils.Log.WithField("email", registerData.Email).Info("User registered successfully")
		return c.Redirect("/login")
	}
}

func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := utils.Store.Get(c)
		if err != nil {
			utils.Log.WithError(err).Error("Failed to get session during logout")
			return c.Redirect("/login")
		}

		err = sess.Destroy()
		if err != nil {
			utils.Log.WithError(err).Error("Failed to destroy session during logout")
			return c.Redirect("/login")
		}

		utils.Log.Info("User logged out successfully")
		return c.Redirect("/login")
	}
}
