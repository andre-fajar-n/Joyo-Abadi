package controllers

import (
	"joyo-abadi/models"
	"joyo-abadi/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ListProducts renders the products list page
func ListProducts(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current user ID
		userID := utils.GetLocalUserID(c)

		var products []models.Product

		// Query products with user information
		if err := db.Preload("User").Find(&products).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to fetch products")
			return c.Status(fiber.StatusInternalServerError).Render("pages/error", fiber.Map{
				"Title":        "Error",
				"ErrorCode":    500,
				"ErrorMessage": "Failed to fetch products",
			}, "base")
		}

		return c.Render("pages/products/index", fiber.Map{
			"Title":           "Products",
			"Products":        products,
			"CurrentUserID":   userID,
			"IsAuthenticated": true,
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

		// Validate product data
		if product.Name == "" || product.Price <= 0 {
			return c.Status(fiber.StatusBadRequest).Render("pages/products/create", fiber.Map{
				"Title":           "Create Product",
				"Error":           "Product name and price are required. Price must be greater than 0.",
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		if err := db.Create(&product).Error; err != nil {
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
		if err := db.First(&product, id).Error; err != nil {
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

		return c.Render("pages/products/edit", fiber.Map{
			"Title":           "Edit Product",
			"Product":         product,
			"IsAuthenticated": true,
		}, "base")
	}
}

// UpdateProduct handles the product update
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
		if err := db.First(&product, id).Error; err != nil {
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

		// Store the old values for logging
		oldName := product.Name
		oldPrice := product.Price
		oldDescription := product.Description

		// Parse the updated values
		if err := c.BodyParser(&product); err != nil {
			utils.Log.WithError(err).Warn("Failed to parse product data")
			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Invalid product data",
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		// Ensure the user ID doesn't change
		product.UserID = userID

		// Validate product data
		if product.Name == "" || product.Price <= 0 {
			return c.Status(fiber.StatusBadRequest).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Product name and price are required. Price must be greater than 0.",
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		if err := db.Save(&product).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to update product")
			return c.Status(fiber.StatusInternalServerError).Render("pages/products/edit", fiber.Map{
				"Title":           "Edit Product",
				"Error":           "Failed to update product",
				"Product":         product,
				"IsAuthenticated": true,
			}, "base")
		}

		utils.Log.WithFields(map[string]interface{}{
			"product_id":          product.ID,
			"user_id":             userID,
			"name_changed":        oldName != product.Name,
			"price_changed":       oldPrice != product.Price,
			"description_changed": oldDescription != product.Description,
			"old_name":            oldName,
			"new_name":            product.Name,
			"old_price":           oldPrice,
			"new_price":           product.Price,
		}).Info("Product updated successfully")

		return c.Redirect("/products")
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

		// Check if product is used in any transactions
		var count int64
		if err := db.Model(&models.Transaction{}).Where("product_id = ?", id).Count(&count).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to check product usage in transactions")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to check product usage",
			})
		}

		if count > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Cannot delete product that is used in transactions",
			})
		}

		if err := db.Delete(&product).Error; err != nil {
			utils.Log.WithError(err).Error("Failed to delete product")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to delete product",
			})
		}

		utils.Log.WithFields(map[string]interface{}{
			"product_id": id,
			"user_id":    userID,
		}).Info("Product deleted successfully")

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Product deleted successfully",
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
		if err := db.Preload("User").First(&product, id).Error; err != nil {
			utils.Log.WithError(err).Warn("Product not found")
			return c.Redirect("/products")
		}

		return c.Render("pages/products/detail", fiber.Map{
			"Title":           "Product Detail",
			"Product":         product,
			"CurrentUserID":   userID,
			"IsAuthenticated": true,
		}, "base")
	}
}
