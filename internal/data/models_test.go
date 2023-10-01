package data

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func TestGetSQLKeys(t *testing.T) {
	exampleModel := []struct {
		ID    int64          `db:"id"`
		Name  string         `db:"name"`
		Age   int64          `db:"age"`
		Books pq.StringArray `db:"books"`
	}{
		{
			ID:    1,
			Name:  "John",
			Age:   20,
			Books: pq.StringArray{"Book1", "Book2"},
		},
		{
			ID:    2,
			Name:  "Jane",
			Age:   30,
			Books: pq.StringArray{"Le Epic Book", "Book Who cares"},
		},
	}

	for _, v := range exampleModel {
		keys := SqlGenKeys(v)
		t.Log(keys)
	}
}

func TestUpdate(t *testing.T) {
	err := godotenv.Load("../../.env")
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		t.Fatal("DATABASE_URL must be set for tests")
	}
	pSQL, err := sqlx.Connect("postgres", db)
	if err != nil {
		t.Fatal(err)
	}

	invDB := InventoryModel{DB: pSQL}
	inv, err := invDB.GetByDiscordID("206268866714796032")
	if err != nil {
		t.Fatal(err)
	}
	inv.Notes = append(inv.Notes, "TestUpdate")
	err = invDB.Update(inv)
	if err != nil {
		t.Fatal(err)
	}

}
