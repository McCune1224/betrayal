package data

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrRecordAlreadyExists = errors.New("record already exists")
)

// All interested models to be used in the application
type Models struct {
	Roles     RoleModel
	Insults   InsultModel
	Abilities AbilityModel
	Statuses  StatusModel
	Items     ItemModel
}

// NewModels creates a new instance of the Models struct and attaches the database connection to it.
func NewModels(db *sqlx.DB) Models {

	ModelHandlers := Models{
		Roles:     RoleModel{DB: db},
		Insults:   InsultModel{DB: db},
		Abilities: AbilityModel{DB: db},
		Statuses:  StatusModel{DB: db},
		Items:     ItemModel{DB: db},
	}
	return ModelHandlers
}
