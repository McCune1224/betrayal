package data

import (
	"gorm.io/gorm"
)

// All interested models to be used in the application
type Models struct {
	Roles    RoleModel
	Insults  InsultModel
	Abilitys AbilityModel
}

// NewModels creates a new instance of the Models struct and attaches the database connection to it.
func NewModels(db *gorm.DB, auto ...bool) Models {

	ModelHandlers := Models{
		Roles:    RoleModel{DB: db},
		Insults:  InsultModel{DB: db},
		Abilitys: AbilityModel{DB: db},
	}

	if len(auto) > 0 && auto[0] {
		// db.Migrator().DropTable(
		// 	&Role{},
		// 	&Insult{},
		// )
		// db.AutoMigrate(
		// 	Role{},
		// 	Insult{},
		// 	Ability{},
		// )
	}
	return ModelHandlers
}
