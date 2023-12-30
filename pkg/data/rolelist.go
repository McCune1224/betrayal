package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type RoleList struct {
	ID    int            `db:"id" json:"id"`
	Roles pq.StringArray `db:"roles" json:"roles"`
}

type RoleListModel struct {
	DB *sqlx.DB
}

func (r *RoleListModel) Insert(roleList RoleList) error {
	_, err := r.DB.Exec("INSERT INTO active_roles (roles) VALUES ($1)", roleList.Roles)
	return err
}

func (r *RoleListModel) Update(roleList RoleList) error {
	_, err := r.DB.Exec("UPDATE active_roles SET roles = $1 WHERE id = 1", roleList.Roles)
	return err
}

func (r *RoleListModel) Get() (*RoleList, error) {
	var roleList RoleList
	// Really only one row in the table so just select all and grab the first
	err := r.DB.Get(&roleList, "SELECT * FROM active_roles")
	return &roleList, err
}

func (r *RoleListModel) Delete(roleList RoleList) error {
	_, err := r.DB.Exec("DELETE FROM active_roles WHERE id = 1")
	return err
}
