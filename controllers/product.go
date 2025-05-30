package controllers

import (
	"fmt"
	"joyo-abadi/models"
	"joyo-abadi/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UnitFormData represents unit data from form
type UnitFormData struct {
	ID             uint
	UnitName       string
	ConversionRate float64
	Price          float64
	Quantity       int
	Description    string
	IsActive       bool
	Delete         bool
}

// ListProducts renders the products list page
func ListProducts(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID := utils.GetLocalUserID(c)

		var products []models.Product

		// Query products with user information and units
		if err := db.Preload("User").Preload("Units").Where("is_active = ?", true).Find(&products).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to fetch products")
			return c.Status(fiber.StatusInternalServerError).Render("pages/error", fiber.Map{
				"Title":        "Error",
				"ErrorCode":    500,
				"ErrorMessage": "Failed to fetch products",
			}, "base")
		}

		// Prepare products with calculated values
		productsWithCalc := make([]fiber.Map, len(products))
		for i, product := range products {
			activeUnits := product.GetActiveUnits()
			totalBaseQuantity := product.GetTotalBaseQuantity()
			hasStock := product.HasStock()

			productsWithCalc[i] = fiber.Map{
				"ID":                product.ID,
				"Name":              product.Name,
				"BaseUnitName":      product.BaseUnitName,
				"Price":             product.Price,
				"Quantity":          product.Quantity,
				"UserID":            product.UserID,
				"User":              product.User,
				"Description":       product.Description,
				"IsActive":          product.IsActive,
				"Units":             product.Units,
				"ActiveUnits":       activeUnits,
				"ActiveUnitCount":   len(activeUnits),
				"TotalUnitCount":    len(product.Units),
				"TotalBaseQuantity": totalBaseQuantity,
				"HasStock":          hasStock,
				"CreatedAt":         product.CreatedAt,
				"UpdatedAt":         product.UpdatedAt,
			}
		}

		return c.Render("pages/products/index", fiber.Map{
			"Title":            "Products",
			"Products":         products,
			"ProductsWithCalc": productsWithCalc,
			"CurrentUserID":    userID,
			"IsAuthenticated":  true,
		}, "base")
	}
}

// RenderCreateProduct renders the create product form
func RenderCreateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("pages/products/create", fiber.Map{
			"Title":           "Create Product",
			"IsAuthenticated": true,
		}, "base")
	}
}

// CreateProduct handles the product creation
func CreateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID := utils.GetLocalUserID(c)

		var product models.Product
		if err := c.BodyParser(&product); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse product data")
			return c.Status(fiber.StatusBadRequest).Render("pages/products/create", fiber.Map{
				"Title":           "Create Product",
				"Error":           "Invalid product data",
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		// Set the user ID for the product
		product.UserID = userID
		product.IsActive = true

		// Set default base unit name if not provided
		if product.BaseUnitName == "" {
			product.BaseUnitName = "piece"
		}

		// Validate product data using validator
		validationErrors := utils.ValidateStruct(product)
		if len(validationErrors) > 0 {
			errorMessage := "Validation errors: " + strings.Join(validationErrors, ", ")
			utils.Log.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"product_data":      product,
			}).Warn("Product validation failed")

			return c.Status(fiber.StatusBadRequest).Render("pages/products/create", fiber.Map{
				"Title":           "Create Product",
				"Error":           errorMessage,
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		// Create product in a transaction to ensure base unit is created
		err := db.Transaction(func(tx *gorm.DB) error {
			// Create the product
			if err := tx.Create(&product).Error; err != nil {
				return err
			}

			// Create the base unit
			baseUnit := models.ProductUnit{
				ProductID:      product.ID,
				UnitName:       product.BaseUnitName,
				ConversionRate: 1.0, // Base unit always has conversion rate of 1
				Price:          product.Price,
				Quantity:       product.Quantity,
				IsBaseUnit:     true,
				IsActive:       true,
				Description:    "Base unit",
			}

			return tx.Create(&baseUnit).Error
		})

		if err != nil {
			utils.Log.WithError(err).Error("Failed to create product")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/create", fiber.Map{
				"Title":           "Create Product",
				"Error":           "Failed to create product",
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		utils.Log.WithFields(map[string]interface{}{
			"product_id": product.ID,
			"user_id":    userID,
		}).Info("Product created successfully")

		return c.Redirect("/products")
	}
}

// RenderEditProduct renders the edit product form
func RenderEditProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID := utils.GetLocalUserID(c)

		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		var product models.Product
		if err := db.Preload("Units").First(&product, id).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found")
			return c.Redirect("/products")
		}

		// Check if the user owns this product or is an admin
		// You can implement admin check here if needed
		if product.UserID != userID {
			utils.Log.WithFields(map[string]interface{}{
				"product_id":       product.ID,
				"product_owner_id": product.UserID,
				"current_user_id":  userID,
			}).Warn("Unauthorized attempt to edit product")

			return c.Status(fiber.StatusForbidden).Render("pages/error", fiber.Map{
				"Title":        "Forbidden",
				"ErrorCode":    403,
				"ErrorMessage": "You don't have permission to edit this product",
			}, "base")
		}

		// Prepare units with calculated values (similar to detail page)
		units := product.Units
		unitsWithCalc := make([]fiber.Map, len(units))
		totalUnits := len(units)
		activeUnits := 0
		totalBaseQuantity := 0.0
		var minPrice, maxPrice float64

		for i, unit := range units {
			pricePerBase := unit.GetPricePerBaseUnit()
			baseQuantity := unit.GetBaseQuantity()
			totalBaseQuantity += baseQuantity

			if unit.IsActive {
				activeUnits++
				if activeUnits == 1 || pricePerBase < minPrice {
					minPrice = pricePerBase
				}
				if activeUnits == 1 || pricePerBase > maxPrice {
					maxPrice = pricePerBase
				}
			}

			unitsWithCalc[i] = fiber.Map{
				"ID":             unit.ID,
				"UnitName":       unit.UnitName,
				"ConversionRate": unit.ConversionRate,
				"Price":          unit.Price,
				"Quantity":       unit.Quantity,
				"IsBaseUnit":     unit.IsBaseUnit,
				"IsActive":       unit.IsActive,
				"Description":    unit.Description,
				"PricePerBase":   pricePerBase,
				"BaseQuantity":   baseQuantity,
				"CreatedAt":      unit.CreatedAt,
				"UpdatedAt":      unit.UpdatedAt,
			}
		}

		// Calculate savings percentage
		savingsPercent := 0.0
		if activeUnits > 1 && maxPrice > minPrice {
			savingsPercent = ((maxPrice - minPrice) / maxPrice) * 100
		}

		return c.Render("pages/products/edit", fiber.Map{
			"Title":             "Edit Product",
			"Product":           product,
			"Units":             units,
			"UnitsWithCalc":     unitsWithCalc,
			"TotalUnits":        totalUnits,
			"ActiveUnits":       activeUnits,
			"TotalBaseQuantity": totalBaseQuantity,
			"MinPrice":          minPrice,
			"MaxPrice":          maxPrice,
			"SavingsPercent":    savingsPercent,
			"IsAuthenticated":   true,
		}, "base")
	}
}

// UpdateProduct handles the product and units update in a single transaction
func UpdateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID := utils.GetLocalUserID(c)

		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		var product models.Product
		if err := db.Preload("Units").First(&product, id).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found")
			return c.Redirect("/products")
		}

		// Check if the user owns this product or is an admin
		if product.UserID != userID {
			utils.Log.WithFields(map[string]interface{}{
				"product_id":       product.ID,
				"product_owner_id": product.UserID,
				"current_user_id":  userID,
			}).Warn("Unauthorized attempt to update product")

			return c.Status(fiber.StatusForbidden).Render("pages/error", fiber.Map{
				"Title":        "Forbidden",
				"ErrorCode":    403,
				"ErrorMessage": "You don't have permission to update this product",
			}, "base")
		}

		// Parse form data
		formData := make(map[string]interface{})
		if err := c.BodyParser(&formData); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse form data")
			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Invalid form data",
				"Product":         product,
				"Units":           product.Units,
				"IsAuthenticated": true,
			}, "base")
		}

		// Store old values for logging
		oldName := product.Name
		oldPrice := product.Price
		oldQuantity := product.Quantity
		oldDescription := product.Description

		// Update product fields
		if name, ok := formData["name"].(string); ok {
			product.Name = name
		}
		if baseUnitName, ok := formData["base_unit_name"].(string); ok {
			product.BaseUnitName = baseUnitName
		}
		if priceStr, ok := formData["price"].(string); ok {
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				product.Price = price
			}
		}
		if quantityStr, ok := formData["quantity"].(string); ok {
			if quantity, err := strconv.Atoi(quantityStr); err == nil {
				product.Quantity = quantity
			}
		}
		if description, ok := formData["description"].(string); ok {
			product.Description = description
		}

		// Ensure the user ID doesn't change
		product.UserID = userID

		// Validate product data
		validationErrors := utils.ValidateStruct(product)
		if len(validationErrors) > 0 {
			errorMessage := "Product validation errors: " + strings.Join(validationErrors, ", ")
			utils.Log.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"product_data":      product,
			}).Warn("Product update validation failed")

			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           errorMessage,
				"Product":         product,
				"Units":           product.Units,
				"IsAuthenticated": true,
			}, "base")
		}

		// Parse units data
		unitsData := parseUnitsFromForm(c)

		// Validate units data
		var allValidationErrors []string
		for i, unitData := range unitsData {
			unit := models.ProductUnit{
				UnitName:       unitData.UnitName,
				ConversionRate: unitData.ConversionRate,
				Price:          unitData.Price,
				Quantity:       unitData.Quantity,
				Description:    unitData.Description,
				IsActive:       unitData.IsActive,
				ProductID:      uint(id),
			}

			unitValidationErrors := utils.ValidateStruct(unit)
			if len(unitValidationErrors) > 0 {
				for _, err := range unitValidationErrors {
					allValidationErrors = append(allValidationErrors, fmt.Sprintf("Unit %d: %s", i+1, err))
				}
			}
		}

		if len(allValidationErrors) > 0 {
			errorMessage := "Validation errors: " + strings.Join(allValidationErrors, ", ")
			utils.Log.WithFields(map[string]interface{}{
				"validation_errors": allValidationErrors,
				"units_data":        unitsData,
			}).Warn("Units validation failed")

			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           errorMessage,
				"Product":         product,
				"Units":           product.Units,
				"IsAuthenticated": true,
			}, "base")
		}

		// Start transaction
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Update product
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			utils.Log.WithError(err).Error("Failed to update product")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Failed to update product",
				"Product":         product,
				"Units":           product.Units,
				"IsAuthenticated": true,
			}, "base")
		}

		// Handle units
		if err := updateProductUnits(tx, uint(id), unitsData); err != nil {
			tx.Rollback()
			utils.Log.WithError(err).Error("Failed to update product units")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Failed to update product units: " + err.Error(),
				"Product":         product,
				"Units":           product.Units,
				"IsAuthenticated": true,
			}, "base")
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			utils.Log.WithError(err).Error("Failed to commit transaction")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Failed to save changes",
				"Product":         product,
				"Units":           product.Units,
				"IsAuthenticated": true,
			}, "base")
		}

		utils.Log.WithFields(map[string]interface{}{
			"product_id":          product.ID,
			"user_id":             userID,
			"name_changed":        oldName != product.Name,
			"price_changed":       oldPrice != product.Price,
			"qty_changed":         oldQuantity != product.Quantity,
			"description_changed": oldDescription != product.Description,
			"units_count":         len(unitsData),
		}).Info("Product and units updated successfully")

		return c.Redirect("/products/" + strconv.FormatUint(uint64(product.ID), 10))
	}
}

// DeleteProduct handles the product deletion
func DeleteProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID, isAuthenticated := utils.GetUserID(c)
		if !isAuthenticated {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Unauthorized",
			})
		}

		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid product ID",
			})
		}

		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Product not found",
			})
		}

		// Check if the user owns this product or is an admin
		if product.UserID != userID {
			utils.Log.WithFields(map[string]interface{}{
				"product_id":       product.ID,
				"product_owner_id": product.UserID,
				"current_user_id":  userID,
			}).Warn("Unauthorized attempt to delete product")

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "You don't have permission to delete this product",
			})
		}

		// Start transaction for safe deletion
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Check if product is used in any transactions
		var transactionCount int64
		if err := tx.Model(&models.Transaction{}).Where("product_id = ?", id).Count(&transactionCount).Error; err != nil {
			tx.Rollback()
			utils.Log.WithError(err).Error("Failed to check product usage in transactions")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to check product usage",
			})
		}

		if transactionCount > 0 {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Cannot delete product that is used in transactions",
			})
		}

		// Get count of units for logging
		var unitCount int64
		if err := tx.Model(&models.ProductUnit{}).Where("product_id = ?", id).Count(&unitCount).Error; err != nil {
			tx.Rollback()
			utils.Log.WithError(err).Error("Failed to count product units")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to check product units",
			})
		}

		// Delete all product units first (due to foreign key constraints)
		if err := tx.Where("product_id = ?", id).Delete(&models.ProductUnit{}).Error; err != nil {
			tx.Rollback()
			utils.Log.WithError(err).Error("Failed to delete product units")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to delete product units",
			})
		}

		// Delete the product
		if err := tx.Delete(&product).Error; err != nil {
			tx.Rollback()
			utils.Log.WithError(err).Error("Failed to delete product")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to delete product",
			})
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			utils.Log.WithError(err).Error("Failed to commit product deletion transaction")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to complete product deletion",
			})
		}

		utils.Log.WithFields(map[string]interface{}{
			"product_id":    id,
			"product_name":  product.Name,
			"user_id":       userID,
			"units_deleted": unitCount,
		}).Info("Product and all related units deleted successfully")

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Product and all related units deleted successfully",
		})
	}
}

// GetProductDetail renders the product detail page
func GetProductDetail(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID := utils.GetLocalUserID(c)

		id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		var product models.Product
		if err := db.Preload("User").Preload("Units").First(&product, id).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found")
			return c.Redirect("/products")
		}

		// Prepare units with calculated values
		units := product.GetActiveUnits()
		unitsWithCalc := make([]fiber.Map, len(units))
		for i, unit := range units {
			pricePerBase := unit.GetPricePerBaseUnit()
			baseQuantity := unit.GetBaseQuantity()

			unitsWithCalc[i] = fiber.Map{
				"ID":             unit.ID,
				"UnitName":       unit.UnitName,
				"ConversionRate": unit.ConversionRate,
				"Price":          unit.Price,
				"Quantity":       unit.Quantity,
				"IsBaseUnit":     unit.IsBaseUnit,
				"IsActive":       unit.IsActive,
				"Description":    unit.Description,
				"PricePerBase":   pricePerBase,
				"BaseQuantity":   baseQuantity,
				"CreatedAt":      unit.CreatedAt,
				"UpdatedAt":      unit.UpdatedAt,
			}
		}

		return c.Render("pages/products/detail", fiber.Map{
			"Title":             "Product Detail",
			"Product":           product,
			"Units":             units,
			"UnitsWithCalc":     unitsWithCalc,
			"TotalBaseQuantity": product.GetTotalBaseQuantity(),
			"HasStock":          product.HasStock(),
			"CurrentUserID":     userID,
			"IsAuthenticated":   true,
		}, "base")
	}
}

// parseUnitsFromForm parses unit data from form submission
func parseUnitsFromForm(c *fiber.Ctx) []UnitFormData {
	var units []UnitFormData

	// Get all form values
	formData := c.AllParams()

	// Parse units array from form
	i := 0
	for {
		// Check if this unit index exists
		unitNameKey := fmt.Sprintf("units[%d][unit_name]", i)
		unitName := c.FormValue(unitNameKey)

		if unitName == "" {
			break // No more units
		}

		unit := UnitFormData{
			UnitName: unitName,
		}

		// Parse ID
		if idStr := c.FormValue(fmt.Sprintf("units[%d][id]", i)); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				unit.ID = uint(id)
			}
		}

		// Parse conversion rate
		if convStr := c.FormValue(fmt.Sprintf("units[%d][conversion_rate]", i)); convStr != "" {
			if conv, err := strconv.ParseFloat(convStr, 64); err == nil {
				unit.ConversionRate = conv
			}
		}

		// Parse price
		if priceStr := c.FormValue(fmt.Sprintf("units[%d][price]", i)); priceStr != "" {
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				unit.Price = price
			}
		}

		// Parse quantity
		if qtyStr := c.FormValue(fmt.Sprintf("units[%d][quantity]", i)); qtyStr != "" {
			if qty, err := strconv.Atoi(qtyStr); err == nil {
				unit.Quantity = qty
			}
		}

		// Parse description
		unit.Description = c.FormValue(fmt.Sprintf("units[%d][description]", i))

		// Parse is_active
		isActiveStr := c.FormValue(fmt.Sprintf("units[%d][is_active]", i))
		unit.IsActive = isActiveStr == "true"

		// Parse delete flag
		deleteStr := c.FormValue(fmt.Sprintf("units[%d][delete]", i))
		unit.Delete = deleteStr == "true"

		units = append(units, unit)
		i++
	}

	// Log parsed units for debugging
	utils.Log.WithFields(map[string]interface{}{
		"units_count": len(units),
		"form_data":   formData,
	}).Debug("Parsed units from form")

	return units
}

// updateProductUnits handles creating, updating, and deleting product units
func updateProductUnits(tx *gorm.DB, productID uint, unitsData []UnitFormData) error {
	// Get existing units
	var existingUnits []models.ProductUnit
	if err := tx.Where("product_id = ?", productID).Find(&existingUnits).Error; err != nil {
		return fmt.Errorf("failed to fetch existing units: %w", err)
	}

	// Create a map of existing units by ID
	existingUnitsMap := make(map[uint]*models.ProductUnit)
	for i := range existingUnits {
		existingUnitsMap[existingUnits[i].ID] = &existingUnits[i]
	}

	// Track which existing units are being updated
	updatedUnitIDs := make(map[uint]bool)

	// Process each unit from form
	for _, unitData := range unitsData {
		if unitData.Delete {
			// Delete unit if it exists
			if unitData.ID > 0 {
				if err := tx.Delete(&models.ProductUnit{}, unitData.ID).Error; err != nil {
					return fmt.Errorf("failed to delete unit %d: %w", unitData.ID, err)
				}
				utils.Log.WithFields(map[string]interface{}{
					"unit_id":    unitData.ID,
					"unit_name":  unitData.UnitName,
					"product_id": productID,
				}).Info("Product unit deleted")
			}
			continue
		}

		if unitData.ID > 0 {
			// Update existing unit
			if existingUnit, exists := existingUnitsMap[unitData.ID]; exists {
				existingUnit.UnitName = unitData.UnitName
				existingUnit.ConversionRate = unitData.ConversionRate
				existingUnit.Price = unitData.Price
				existingUnit.Quantity = unitData.Quantity
				existingUnit.Description = unitData.Description
				existingUnit.IsActive = unitData.IsActive

				if err := tx.Save(existingUnit).Error; err != nil {
					return fmt.Errorf("failed to update unit %d: %w", unitData.ID, err)
				}

				updatedUnitIDs[unitData.ID] = true

				utils.Log.WithFields(map[string]interface{}{
					"unit_id":    unitData.ID,
					"unit_name":  unitData.UnitName,
					"product_id": productID,
				}).Info("Product unit updated")
			}
		} else {
			// Create new unit
			newUnit := models.ProductUnit{
				ProductID:      productID,
				UnitName:       unitData.UnitName,
				ConversionRate: unitData.ConversionRate,
				Price:          unitData.Price,
				Quantity:       unitData.Quantity,
				Description:    unitData.Description,
				IsActive:       unitData.IsActive,
				IsBaseUnit:     false, // New units are never base units
			}

			if err := tx.Create(&newUnit).Error; err != nil {
				return fmt.Errorf("failed to create unit: %w", err)
			}

			utils.Log.WithFields(map[string]interface{}{
				"unit_id":    newUnit.ID,
				"unit_name":  unitData.UnitName,
				"product_id": productID,
			}).Info("Product unit created")
		}
	}

	return nil
}
