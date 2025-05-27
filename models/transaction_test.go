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
