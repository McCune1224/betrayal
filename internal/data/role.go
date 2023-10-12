package data

import (
	"github.com/jmoiron/sqlx"
)

// How roles are stored in the database.
type Role struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Alignment   string `db:"alignment"`
	CreatedAt   string `db:"created_at"`
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

func (rm *RoleModel) GetAll() ([]*Role, error) {
	var roles []*Role
	err := rm.DB.Select(&roles, "SELECT * FROM roles")
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (rm *RoleModel) InsertJoinAbility(roleID int64, abilityID int64) error {
	query := `INSERT INTO roles_abilities (role_id, ability_id) VALUES ($1, $2)`
	_, err := rm.DB.Exec(query, roleID, abilityID)
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) InsertJoinPerk(roleID int64, perkID int64) error {
	query := `INSERT INTO roles_perks (role_id, perk_id) VALUES ($1, $2)`
	_, err := rm.DB.Exec(query, roleID, perkID)
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) GetAbilities(roleID int64) ([]*Ability, error) {
	query := `SELECT (abilities.*) FROM abilities INNER JOIN roles_abilities ON abilities.id = roles_abilities.ability_id WHERE roles_abilities.role_id = $1`
	var abilities []*Ability
	err := rm.DB.Select(
		&abilities,
		query,
		roleID,
	)
	if err != nil {
		return nil, err
	}
	return abilities, nil
}

func (rm *RoleModel) GetPerks(RoleID int64) ([]*Perk, error) {
	query := `SELECT (perks.*) FROM perks INNER JOIN roles_perks ON perks.id = roles_perks.perk_id WHERE roles_perks.role_id = $1`
	var perks []*Perk
	err := rm.DB.Select(
		&perks,
		query,
		RoleID,
	)
	if err != nil {
		return nil, err
	}
	return perks, nil
}

func (rm *RoleModel) Upsert(r *Role) error {
	query := `INSERT INTO roles (name, description, alignment)
    VALUES ($1, $2, $3)
    ON CONFLICT (name) DO UPDATE
    SET name = $1, description = $2, alignment = $3`
	_, err := rm.DB.Exec(
		query,
		r.Name,
		r.Description,
		r.Alignment,
	)
	if err != nil {
		return err
	}
	return nil
}

func (rm *RoleModel) GetByAbilityID(abilityID int64) (*Role, error) {
	var r Role
	err := rm.DB.Get(
		&r,
		`SELECT (roles.*) FROM roles INNER JOIN roles_abilities ON roles.id = roles_abilities.role_id WHERE roles_abilities.ability_id = $1`,
		abilityID,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
