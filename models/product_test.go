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

func TestProductWithUnits(t *testing.T) {
	db := setupTestDB()

	// Create test user
	user := User{Email: "productunits@example.com", Password: "password"}
	db.Create(&user)

	// Create product
	product := Product{
		Name:         "Water",
		BaseUnitName: "bottle",
		Price:        1.50,
		Quantity:     100,
		UserID:       user.ID,
		IsActive:     true,
	}
	db.Create(&product)

	// Create units
	units := []ProductUnit{
		{ProductID: product.ID, UnitName: "bottle", ConversionRate: 1.0, Price: 1.50, Quantity: 100, IsBaseUnit: true, IsActive: true},
		{ProductID: product.ID, UnitName: "box", ConversionRate: 12.0, Price: 18.00, Quantity: 5, IsBaseUnit: false, IsActive: true},
		{ProductID: product.ID, UnitName: "case", ConversionRate: 24.0, Price: 35.00, Quantity: 2, IsBaseUnit: false, IsActive: true},
		{ProductID: product.ID, UnitName: "pallet", ConversionRate: 1000.0, Price: 1400.00, Quantity: 0, IsBaseUnit: false, IsActive: false},
	}

	for i, unit := range units {
		db.Create(&unit)
		// Explicitly set the pallet unit as inactive after creation
		if unit.UnitName == "pallet" {
			db.Model(&ProductUnit{}).Where("id = ?", unit.ID).Update("is_active", false)
		}
		units[i] = unit // Update the slice with the created unit (including ID)
	}

	// Reload product with units
	db.Preload("Units").First(&product, product.ID)

	t.Run("GetBaseUnit", func(t *testing.T) {
		baseUnit := product.GetBaseUnit()
		assert.NotNil(t, baseUnit)
		assert.Equal(t, "bottle", baseUnit.UnitName)
		assert.True(t, baseUnit.IsBaseUnit)
		assert.Equal(t, 1.0, baseUnit.ConversionRate)
	})

	t.Run("GetActiveUnits", func(t *testing.T) {
		activeUnits := product.GetActiveUnits()
		assert.Len(t, activeUnits, 3) // bottle, box, case (pallet is inactive)
	})

	t.Run("GetUnitByName", func(t *testing.T) {
		boxUnit := product.GetUnitByName("box")
		assert.NotNil(t, boxUnit)
		assert.Equal(t, "box", boxUnit.UnitName)
		assert.Equal(t, 12.0, boxUnit.ConversionRate)

		// Test inactive unit
		palletUnit := product.GetUnitByName("pallet")
		assert.Nil(t, palletUnit) // Should be nil because it's inactive

		// Test non-existent unit
		nonExistent := product.GetUnitByName("gallon")
		assert.Nil(t, nonExistent)
	})

	t.Run("GetTotalBaseQuantity", func(t *testing.T) {
		totalBase := product.GetTotalBaseQuantity()
		// bottle: 100 * 1 = 100
		// box: 5 * 12 = 60
		// case: 2 * 24 = 48
		// pallet: 0 * 1000 = 0 (but inactive, so not counted)
		// Total: 100 + 60 + 48 = 208
		assert.Equal(t, 208.0, totalBase)
	})

	t.Run("HasStock", func(t *testing.T) {
		assert.True(t, product.HasStock())

		// Create product with no stock
		emptyProduct := Product{
			Name:         "Empty Product",
			BaseUnitName: "piece",
			UserID:       user.ID,
			IsActive:     true,
		}
		db.Create(&emptyProduct)

		emptyUnit := ProductUnit{
			ProductID:      emptyProduct.ID,
			UnitName:       "piece",
			ConversionRate: 1.0,
			Price:          10.00,
			Quantity:       0,
			IsBaseUnit:     true,
			IsActive:       true,
		}
		db.Create(&emptyUnit)

		// Reload with units
		db.Preload("Units").First(&emptyProduct, emptyProduct.ID)
		assert.False(t, emptyProduct.HasStock())
	})
}
