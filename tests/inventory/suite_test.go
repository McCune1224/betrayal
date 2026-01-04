package inventory

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type InventoryTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
}

func (its *InventoryTestSuite) SetupTest() {
	godotenv.Load(".env")
	godotenv.Load("../.env")

	if !postgresSocketAvailable() {
		its.T().Skip("skipping inventory tests: postgres socket not available at /tmp/.s.PGSQL.5432")
	}

	pools, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		its.FailNow(err.Error())
	}
	its.DB = pools
}

func TestInventorySuite(t *testing.T) {
	suite.Run(t, new(InventoryTestSuite))
}

func postgresSocketAvailable() bool {
	if _, err := os.Stat("/tmp/.s.PGSQL.5432"); err != nil {
		return false
	}
	return true
}
