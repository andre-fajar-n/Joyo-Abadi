package models

import "gorm.io/gorm"

// ProductUnit represents different units for a product (e.g., box, dozen, bottle)
type ProductUnit struct {
	gorm.Model
	ProductID      uint    `json:"product_id" form:"product_id" gorm:"not null;index"`
	Product        Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	UnitName       string  `json:"unit_name" form:"unit_name" gorm:"not null;size:50"`               // e.g., "box", "dozen", "bottle"
	ConversionRate float64 `json:"conversion_rate" form:"conversion_rate" gorm:"not null;default:1"` // How many base units this represents
	Price          float64 `json:"price" form:"price" gorm:"not null"`                               // Price for this unit
	Quantity       int     `json:"quantity" form:"quantity" gorm:"not null;default:0"`               // Stock quantity for this unit
	IsBaseUnit     bool    `json:"is_base_unit" form:"is_base_unit" gorm:"default:false"`            // Whether this is the base unit
	IsActive       bool    `json:"is_active" form:"is_active" gorm:"default:true"`                   // Whether this unit is currently available
	Description    string  `json:"description" form:"description" gorm:"size:255"`                   // Optional description for the unit
}

// TableName sets the table name for ProductUnit
func (ProductUnit) TableName() string {
	return "product_units"
}

// GetBaseQuantity calculates the quantity in base units
func (pu *ProductUnit) GetBaseQuantity() float64 {
	return float64(pu.Quantity) * pu.ConversionRate
}

// GetPricePerBaseUnit calculates the price per base unit
func (pu *ProductUnit) GetPricePerBaseUnit() float64 {
	if pu.ConversionRate == 0 {
		return pu.Price
	}
	return pu.Price / pu.ConversionRate
}

// CanFulfillQuantity checks if this unit can fulfill the requested quantity
func (pu *ProductUnit) CanFulfillQuantity(requestedQty int) bool {
	return pu.IsActive && pu.Quantity >= requestedQty
}

// ConvertToBaseUnits converts a quantity of this unit to base units
func (pu *ProductUnit) ConvertToBaseUnits(quantity int) float64 {
	return float64(quantity) * pu.ConversionRate
}

// ConvertFromBaseUnits converts a base unit quantity to this unit (rounded down)
func (pu *ProductUnit) ConvertFromBaseUnits(baseQuantity float64) int {
	if pu.ConversionRate == 0 {
		return 0
	}
	return int(baseQuantity / pu.ConversionRate)
}
