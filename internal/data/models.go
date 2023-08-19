package data

import (
	"gorm.io/gorm"
)

// All interested models to be used in the application
type Models struct {
	Roles RoleModel
}

// NewModels creates a new instance of the Models struct and attaches the database connection to it.
func NewModels(db *gorm.DB, auto ...bool) Models {

	ModelHandlers := Models{
		Roles: RoleModel{DB: db},
	}

	if len(auto) > 0 && auto[0] {
		db.AutoMigrate(Role{})
	}
	return ModelHandlers
}
