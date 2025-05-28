package models

import "gorm.io/gorm"

// Transaction represents a sale or purchase transaction
type Transaction struct {
	gorm.Model
	ProductID     uint         `json:"product_id" form:"product_id" gorm:"not null;index"`
	Product       Product      `json:"product" gorm:"foreignKey:ProductID"`
	ProductUnitID *uint        `json:"product_unit_id" form:"product_unit_id" gorm:"index"`                        // Which unit was used (nullable for backward compatibility)
	ProductUnit   *ProductUnit `json:"product_unit" gorm:"foreignKey:ProductUnitID"`                               // Unit details
	UnitName      string       `json:"unit_name" form:"unit_name" gorm:"size:50"`                                  // Unit name for reference (denormalized for performance)
	Quantity      int          `json:"quantity" form:"quantity" gorm:"not null"`                                   // Quantity in the specified unit
	UnitPrice     float64      `json:"unit_price" form:"unit_price" gorm:"not null"`                               // Price per unit
	Total         float64      `json:"total" form:"total" gorm:"not null"`                                         // Total amount (quantity * unit_price)
	BaseQuantity  float64      `json:"base_quantity" form:"base_quantity" gorm:"not null"`                         // Equivalent quantity in base units
	Type          string       `json:"type" form:"type" gorm:"not null;size:20;check:type IN ('purchase','sale')"` // "purchase" or "sale"
	Notes         string       `json:"notes" form:"notes" gorm:"size:500"`                                         // Optional transaction notes
}

// CalculateTotal calculates the total amount for this transaction
func (t *Transaction) CalculateTotal() {
	t.Total = float64(t.Quantity) * t.UnitPrice
}

// GetEffectiveUnitName returns the unit name, preferring ProductUnit.UnitName over UnitName field
func (t *Transaction) GetEffectiveUnitName() string {
	if t.ProductUnit != nil {
		return t.ProductUnit.UnitName
	}
	if t.UnitName != "" {
		return t.UnitName
	}
	return "piece" // default fallback
}

// IsValidType checks if the transaction type is valid
func (t *Transaction) IsValidType() bool {
	return t.Type == "purchase" || t.Type == "sale"
}
