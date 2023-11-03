package commands

import (
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/mccune1224/betrayal/internal/data"
	"golang.org/x/exp/slices"
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

func TestGenerateRoleSelectPool(t *testing.T) {
	db := DbInit()
	listDB := data.RoleListModel{DB: db}
	roleDB := data.RoleModel{DB: db}
	activeRoles, err := listDB.Get()
	if err != nil {
		t.Fatal(err)
	}

	ar := activeRoles.Roles
	empressIndex := slices.Index(ar, "Empress")
	if empressIndex != -1 {
		ar = append(ar[:empressIndex], ar[empressIndex+1:]...)
	}
	activeRolePool, err := roleDB.GetBulkByName(ar)
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range activeRolePool {
		t.Log(role.Name)
	}
}
