package data

import (
	"github.com/mccune1224/betrayal/internal/data/enum"
	"gorm.io/gorm"
)

// General representation of a role in the game in db.
type Role struct {
	gorm.Model
	Name      string                  `gorm:"unique;not null"`
	Alignment enum.ActionCategoryEnum `gorm:"not null"`
	IsActive  bool                    `gorm:"not null;default:true"`
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
