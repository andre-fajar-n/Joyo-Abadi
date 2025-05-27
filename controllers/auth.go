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
	errorMsg, err := utils.GetDataByKey(c, "error_login")
	if err != nil {
		utils.Log.WithError(err).Error("Failed to get error message from session")
	}

	err = utils.DestroyByKey(c, "error_login")
	if err != nil {
		utils.Log.WithError(err).Error("Failed to destroy error message from session")
	}

	utils.Log.Info("Rendering login page")
	return c.Render("pages/login", fiber.Map{
		"Title":      "Login",
		"ErrorLogin": errorMsg,
	}, "base")
}

// RenderRegisterPage renders the register page
func RenderRegisterPage(c *fiber.Ctx) error {
	errorMsg, err := utils.GetDataByKey(c, "error_register")
	if err != nil {
		utils.Log.WithError(err).Error("Failed to get error message from session")
	}

	err = utils.DestroyByKey(c, "error_register")
	if err != nil {
		utils.Log.WithError(err).Error("Failed to destroy error message from session")
	}

	utils.Log.Info("Rendering register page")
	return c.Render("pages/register", fiber.Map{
		"Title":         "Register",
		"ErrorRegister": errorMsg,
	}, "base")
}

// Login handles user login
func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginData models.User
		if err := c.BodyParser(&loginData); err != nil {
			utils.SetDataByKey(c, "error_login", "Invalid login data")
			utils.Log.WithError(err).Warn("Failed to parse login request body")
			return c.Redirect("/login")
		}

		var user models.User
		if err := db.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
			utils.SetDataByKey(c, "error_login", "Invalid email")
			utils.Log.WithField("email", loginData.Email).Warn("User not found during login")
			return c.Redirect("/login")
		}

		// Compare hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
			utils.SetDataByKey(c, "error_login", "Invalid password")
			utils.Log.WithField("email", user.Email).Warn("Invalid password attempt")
			return c.Redirect("/login")
		}

		utils.SetDataByKey(c, "userID", user.ID)
		utils.Log.WithField("email", loginData.Email).Info("User logged in successfully")
		return c.Redirect("/")
	}
}

// Register handles user registration
func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var registerData models.User
		if err := c.BodyParser(&registerData); err != nil {
			utils.SetDataByKey(c, "error_register", "Invalid registration data")
			utils.Log.WithError(err).Warn("Failed to parse register request body")
			return c.Redirect("/register")
		}

		// Hash the password before saving
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerData.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.SetDataByKey(c, "error_register", "Registration failed. Please try again.")
			utils.Log.WithError(err).Error("Password hashing failed")
			return c.Redirect("/register")
		}
		registerData.Password = string(hashedPassword)

		// Save user to database
		if err := db.Create(&registerData).Error; err != nil {
			utils.SetDataByKey(c, "error_register", "Email already exists or registration failed.")
			utils.Log.WithError(err).Error("Failed to create user in database")
			return c.Redirect("/register")
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
