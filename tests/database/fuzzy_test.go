package database

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/stretchr/testify/suite"
)

type FuzzyTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
	Q  *models.Queries
}

func (f *FuzzyTestSuite) SetupTest() {
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")
	pools, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_POOLER_URL"))
	if err != nil {
		f.FailNow(err.Error())
	}
	f.DB = pools
	f.Q = models.New(f.DB)
}

func TestFuzzySuite(t *testing.T) {
	suite.Run(t, new(FuzzyTestSuite))
}
