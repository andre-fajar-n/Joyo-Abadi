package controllers

import (
	"joyo-abadi/models"
	"joyo-abadi/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ListProductUnits renders the product units management page
func ListProductUnits(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := utils.GetLocalUserID(c)

		productID, err := strconv.ParseUint(c.Params("productId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		// Verify product ownership
		var product models.Product
		if err := db.Preload("Units").Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found or access denied")
			return c.Redirect("/products")
		}

		// Prepare units with calculated values
		unitsWithCalc := make([]fiber.Map, len(product.Units))
		totalBaseQuantity := 0.0
		activeCount := 0
		var minPrice float64 = 999999.0
		var maxPrice float64 = 0.0

		for i, unit := range product.Units {
			pricePerBase := unit.GetPricePerBaseUnit()
			baseQuantity := unit.GetBaseQuantity()

			if unit.IsActive {
				activeCount++
				totalBaseQuantity += baseQuantity
				if pricePerBase < minPrice {
					minPrice = pricePerBase
				}
				if pricePerBase > maxPrice {
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

		// Reset prices if no active units found
		if activeCount == 0 {
			minPrice = float64(0)
			maxPrice = float64(0)
		}

		// Calculate savings percentage
		savingsPercent := 0.0
		if maxPrice > 0 && minPrice < maxPrice {
			savingsPercent = ((maxPrice - minPrice) / maxPrice) * 100
		}

		return c.Render("pages/products/units", fiber.Map{
			"Title":             "Manage Product Units",
			"Product":           product,
			"Units":             product.Units,
			"UnitsWithCalc":     unitsWithCalc,
			"TotalUnits":        len(product.Units),
			"ActiveUnits":       activeCount,
			"TotalBaseQuantity": totalBaseQuantity,
			"MinPrice":          minPrice,
			"MaxPrice":          maxPrice,
			"SavingsPercent":    savingsPercent,
			"IsAuthenticated":   true,
		}, "base")
	}
}

// RenderCreateProductUnit renders the create product unit form
func RenderCreateProductUnit(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := utils.GetLocalUserID(c)

		productID, err := strconv.ParseUint(c.Params("productId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		// Verify product ownership
		var product models.Product
		if err := db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found or access denied")
			return c.Redirect("/products")
		}

		return c.Render("pages/products/create_unit", fiber.Map{
			"Title":           "Add Product Unit",
			"Product":         product,
			"IsAuthenticated": true,
		}, "base")
	}
}

// CreateProductUnit handles the creation of a new product unit
func CreateProductUnit(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := utils.GetLocalUserID(c)

		productID, err := strconv.ParseUint(c.Params("productId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		// Verify product ownership
		var product models.Product
		if err := db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found or access denied")
			return c.Redirect("/products")
		}

		var unit models.ProductUnit
		if err := c.BodyParser(&unit); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse unit data")
			return c.Status(fiber.StatusBadRequest).Render("pages/products/create_unit", fiber.Map{
				"Title":           "Add Product Unit",
				"Product":         product,
				"Error":           "Invalid unit data",
				"Unit":            unit,
				"IsAuthenticated": true,
			}, "base")
		}

		// Set product ID
		unit.ProductID = uint(productID)

		// Debug logging to see what we received
		utils.Log.WithFields(map[string]interface{}{
			"unit_name":       unit.UnitName,
			"conversion_rate": unit.ConversionRate,
			"price":           unit.Price,
			"quantity":        unit.Quantity,
			"description":     unit.Description,
		}).Debug("Received unit data from form")

		// Validate unit data using validator
		validationErrors := utils.ValidateStruct(unit)
		if len(validationErrors) > 0 {
			errorMessage := "Validation errors: " + strings.Join(validationErrors, ", ")
			utils.Log.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"unit_data":         unit,
			}).Warn("Unit validation failed")

			return c.Status(fiber.StatusBadRequest).Render("pages/products/create_unit", fiber.Map{
				"Title":           "Add Product Unit",
				"Product":         product,
				"Error":           errorMessage,
				"Unit":            unit,
				"IsAuthenticated": true,
			}, "base")
		}

		// Check for duplicate unit names
		var existingUnit models.ProductUnit
		if err := db.Where("product_id = ? AND unit_name = ? AND is_active = ?", productID, unit.UnitName, true).First(&existingUnit).Error; err == nil {
			return c.Status(fiber.StatusBadRequest).Render("pages/products/create_unit", fiber.Map{
				"Title":           "Add Product Unit",
				"Product":         product,
				"Error":           "A unit with this name already exists for this product",
				"Unit":            unit,
				"IsAuthenticated": true,
			}, "base")
		}

		// Set default values
		if unit.Quantity < 0 {
			unit.Quantity = 0
		}
		unit.IsActive = true

		if err := db.Create(&unit).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to create product unit")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/create_unit", fiber.Map{
				"Title":           "Add Product Unit",
				"Product":         product,
				"Error":           "Failed to create product unit",
				"Unit":            unit,
				"IsAuthenticated": true,
			}, "base")
		}

		utils.Log.WithFields(map[string]interface{}{
			"product_id":      productID,
			"unit_id":         unit.ID,
			"unit_name":       unit.UnitName,
			"conversion_rate": unit.ConversionRate,
			"user_id":         userID,
		}).Info("Product unit created successfully")

		return c.Redirect("/products/" + c.Params("productId") + "/units")
	}
}

// RenderEditProductUnit renders the edit product unit form
func RenderEditProductUnit(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := utils.GetLocalUserID(c)

		productID, err := strconv.ParseUint(c.Params("productId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		unitID, err := strconv.ParseUint(c.Params("unitId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid unit ID")
			return c.Redirect("/products/" + c.Params("productId") + "/units")
		}

		// Verify product ownership and get unit
		var unit models.ProductUnit
		if err := db.Preload("Product").Where("id = ? AND product_id = ?", unitID, productID).First(&unit).Error; err != nil {
			utils.Log.WithError(err).Warn("Unit not found")
			return c.Redirect("/products/" + c.Params("productId") + "/units")
		}

		if unit.Product.UserID != userID {
			utils.Log.WithFields(map[string]interface{}{
				"unit_id":         unitID,
				"product_id":      productID,
				"product_user_id": unit.Product.UserID,
				"current_user_id": userID,
			}).Warn("Unauthorized attempt to edit product unit")
			return c.Status(fiber.StatusForbidden).Render("pages/error", fiber.Map{
				"Title":        "Forbidden",
				"ErrorCode":    403,
				"ErrorMessage": "You don't have permission to edit this unit",
			}, "base")
		}

		// Calculate values for template
		pricePerBase := unit.GetPricePerBaseUnit()
		baseQuantity := unit.GetBaseQuantity()
		stockValue := unit.Price * float64(unit.Quantity)

		return c.Render("pages/products/edit_unit", fiber.Map{
			"Title":           "Edit Product Unit",
			"Product":         unit.Product,
			"Unit":            unit,
			"PricePerBase":    pricePerBase,
			"BaseQuantity":    baseQuantity,
			"StockValue":      stockValue,
			"IsAuthenticated": true,
		}, "base")
	}
}

// UpdateProductUnit handles the update of a product unit
func UpdateProductUnit(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := utils.GetLocalUserID(c)

		productID, err := strconv.ParseUint(c.Params("productId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid product ID")
			return c.Redirect("/products")
		}

		unitID, err := strconv.ParseUint(c.Params("unitId"), 10, 32)
		if err != nil {
			utils.Log.WithError(err).Warn("Invalid unit ID")
			return c.Redirect("/products/" + c.Params("productId") + "/units")
		}

		// Get existing unit and verify ownership
		var unit models.ProductUnit
		if err := db.Preload("Product").Where("id = ? AND product_id = ?", unitID, productID).First(&unit).Error; err != nil {
			utils.Log.WithError(err).Warn("Unit not found")
			return c.Redirect("/products/" + c.Params("productId") + "/units")
		}

		if unit.Product.UserID != userID {
			utils.Log.WithFields(map[string]interface{}{
				"unit_id":         unitID,
				"product_id":      productID,
				"product_user_id": unit.Product.UserID,
				"current_user_id": userID,
			}).Warn("Unauthorized attempt to update product unit")
			return c.Status(fiber.StatusForbidden).Render("pages/error", fiber.Map{
				"Title":        "Forbidden",
				"ErrorCode":    403,
				"ErrorMessage": "You don't have permission to update this unit",
			}, "base")
		}

		// Store old values for logging
		oldUnitName := unit.UnitName
		oldConversionRate := unit.ConversionRate
		oldPrice := unit.Price
		oldQuantity := unit.Quantity

		// Parse updated values
		var updatedUnit models.ProductUnit
		if err := c.BodyParser(&updatedUnit); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse unit data")
			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit_unit", fiber.Map{
				"Title":           "Edit Product Unit",
				"Product":         unit.Product,
				"Unit":            unit,
				"Error":           "Invalid unit data",
				"IsAuthenticated": true,
			}, "base")
		}

		// Update fields
		unit.UnitName = updatedUnit.UnitName
		unit.ConversionRate = updatedUnit.ConversionRate
		unit.Price = updatedUnit.Price
		unit.Quantity = updatedUnit.Quantity
		unit.Description = updatedUnit.Description
		unit.IsActive = updatedUnit.IsActive

		// Debug logging to see what we received
		utils.Log.WithFields(map[string]interface{}{
			"unit_name":       unit.UnitName,
			"conversion_rate": unit.ConversionRate,
			"price":           unit.Price,
			"quantity":        unit.Quantity,
			"description":     unit.Description,
			"is_active":       unit.IsActive,
		}).Debug("Received updated unit data from form")

		// Validate unit data using validator
		validationErrors := utils.ValidateStruct(unit)
		if len(validationErrors) > 0 {
			errorMessage := "Validation errors: " + strings.Join(validationErrors, ", ")
			utils.Log.WithFields(map[string]interface{}{
				"validation_errors": validationErrors,
				"unit_data":         unit,
			}).Warn("Unit update validation failed")

			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit_unit", fiber.Map{
				"Title":           "Edit Product Unit",
				"Product":         unit.Product,
				"Unit":            unit,
				"Error":           errorMessage,
				"IsAuthenticated": true,
			}, "base")
		}

		// Check for duplicate unit names (excluding current unit)
		var existingUnit models.ProductUnit
		if err := db.Where("product_id = ? AND unit_name = ? AND is_active = ? AND id != ?", productID, unit.UnitName, true, unitID).First(&existingUnit).Error; err == nil {
			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit_unit", fiber.Map{
				"Title":           "Edit Product Unit",
				"Product":         unit.Product,
				"Unit":            unit,
				"Error":           "A unit with this name already exists for this product",
				"IsAuthenticated": true,
			}, "base")
		}

		if err := db.Save(&unit).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to update product unit")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/edit_unit", fiber.Map{
				"Title":           "Edit Product Unit",
				"Product":         unit.Product,
				"Unit":            unit,
				"Error":           "Failed to update product unit",
				"IsAuthenticated": true,
			}, "base")
		}

		utils.Log.WithFields(map[string]interface{}{
			"unit_id":                 unitID,
			"product_id":              productID,
			"user_id":                 userID,
			"name_changed":            oldUnitName != unit.UnitName,
			"conversion_rate_changed": oldConversionRate != unit.ConversionRate,
			"price_changed":           oldPrice != unit.Price,
			"quantity_changed":        oldQuantity != unit.Quantity,
		}).Info("Product unit updated successfully")

		return c.Redirect("/products/" + c.Params("productId") + "/units")
	}
}
