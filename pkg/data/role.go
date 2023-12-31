package data

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/util"
)

// How roles are stored in the database.
type Role struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	Alignment   string `db:"alignment" json:"alignment"`
	CreatedAt   string `db:"created_at" json:"created_at"`
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
	// Make it find closest match if no exact match
	query := `SELECT * FROM roles WHERE name ILIKE '%' || $1 || '%'`
	err := rm.DB.Get(&r, query, name)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (rm *RoleModel) GetByFuzzy(name string) (*Role, error) {
	var r Role
	rm.DB.Get(&r, "SELECT * from roles WHERE name ILIKE $1", name)
	if r.Name == name {
		return &r, nil
	}
	var rChocies []Role
	if len(name) < 2 {
		return nil, errors.New("Name must be at least 2 characters")
	}
	err := rm.DB.Select(&rChocies, "SELECT * FROM roles")
	if err != nil {
		return nil, err
	}
	strChoices := make([]string, len(rChocies))
	for i, role := range rChocies {
		strChoices[i] = role.Name
	}
	best, _ := util.FuzzyFind(name, strChoices)
	for _, role := range rChocies {
		if role.Name == best {
			r = role
		}
	}
	return &r, nil
}

func (rm *RoleModel) GetBulkByName(names pq.StringArray) ([]*Role, error) {
	var roles []*Role
	query := `SELECT * FROM roles WHERE name = ANY($1)`
	err := rm.DB.Select(&roles, query, names)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (rm *RoleModel) GetAllByID(ids []int64) ([]*Role, error) {
	var roles []*Role
	err := rm.DB.Select(&roles, "SELECT * FROM roles WHERE id = ANY($1)", ids)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (rm *RoleModel) GetAllNames() ([]string, error) {
	var names []string
	err := rm.DB.Select(&names, "SELECT name FROM roles")
	if err != nil {
		return nil, err
	}
	return names, nil
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

func (rm *RoleModel) GetAllByAbilityID(abilityID int64) ([]Role, error) {
	var roles []Role
	err := rm.DB.Select(
		&roles,
		`SELECT (roles.*) FROM roles INNER JOIN roles_abilities ON roles.id = roles_abilities.role_id WHERE roles_abilities.ability_id = $1`,
		abilityID,
	)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (rm *RoleModel) GetByPerkID(perkID int64) (*Role, error) {
	var r Role
	err := rm.DB.Get(
		&r,
		`SELECT (roles.*) FROM roles INNER JOIN roles_perks ON roles.id = roles_perks.role_id WHERE roles_perks.perk_id = $1`,
		perkID,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (rm *RoleModel) GetAllByPerkID(p *Perk) ([]Role, error) {
	var roles []Role
	err := rm.DB.Select(
		&roles,
		// Select all roles that have the provided perk and search by perk name
		`SELECT (roles.*) FROM roles INNER JOIN roles_perks ON roles.id = roles_perks.role_id INNER JOIN perks ON perks.id = roles_perks.perk_id WHERE perks.name ILIKE '%' || $1 || '%'`,
		// `SELECT (roles.*) FROM roles INNER JOIN roles_perks ON roles.id = roles_perks.role_id INNER JOIN perks ON perks.id = roles_perks.perk_id WHERE perks.name = $1`,
		p.Name,
	)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
