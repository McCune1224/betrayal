package data

import (
	"github.com/mccune1224/betrayal/internal/data/enum"
	"gorm.io/gorm"
)

const (
	// Unlimited charges on certain abilities
	unlimited = -1
)

// Game representation of an AbilityModel.
type Ability struct {
	gorm.Model
	Name       string                    `gorm:"unique;not null"`
	ActionType enum.ActionTypeEnum       `gorm:"not null"`
	Categories []enum.ActionCategoryEnum `gorm:"not null"`
	// -1 is unlimited
	Charges        int             `gorm:"not null"`
	IsAnyAbility   bool            `gorm:"not null;default:false"`
	Rarity         enum.RarityEnum `gorm:"default:COMMON"`
	Effect         string          `gorm:"not null"`
	ShowCategories bool            `gorm:"not null;default:true"`
	detailedEffect string
	OrderPriority  int
}

type AbilityChange struct {
	gorm.Model
	Ability   Ability `gorm:"foreignKey:AbilityID"`
	AbilityID uint    `gorm:"not null"`
	Change    string  `gorm:"not null"`
}

type AbilityAttachment struct {
	gorm.Model
	Ability   Ability `gorm:"foreignKey:AbilityID"`
	AbilityID uint    `gorm:"not null"`
	Roles     []Role  `gorm:"many2many:ability_attachment_roles;"`
}

type AbilityModel struct {
	Db *gorm.DB
}
