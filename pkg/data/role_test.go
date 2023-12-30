package data

import (
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func TestGetAllbyPerkID(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println(err)
	}
	dbURL := os.Getenv("DATABASE_URL")
	pSQL, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	tests := []Perk{
		{
			Name: "Legion",
		},
		{
			Name: "Heat Vision",
		},
		{
			Name: "Rigged Luck",
		},
	}
	roleDB := RoleModel{DB: pSQL}
	for _, testPerk := range tests {
		roles, err := roleDB.GetAllByPerkID(&testPerk)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("--- Perk RESULTS : %s ---", testPerk.Name)
		for _, role := range roles {
			t.Logf("Role: %s", role.Name)
		}
	}
}
