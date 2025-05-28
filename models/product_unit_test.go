package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductUnitModel(t *testing.T) {
	db := setupTestDB()

	// Create test user and product
	user := User{
		Email:    "unituser@example.com",
		Password: "password123",
	}
	db.Create(&user)

	product := Product{
		Name:         "Water",
		BaseUnitName: "bottle",
		Price:        1.50,
		Quantity:     100,
		UserID:       user.ID,
		IsActive:     true,
	}
	db.Create(&product)

	t.Run("Create Product Unit", func(t *testing.T) {
		unit := ProductUnit{
			ProductID:      product.ID,
			UnitName:       "box",
			ConversionRate: 12.0,  // 1 box = 12 bottles
			Price:          18.00, // $18 per box
			Quantity:       5,     // 5 boxes in stock
			IsBaseUnit:     false,
			IsActive:       true,
			Description:    "Box of 12 bottles",
		}

		result := db.Create(&unit)
		assert.NoError(t, result.Error)
		assert.NotZero(t, unit.ID)
		assert.Equal(t, product.ID, unit.ProductID)
		assert.Equal(t, "box", unit.UnitName)
		assert.Equal(t, 12.0, unit.ConversionRate)
		assert.Equal(t, 18.00, unit.Price)
		assert.Equal(t, 5, unit.Quantity)
		assert.False(t, unit.IsBaseUnit)
		assert.True(t, unit.IsActive)
	})

	t.Run("Create Base Unit", func(t *testing.T) {
		baseUnit := ProductUnit{
			ProductID:      product.ID,
			UnitName:       "bottle",
			ConversionRate: 1.0,
			Price:          1.50,
			Quantity:       100,
			IsBaseUnit:     true,
			IsActive:       true,
			Description:    "Base unit - individual bottle",
		}

		result := db.Create(&baseUnit)
		assert.NoError(t, result.Error)
		assert.True(t, baseUnit.IsBaseUnit)
		assert.Equal(t, 1.0, baseUnit.ConversionRate)
	})

	t.Run("Product Unit with Product Relationship", func(t *testing.T) {
		unit := ProductUnit{
			ProductID:      product.ID,
			UnitName:       "dozen",
			ConversionRate: 12.0,
			Price:          15.00,
			Quantity:       10,
			IsActive:       true,
		}
		db.Create(&unit)

		// Fetch unit with product relationship
		var fetchedUnit ProductUnit
		result := db.Preload("Product").First(&fetchedUnit, unit.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, product.Name, fetchedUnit.Product.Name)
		assert.Equal(t, product.ID, fetchedUnit.Product.ID)
	})
}

func TestProductUnitMethods(t *testing.T) {
	t.Run("GetBaseQuantity", func(t *testing.T) {
		unit := ProductUnit{
			Quantity:       5,
			ConversionRate: 12.0,
		}

		baseQty := unit.GetBaseQuantity()
		assert.Equal(t, 60.0, baseQty) // 5 * 12 = 60
	})

	t.Run("GetPricePerBaseUnit", func(t *testing.T) {
		unit := ProductUnit{
			Price:          18.00,
			ConversionRate: 12.0,
		}

		pricePerBase := unit.GetPricePerBaseUnit()
		assert.Equal(t, 1.50, pricePerBase) // 18.00 / 12 = 1.50
	})

	t.Run("GetPricePerBaseUnit with Zero Conversion", func(t *testing.T) {
		unit := ProductUnit{
			Price:          10.00,
			ConversionRate: 0.0,
		}

		pricePerBase := unit.GetPricePerBaseUnit()
		assert.Equal(t, 10.00, pricePerBase) // Should return price when conversion is 0
	})

	t.Run("CanFulfillQuantity", func(t *testing.T) {
		unit := ProductUnit{
			Quantity: 10,
			IsActive: true,
		}

		assert.True(t, unit.CanFulfillQuantity(5))   // Can fulfill 5 out of 10
		assert.True(t, unit.CanFulfillQuantity(10))  // Can fulfill exactly 10
		assert.False(t, unit.CanFulfillQuantity(15)) // Cannot fulfill 15 out of 10

		// Inactive unit cannot fulfill any quantity
		unit.IsActive = false
		assert.False(t, unit.CanFulfillQuantity(1))
	})

	t.Run("ConvertToBaseUnits", func(t *testing.T) {
		unit := ProductUnit{
			ConversionRate: 12.0,
		}

		baseUnits := unit.ConvertToBaseUnits(3)
		assert.Equal(t, 36.0, baseUnits) // 3 * 12 = 36
	})

	t.Run("ConvertFromBaseUnits", func(t *testing.T) {
		unit := ProductUnit{
			ConversionRate: 12.0,
		}

		units := unit.ConvertFromBaseUnits(36.0)
		assert.Equal(t, 3, units) // 36 / 12 = 3

		// Test rounding down
		units = unit.ConvertFromBaseUnits(37.0)
		assert.Equal(t, 3, units) // 37 / 12 = 3.08... rounded down to 3
	})

	t.Run("ConvertFromBaseUnits with Zero Conversion", func(t *testing.T) {
		unit := ProductUnit{
			ConversionRate: 0.0,
		}

		units := unit.ConvertFromBaseUnits(10.0)
		assert.Equal(t, 0, units) // Should return 0 when conversion rate is 0
	})
}

func TestProductUnitQueries(t *testing.T) {
	db := setupTestDB()

	// Create test data
	user := User{Email: "queryuser@example.com", Password: "password"}
	db.Create(&user)

	product := Product{
		Name:         "Test Product",
		BaseUnitName: "piece",
		Price:        10.00,
		Quantity:     100,
		UserID:       user.ID,
		IsActive:     true,
	}
	db.Create(&product)

	units := []ProductUnit{
		{ProductID: product.ID, UnitName: "piece", ConversionRate: 1.0, Price: 10.00, Quantity: 100, IsBaseUnit: true, IsActive: true},
		{ProductID: product.ID, UnitName: "box", ConversionRate: 10.0, Price: 95.00, Quantity: 5, IsBaseUnit: false, IsActive: true},
		{ProductID: product.ID, UnitName: "case", ConversionRate: 50.0, Price: 450.00, Quantity: 2, IsBaseUnit: false, IsActive: true},
		{ProductID: product.ID, UnitName: "pallet", ConversionRate: 1000.0, Price: 8500.00, Quantity: 0, IsBaseUnit: false, IsActive: false}, // Inactive
	}

	for i, unit := range units {
		db.Create(&unit)
		// Explicitly set the pallet unit as inactive after creation
		if unit.UnitName == "pallet" {
			db.Model(&ProductUnit{}).Where("id = ?", unit.ID).Update("is_active", false)
		}
		units[i] = unit // Update the slice with the created unit (including ID)
	}

	t.Run("Find Units by Product", func(t *testing.T) {
		var productUnits []ProductUnit
		result := db.Where("product_id = ?", product.ID).Find(&productUnits)
		assert.NoError(t, result.Error)
		assert.Len(t, productUnits, 4)
	})

	t.Run("Find Active Units", func(t *testing.T) {
		var activeUnits []ProductUnit
		result := db.Where("product_id = ? AND is_active = ?", product.ID, true).Find(&activeUnits)
		assert.NoError(t, result.Error)
		assert.Len(t, activeUnits, 3)
	})

	t.Run("Find Base Unit", func(t *testing.T) {
		var baseUnit ProductUnit
		result := db.Where("product_id = ? AND is_base_unit = ?", product.ID, true).First(&baseUnit)
		assert.NoError(t, result.Error)
		assert.Equal(t, "piece", baseUnit.UnitName)
		assert.Equal(t, 1.0, baseUnit.ConversionRate)
	})

	t.Run("Find Units with Stock", func(t *testing.T) {
		var unitsWithStock []ProductUnit
		result := db.Where("product_id = ? AND quantity > ? AND is_active = ?", product.ID, 0, true).Find(&unitsWithStock)
		assert.NoError(t, result.Error)
		assert.Len(t, unitsWithStock, 3)
	})

	t.Run("Find Units by Name", func(t *testing.T) {
		var boxUnit ProductUnit
		result := db.Where("product_id = ? AND unit_name = ?", product.ID, "box").First(&boxUnit)
		assert.NoError(t, result.Error)
		assert.Equal(t, "box", boxUnit.UnitName)
		assert.Equal(t, 10.0, boxUnit.ConversionRate)
	})
}
