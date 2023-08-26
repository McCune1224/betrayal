package data

import (
	"gorm.io/gorm"
)

type Perk struct {
	gorm.Model
	Name          string
	Categories    string
	Effect        string
	OrderPriority int
}

type PerkAttachment struct {
	gorm.Model
	Abilities Ability
	AbilityID uint
	Roles     []Role
}

type PerkModel struct {
	DB *gorm.DB
}
