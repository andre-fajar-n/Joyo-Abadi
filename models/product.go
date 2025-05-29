package models

import "gorm.io/gorm"

// Product represents a product in the system
type Product struct {
	gorm.Model
	Name         string        `json:"name" form:"name" gorm:"not null" validate:"required,min=1,max=100"`
	BaseUnitName string        `json:"base_unit_name" form:"base_unit_name" gorm:"not null;default:'piece'" validate:"required,min=1,max=50,safe_text"` // e.g., "piece", "gram", "liter"
	Price        float64       `json:"price" form:"price" gorm:"not null" validate:"gt=0"`                                                              // Price for base unit (deprecated, use ProductUnits)
	Quantity     int           `json:"quantity" form:"quantity" gorm:"not null" validate:"gte=0"`                                                       // Base unit quantity (deprecated, use ProductUnits)
	UserID       uint          `json:"user_id" gorm:"nullable"`
	User         *User         `json:"user" gorm:"foreignKey:UserID"`
	Description  string        `json:"description" form:"description" validate:"max=500"`
	Units        []ProductUnit `json:"units" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"` // Available units for this product
	IsActive     bool          `json:"is_active" form:"is_active" gorm:"default:true"`                // Whether product is active
}

// GetBaseUnit returns the base unit for this product
func (p *Product) GetBaseUnit() *ProductUnit {
	for _, unit := range p.Units {
		if unit.IsBaseUnit {
			return &unit
		}
	}
	return nil
}

// GetActiveUnits returns all active units for this product
func (p *Product) GetActiveUnits() []ProductUnit {
	var activeUnits []ProductUnit
	for _, unit := range p.Units {
		if unit.IsActive {
			activeUnits = append(activeUnits, unit)
		}
	}
	return activeUnits
}

// GetUnitByName returns a unit by its name
func (p *Product) GetUnitByName(unitName string) *ProductUnit {
	for _, unit := range p.Units {
		if unit.UnitName == unitName && unit.IsActive {
			return &unit
		}
	}
	return nil
}

// GetTotalBaseQuantity calculates total quantity in base units across all units
func (p *Product) GetTotalBaseQuantity() float64 {
	total := 0.0
	for _, unit := range p.Units {
		if unit.IsActive {
			total += unit.GetBaseQuantity()
		}
	}
	return total
}

// HasStock checks if the product has any stock in any unit
func (p *Product) HasStock() bool {
	for _, unit := range p.Units {
		if unit.IsActive && unit.Quantity > 0 {
			return true
		}
	}
	return false
}
