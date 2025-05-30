package migrations

import (
	"joyo-abadi/models"
	"log"

	"gorm.io/gorm"
)

// MigrateProductUnits performs the migration to add product units support
func MigrateProductUnits(db *gorm.DB) error {
	log.Println("Starting product units migration...")

	// Create the product_units table
	if err := db.AutoMigrate(&models.ProductUnit{}); err != nil {
		log.Printf("Failed to create product_units table: %v", err)
		return err
	}

	// Add new columns to products table
	if err := db.AutoMigrate(&models.Product{}); err != nil {
		log.Printf("Failed to update products table: %v", err)
		return err
	}

	// Add new columns to transactions table
	if err := db.AutoMigrate(&models.Transaction{}); err != nil {
		log.Printf("Failed to update transactions table: %v", err)
		return err
	}

	// Migrate existing products to have a base unit
	if err := migrateExistingProducts(db); err != nil {
		log.Printf("Failed to migrate existing products: %v", err)
		return err
	}

	// Migrate existing transactions to include unit information
	if err := migrateExistingTransactions(db); err != nil {
		log.Printf("Failed to migrate existing transactions: %v", err)
		return err
	}

	log.Println("Product units migration completed successfully!")
	return nil
}

// migrateExistingProducts creates base units for existing products
func migrateExistingProducts(db *gorm.DB) error {
	log.Println("Migrating existing products...")

	var products []models.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}

	for _, product := range products {
		// Check if product already has units
		var unitCount int64
		if err := db.Model(&models.ProductUnit{}).Where("product_id = ?", product.ID).Count(&unitCount).Error; err != nil {
			return err
		}

		if unitCount == 0 {
			// Create base unit for existing product
			baseUnit := models.ProductUnit{
				ProductID:      product.ID,
				UnitName:       product.BaseUnitName,
				ConversionRate: 1.0, // Base unit has conversion rate of 1
				Price:          product.Price,
				Quantity:       product.Quantity,
				IsBaseUnit:     true,
				IsActive:       true,
				Description:    "Base unit (migrated from legacy product)",
			}

			if err := db.Create(&baseUnit).Error; err != nil {
				log.Printf("Failed to create base unit for product %d: %v", product.ID, err)
				return err
			}

			log.Printf("Created base unit for product: %s (ID: %d)", product.Name, product.ID)
		}
	}

	return nil
}

// migrateExistingTransactions updates existing transactions with unit information
func migrateExistingTransactions(db *gorm.DB) error {
	log.Println("Migrating existing transactions...")

	var transactions []models.Transaction
	if err := db.Preload("Product").Find(&transactions).Error; err != nil {
		return err
	}

	for _, transaction := range transactions {
		// Skip if transaction already has unit information
		if transaction.ProductUnitID != nil || transaction.UnitName != "" {
			continue
		}

		// Find the base unit for this product
		var baseUnit models.ProductUnit
		if err := db.Where("product_id = ? AND is_base_unit = ?", transaction.ProductID, true).First(&baseUnit).Error; err != nil {
			log.Printf("Warning: Could not find base unit for product %d in transaction %d", transaction.ProductID, transaction.ID)
			// Set default values for backward compatibility
			transaction.UnitName = "piece"
			transaction.UnitPrice = transaction.Total / float64(transaction.Quantity)
			transaction.BaseQuantity = float64(transaction.Quantity)
		} else {
			// Update transaction with base unit information
			transaction.ProductUnitID = &baseUnit.ID
			transaction.UnitName = baseUnit.UnitName
			transaction.UnitPrice = transaction.Total / float64(transaction.Quantity)
			transaction.BaseQuantity = float64(transaction.Quantity) * baseUnit.ConversionRate
		}

		if err := db.Save(&transaction).Error; err != nil {
			log.Printf("Failed to update transaction %d: %v", transaction.ID, err)
			return err
		}
	}

	log.Printf("Migrated %d transactions", len(transactions))
	return nil
}

// RollbackProductUnits rolls back the product units migration (for development/testing)
func RollbackProductUnits(db *gorm.DB) error {
	log.Println("Rolling back product units migration...")

	// Drop the product_units table
	if err := db.Migrator().DropTable(&models.ProductUnit{}); err != nil {
		log.Printf("Failed to drop product_units table: %v", err)
		return err
	}

	// Remove new columns from products table (this is more complex and might not be necessary)
	// Note: GORM doesn't have built-in column dropping, so this would require raw SQL
	// For now, we'll just log that manual cleanup might be needed

	log.Println("Product units migration rolled back. Note: You may need to manually remove new columns from products and transactions tables.")
	return nil
}
