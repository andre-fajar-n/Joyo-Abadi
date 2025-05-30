package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionModel(t *testing.T) {
	db := setupTestDB()

	// Create test user and product
	user := User{
		Email:    "transactionuser@example.com",
		Password: "password123",
	}
	db.Create(&user)

	product := Product{
		Name:     "Transaction Product",
		Price:    50.00,
		Quantity: 10,
		UserID:   user.ID,
	}
	db.Create(&product)

	t.Run("Create Sale Transaction", func(t *testing.T) {
		transaction := Transaction{
			ProductID: product.ID,
			Quantity:  2,
			Total:     100.00,
			Type:      "sale",
		}

		result := db.Create(&transaction)
		assert.NoError(t, result.Error)
		assert.NotZero(t, transaction.ID)
		assert.Equal(t, product.ID, transaction.ProductID)
		assert.Equal(t, 2, transaction.Quantity)
		assert.Equal(t, 100.00, transaction.Total)
		assert.Equal(t, "sale", transaction.Type)
	})

	t.Run("Create Purchase Transaction", func(t *testing.T) {
		transaction := Transaction{
			ProductID: product.ID,
			Quantity:  5,
			Total:     250.00,
			Type:      "purchase",
		}

		result := db.Create(&transaction)
		assert.NoError(t, result.Error)
		assert.Equal(t, "purchase", transaction.Type)
	})

	t.Run("Transaction with Product Relationship", func(t *testing.T) {
		transaction := Transaction{
			ProductID: product.ID,
			Quantity:  1,
			Total:     50.00,
			Type:      "sale",
		}
		db.Create(&transaction)

		// Fetch transaction with product relationship
		var fetchedTransaction Transaction
		result := db.Preload("Product").First(&fetchedTransaction, transaction.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, product.Name, fetchedTransaction.Product.Name)
		assert.Equal(t, product.Price, fetchedTransaction.Product.Price)
	})
}

func TestTransactionQueries(t *testing.T) {
	db := setupTestDB()

	// Create test data
	user := User{
		Email:    "queryuser@example.com",
		Password: "password123",
	}
	db.Create(&user)

	product1 := Product{Name: "Product 1", Price: 10.00, Quantity: 10, UserID: user.ID}
	product2 := Product{Name: "Product 2", Price: 20.00, Quantity: 10, UserID: user.ID}
	db.Create(&product1)
	db.Create(&product2)

	transactions := []Transaction{
		{ProductID: product1.ID, Quantity: 2, Total: 20.00, Type: "sale"},
		{ProductID: product1.ID, Quantity: 5, Total: 50.00, Type: "purchase"},
		{ProductID: product2.ID, Quantity: 1, Total: 20.00, Type: "sale"},
		{ProductID: product2.ID, Quantity: 3, Total: 60.00, Type: "purchase"},
	}

	for _, transaction := range transactions {
		db.Create(&transaction)
	}

	t.Run("Find Transactions by Type", func(t *testing.T) {
		var sales []Transaction
		result := db.Where("type = ?", "sale").Find(&sales)
		assert.NoError(t, result.Error)
		assert.Len(t, sales, 2)

		var purchases []Transaction
		result = db.Where("type = ?", "purchase").Find(&purchases)
		assert.NoError(t, result.Error)
		assert.Len(t, purchases, 2)
	})

	t.Run("Find Transactions by Product", func(t *testing.T) {
		var product1Transactions []Transaction
		result := db.Where("product_id = ?", product1.ID).Find(&product1Transactions)
		assert.NoError(t, result.Error)
		assert.Len(t, product1Transactions, 2)
	})

	t.Run("Calculate Total Sales", func(t *testing.T) {
		var totalSales float64
		result := db.Model(&Transaction{}).Where("type = ?", "sale").Select("SUM(total)").Scan(&totalSales)
		assert.NoError(t, result.Error)
		assert.Equal(t, 40.00, totalSales)
	})

	t.Run("Calculate Total Purchases", func(t *testing.T) {
		var totalPurchases float64
		result := db.Model(&Transaction{}).Where("type = ?", "purchase").Select("SUM(total)").Scan(&totalPurchases)
		assert.NoError(t, result.Error)
		assert.Equal(t, 110.00, totalPurchases)
	})
}

func TestTransactionValidation(t *testing.T) {
	db := setupTestDB()

	// Create test product
	user := User{Email: "test@example.com", Password: "password"}
	db.Create(&user)

	product := Product{Name: "Test Product", Price: 10.00, Quantity: 5, UserID: user.ID}
	db.Create(&product)

	t.Run("Valid Transaction Types", func(t *testing.T) {
		validTypes := []string{"sale", "purchase"}

		for _, transactionType := range validTypes {
			transaction := Transaction{
				ProductID: product.ID,
				Quantity:  1,
				Total:     10.00,
				Type:      transactionType,
			}
			result := db.Create(&transaction)
			assert.NoError(t, result.Error)
		}
	})

	t.Run("Zero Quantity Transaction", func(t *testing.T) {
		transaction := Transaction{
			ProductID: product.ID,
			Quantity:  0,
			Total:     0.00,
			Type:      "sale",
		}
		result := db.Create(&transaction)
		assert.NoError(t, result.Error) // GORM allows this, business logic should validate
	})
}

func TestTransactionWithUnits(t *testing.T) {
	db := setupTestDB()

	// Create test data
	user := User{Email: "transactionunits@example.com", Password: "password"}
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

	// Create units
	bottleUnit := ProductUnit{
		ProductID:      product.ID,
		UnitName:       "bottle",
		ConversionRate: 1.0,
		Price:          1.50,
		Quantity:       100,
		IsBaseUnit:     true,
		IsActive:       true,
	}
	db.Create(&bottleUnit)

	boxUnit := ProductUnit{
		ProductID:      product.ID,
		UnitName:       "box",
		ConversionRate: 12.0,
		Price:          18.00,
		Quantity:       5,
		IsBaseUnit:     false,
		IsActive:       true,
	}
	db.Create(&boxUnit)

	t.Run("Create Transaction with Specific Unit", func(t *testing.T) {
		transaction := Transaction{
			ProductID:     product.ID,
			ProductUnitID: &boxUnit.ID,
			UnitName:      boxUnit.UnitName,
			Quantity:      2,
			UnitPrice:     boxUnit.Price,
			BaseQuantity:  float64(2) * boxUnit.ConversionRate, // 2 * 12 = 24
			Type:          "sale",
			Notes:         "Sold 2 boxes",
		}
		transaction.CalculateTotal()

		result := db.Create(&transaction)
		assert.NoError(t, result.Error)
		assert.NotZero(t, transaction.ID)
		assert.Equal(t, boxUnit.ID, *transaction.ProductUnitID)
		assert.Equal(t, "box", transaction.UnitName)
		assert.Equal(t, 2, transaction.Quantity)
		assert.Equal(t, 18.00, transaction.UnitPrice)
		assert.Equal(t, 36.00, transaction.Total) // 2 * 18.00
		assert.Equal(t, 24.0, transaction.BaseQuantity)
		assert.Equal(t, "sale", transaction.Type)
	})

	t.Run("Transaction with Unit Relationship", func(t *testing.T) {
		transaction := Transaction{
			ProductID:     product.ID,
			ProductUnitID: &bottleUnit.ID,
			UnitName:      bottleUnit.UnitName,
			Quantity:      10,
			UnitPrice:     bottleUnit.Price,
			BaseQuantity:  float64(10) * bottleUnit.ConversionRate,
			Type:          "purchase",
		}
		transaction.CalculateTotal()
		db.Create(&transaction)

		// Fetch transaction with relationships
		var fetchedTransaction Transaction
		result := db.Preload("Product").Preload("ProductUnit").First(&fetchedTransaction, transaction.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, product.Name, fetchedTransaction.Product.Name)
		assert.Equal(t, bottleUnit.UnitName, fetchedTransaction.ProductUnit.UnitName)
		assert.Equal(t, bottleUnit.ConversionRate, fetchedTransaction.ProductUnit.ConversionRate)
	})

	t.Run("GetEffectiveUnitName", func(t *testing.T) {
		// Transaction with ProductUnit relationship
		transaction1 := Transaction{
			ProductUnitID: &boxUnit.ID,
			ProductUnit:   &boxUnit,
			UnitName:      "old_name",
		}
		assert.Equal(t, "box", transaction1.GetEffectiveUnitName())

		// Transaction with only UnitName field
		transaction2 := Transaction{
			UnitName: "dozen",
		}
		assert.Equal(t, "dozen", transaction2.GetEffectiveUnitName())

		// Transaction with no unit information
		transaction3 := Transaction{}
		assert.Equal(t, "piece", transaction3.GetEffectiveUnitName())
	})

	t.Run("CalculateTotal", func(t *testing.T) {
		transaction := Transaction{
			Quantity:  5,
			UnitPrice: 12.50,
		}
		transaction.CalculateTotal()
		assert.Equal(t, 62.50, transaction.Total) // 5 * 12.50
	})

	t.Run("IsValidType", func(t *testing.T) {
		validSale := Transaction{Type: "sale"}
		assert.True(t, validSale.IsValidType())

		validPurchase := Transaction{Type: "purchase"}
		assert.True(t, validPurchase.IsValidType())

		invalid := Transaction{Type: "invalid"}
		assert.False(t, invalid.IsValidType())

		empty := Transaction{Type: ""}
		assert.False(t, empty.IsValidType())
	})
}
