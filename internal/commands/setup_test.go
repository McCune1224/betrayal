package commands

import (
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/mccune1224/betrayal/internal/data"
)

func DbInit() *sqlx.DB {
	godotenv.Load("../../.env")
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TestGenerateRolePools(t *testing.T) {
	db := DbInit()
	listDB := data.RoleListModel{DB: db}
	roleDB := data.RoleModel{DB: db}
	activeRoles, err := listDB.Get()
	if err != nil {
		t.Fatal(err)
	}

	ar := activeRoles.Roles
	activeRolePool, err := roleDB.GetBulkByName(ar)
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range activeRolePool {
		t.Log(role.Name)
	}
}
