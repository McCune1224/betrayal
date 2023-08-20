package data

import (
	"github.com/mccune1224/betrayal/internal/data/enum"
	"gorm.io/gorm"
)

type Perk struct {
	gorm.Model
	Name          string `gorm:"unique;not null"`
	Categories    enum.PerkCategoryEnum
	Effect        string `gorm:"not null"`
	OrderPriority int    `gorm:"not null;default:0"`
}

type PerkAttachment struct {
	gorm.Model
	Abilities Ability `gorm:"foreignKey:AbilityID"`
	AbilityID uint    `gorm:"not null"`
	Roles     []Role  `gorm:"many2many:perk_attachment_roles;"`
}

type PerkModel struct {
	Db *gorm.DB
}
