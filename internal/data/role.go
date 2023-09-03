package data

import "github.com/jmoiron/sqlx"

// General representation of a role in the game in db.
type Role struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Alignment   string `db:"alignment"`
	CreatedAt   string `db:"created_at"`
}

// Entire role with all abilities and perks.
type RoleComplete struct {
	Role      Role
	Abilities []Ability
	Perks     []Perk
}

// Join table for roles and abilities 'role_abilities'
type RoleAbility struct {
	RoleID     int64  `db:"role_id"`
	AbilityID  int64  `db:"ability_id"`
	CREATED_AT string `db:"created_at"`
}

// Join table for roles and perks 'roles_perks'
type RolePerk struct {
	RoleID int64 `db:"role_id"`
	PerkID int64 `db:"perk_id"`
}

type RoleModel struct {
	DB *sqlx.DB
}

func (rm *RoleModel) Insert(r *Role) (int64, error) {
	query := `INSERT INTO roles (name, description, alignment) VALUES ($1, $2, $3)`
	_, err := rm.DB.Exec(query, r.Name, r.Description, r.Alignment)
	if err != nil {
		return -1, err
	}
	var lastInsert Role
	err = rm.DB.Get(&lastInsert, "SELECT * FROM roles ORDER BY id DESC LIMIT 1")

	return lastInsert.ID, nil
}

func (rm *RoleModel) Get(id int64) (*Role, error) {
	var r Role
	err := rm.DB.Get(&r, "SELECT * FROM roles WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (rm *RoleModel) GetByName(name string) (*Role, error) {
	var r Role
	err := rm.DB.Get(&r, "SELECT * FROM roles WHERE name ILIKE $1", name)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (rm *RoleModel) Update(r *Role) error {
	query := `UPDATE roles SET name = $1, description = $2, alignment = $3 WHERE id = $4`
	_, err := rm.DB.Exec(query, r.Name, r.Description, r.Alignment, r.ID)
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) Delete(id int64) error {
	_, err := rm.DB.Exec("DELETE FROM roles WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) WipeTable() error {
	_, err := rm.DB.Exec("DELETE FROM roles")
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) GetAll() ([]Role, error) {
	var roles []Role
	err := rm.DB.Select(&roles, "SELECT * FROM roles")
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (rm *RoleModel) JoinAbility(roleID int64, abilityID int64) error {
	query := `INSERT INTO roles_abilities (role_id, ability_id) VALUES ($1, $2)`
	_, err := rm.DB.Exec(query, roleID, abilityID)
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) JoinPerk(roleID int64, perkID int64) error {
	query := `INSERT INTO roles_perks (role_id, perk_id) VALUES ($1, $2)`
	_, err := rm.DB.Exec(query, roleID, perkID)
	if err != nil {
		return err
	}
	return nil
}
