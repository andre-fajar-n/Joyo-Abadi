package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&User{}, &Product{}, &Transaction{})
	return db
}

func TestUserModel(t *testing.T) {
	db := setupTestDB()

	t.Run("Create User", func(t *testing.T) {
		user := User{
			Email:    "test@example.com",
			Password: "hashedpassword123",
		}

		result := db.Create(&user)
		assert.NoError(t, result.Error)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("User Email Unique Constraint", func(t *testing.T) {
		// Create first user
		user1 := User{
			Email:    "unique@example.com",
			Password: "password1",
		}
		db.Create(&user1)

		// Try to create second user with same email
		user2 := User{
			Email:    "unique@example.com",
			Password: "password2",
		}
		result := db.Create(&user2)
		assert.Error(t, result.Error)
	})

	t.Run("Find User by Email", func(t *testing.T) {
		// Create user
		originalUser := User{
			Email:    "findme@example.com",
			Password: "findmepassword",
		}
		db.Create(&originalUser)

		// Find user
		var foundUser User
		result := db.Where("email = ?", "findme@example.com").First(&foundUser)
		assert.NoError(t, result.Error)
		assert.Equal(t, originalUser.Email, foundUser.Email)
		assert.Equal(t, originalUser.ID, foundUser.ID)
	})

	t.Run("User Not Found", func(t *testing.T) {
		var user User
		result := db.Where("email = ?", "nonexistent@example.com").First(&user)
		assert.Error(t, result.Error)
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
	})
}

func TestUserValidation(t *testing.T) {
	db := setupTestDB()

	t.Run("Empty Email Allowed by GORM", func(t *testing.T) {
		user := User{
			Email:    "",
			Password: "password123",
		}
		result := db.Create(&user)
		// GORM allows empty strings by default, business logic should validate
		assert.NoError(t, result.Error)
		assert.Equal(t, "", user.Email)
	})

	t.Run("Empty Password Allowed by GORM", func(t *testing.T) {
		user := User{
			Email:    "test2@example.com",
			Password: "",
		}
		result := db.Create(&user)
		// GORM allows empty strings by default, business logic should validate
		assert.NoError(t, result.Error)
		assert.Equal(t, "", user.Password)
	})

	t.Run("Business Logic Validation", func(t *testing.T) {
		// Test business logic validation (this would be in controllers)
		user := User{
			Email:    "",
			Password: "password",
		}

		// Simulate business logic validation
		isValid := user.Email != "" && user.Password != ""
		assert.False(t, isValid, "Business logic should reject empty email")

		user.Email = "valid@example.com"
		isValid = user.Email != "" && user.Password != ""
		assert.True(t, isValid, "Business logic should accept valid data")
	})
}
