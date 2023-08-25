package data

import (
	"gorm.io/gorm"
)

// General representation of a role in the game in db.
type Role struct {
	gorm.Model
	Name        string    `gorm:"unique;not null"`
	Description string    `gorm:"not null"`
	Alignment   string    `gorm:"not null"`
	IsActive    bool      `gorm:"not null;default:true"`
	Abilities   []Ability `gorm:"many2many:role_abilities;"`
	Perks       []Perk    `gorm:"many2many:role_perks;"`
}

type RoleModel struct {
	DB *gorm.DB
}

func (rm *RoleModel) GetByName(name string) (*Role, error) {
	role := &Role{}
	err := rm.DB.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (rm *RoleModel) GetAvailableRoles() ([]*Role, error) {
	var roles []*Role
	err := rm.DB.Where("is_active = ?", true).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
