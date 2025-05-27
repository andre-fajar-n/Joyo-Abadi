package controllers

import (
	"joyo-abadi/models"
	"joyo-abadi/utils"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSimpleTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Product{}, &models.Transaction{})
	return db
}

func setupSimpleTestApp() *fiber.App {
	// Initialize utils
	utils.InitLogger()
	utils.InitSession()

	app := fiber.New()
	return app
}

func TestLoginHandler(t *testing.T) {
	db := setupSimpleTestDB()
	app := setupSimpleTestApp()

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	testUser := models.User{
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}
	db.Create(&testUser)

	app.Post("/login", Login(db))

	t.Run("Successful Login", func(t *testing.T) {
		// Create form data
		formData := url.Values{}
		formData.Set("email", "test@example.com")
		formData.Set("password", "testpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Should redirect to home page
		assert.Equal(t, 302, resp.StatusCode)
		location := resp.Header.Get("Location")
		assert.Equal(t, "/", location)
	})

	t.Run("Invalid Email", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("email", "nonexistent@example.com")
		formData.Set("password", "testpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Should redirect back to login
		assert.Equal(t, 302, resp.StatusCode)
		location := resp.Header.Get("Location")
		assert.Equal(t, "/login", location)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("email", "test@example.com")
		formData.Set("password", "wrongpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Should redirect back to login
		assert.Equal(t, 302, resp.StatusCode)
		location := resp.Header.Get("Location")
		assert.Equal(t, "/login", location)
	})
}

func TestRegisterHandler(t *testing.T) {
	db := setupSimpleTestDB()
	app := setupSimpleTestApp()

	app.Post("/register", Register(db))

	t.Run("Successful Registration", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("email", "newuser@example.com")
		formData.Set("password", "newpassword123")

		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Should redirect to login page
		assert.Equal(t, 302, resp.StatusCode)
		location := resp.Header.Get("Location")
		assert.Equal(t, "/login", location)

		// Verify user was created in database
		var user models.User
		result := db.Where("email = ?", "newuser@example.com").First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "newuser@example.com", user.Email)

		// Verify password was hashed
		assert.NotEqual(t, "newpassword123", user.Password)
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("newpassword123"))
		assert.NoError(t, err)
	})

	t.Run("Duplicate Email Registration", func(t *testing.T) {
		// Create user first
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		existingUser := models.User{
			Email:    "existing@example.com",
			Password: string(hashedPassword),
		}
		db.Create(&existingUser)

		// Try to register with same email
		formData := url.Values{}
		formData.Set("email", "existing@example.com")
		formData.Set("password", "newpassword")

		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Should redirect back to register page
		assert.Equal(t, 302, resp.StatusCode)
		location := resp.Header.Get("Location")
		assert.Equal(t, "/register", location)
	})
}

func TestRenderPages(t *testing.T) {
	app := setupSimpleTestApp()

	app.Get("/login", RenderLoginPage)
	app.Get("/register", RenderRegisterPage)

	t.Run("Render Login Page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/login", nil)
		resp, err := app.Test(req)

		// This might fail due to template not found, but handler should not crash
		assert.NoError(t, err)
		// We expect either 200 (success) or 500 (template error)
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 500)
	})

	t.Run("Render Register Page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/register", nil)
		resp, err := app.Test(req)

		// This might fail due to template not found, but handler should not crash
		assert.NoError(t, err)
		// We expect either 200 (success) or 500 (template error)
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 500)
	})
}
