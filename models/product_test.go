package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductModel(t *testing.T) {
	db := setupTestDB()

	// Create a user first for foreign key relationship
	user := User{
		Email:    "productowner@example.com",
		Password: "password123",
	}
	db.Create(&user)

	t.Run("Create Product", func(t *testing.T) {
		product := Product{
			Name:        "Test Product",
			Price:       99.99,
			Quantity:    10,
			UserID:      user.ID,
			Description: "A test product",
		}

		result := db.Create(&product)
		assert.NoError(t, result.Error)
		assert.NotZero(t, product.ID)
		assert.Equal(t, "Test Product", product.Name)
		assert.Equal(t, 99.99, product.Price)
		assert.Equal(t, 10, product.Quantity)
		assert.Equal(t, user.ID, product.UserID)
	})

	t.Run("Product with User Relationship", func(t *testing.T) {
		product := Product{
			Name:     "Related Product",
			Price:    50.00,
			Quantity: 5,
			UserID:   user.ID,
		}
		db.Create(&product)

		// Fetch product with user relationship
		var fetchedProduct Product
		result := db.Preload("User").First(&fetchedProduct, product.ID)
		assert.NoError(t, result.Error)
		assert.NotNil(t, fetchedProduct.User)
		assert.Equal(t, user.Email, fetchedProduct.User.Email)
	})

	t.Run("Product without User", func(t *testing.T) {
		product := Product{
			Name:     "Orphan Product",
			Price:    25.00,
			Quantity: 3,
			UserID:   0, // No user assigned
		}

		result := db.Create(&product)
		assert.NoError(t, result.Error)
		assert.Equal(t, uint(0), product.UserID)
	})
}

func TestProductValidation(t *testing.T) {
	db := setupTestDB()

	t.Run("Empty Name Allowed by GORM", func(t *testing.T) {
		product := Product{
			Name:     "",
			Price:    10.00,
			Quantity: 1,
		}
		result := db.Create(&product)
		// GORM allows empty strings by default, business logic should validate
		assert.NoError(t, result.Error)
		assert.Equal(t, "", product.Name)
	})

	t.Run("Zero Price Should Be Allowed", func(t *testing.T) {
		product := Product{
			Name:     "Free Product",
			Price:    0,
			Quantity: 1,
		}
		result := db.Create(&product)
		assert.NoError(t, result.Error)
	})

	t.Run("Negative Quantity Should Be Allowed", func(t *testing.T) {
		product := Product{
			Name:     "Negative Stock",
			Price:    10.00,
			Quantity: -1,
		}
		result := db.Create(&product)
		assert.NoError(t, result.Error)
	})
}

func TestProductQueries(t *testing.T) {
	db := setupTestDB()

	// Create test user
	user := User{
		Email:    "queryuser@example.com",
		Password: "password123",
	}
	db.Create(&user)

	// Create test products
	products := []Product{
		{Name: "Product A", Price: 10.00, Quantity: 5, UserID: user.ID},
		{Name: "Product B", Price: 20.00, Quantity: 3, UserID: user.ID},
		{Name: "Product C", Price: 15.00, Quantity: 0, UserID: user.ID},
	}

	for _, product := range products {
		db.Create(&product)
	}

	t.Run("Find Products by User", func(t *testing.T) {
		var userProducts []Product
		result := db.Where("user_id = ?", user.ID).Find(&userProducts)
		assert.NoError(t, result.Error)
		assert.Len(t, userProducts, 3)
	})

	t.Run("Find Products with Stock", func(t *testing.T) {
		var inStockProducts []Product
		result := db.Where("quantity > ?", 0).Find(&inStockProducts)
		assert.NoError(t, result.Error)
		assert.Len(t, inStockProducts, 2)
	})

	t.Run("Find Products by Price Range", func(t *testing.T) {
		var midRangeProducts []Product
		result := db.Where("price BETWEEN ? AND ?", 10.00, 20.00).Find(&midRangeProducts)
		assert.NoError(t, result.Error)
		assert.Len(t, midRangeProducts, 3)
	})
}
