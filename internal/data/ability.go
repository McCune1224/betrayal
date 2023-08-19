package data

import (
	"github.com/mccune1224/betrayal/internal/data/enum"
	"gorm.io/gorm"
)

// Game representation of an AbilityModel.
type AbilityModel struct {
	gorm.Model
	Name           string                    `gorm:"unique;not null"`
	ActionType     enum.ActionTypeEnum       `gorm:"not null"`
	Categories     []enum.ActionCategoryEnum `gorm:"not null"`
	Charges        int                       `gorm:"not null"`
	IsAnyAbility   bool                      `gorm:"not null;default:false"`
	Rarity         enum.RarityEnum           `gorm:"default:COMMON"`
	Effect         string                    `gorm:"not null"`
	ShowCategories bool                      `gorm:"not null;default:true"`
	detailedEffect string
	OrderPriority  int
}

type AbilityChangeModel struct {
	gorm.Model
}
