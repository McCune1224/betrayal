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

// rollRarityOrder mirrors the `rarity` enum order from migration 000001.
var rollRarityOrder = []models.Rarity{
	models.RarityCOMMON,
	models.RarityUNCOMMON,
	models.RarityRARE,
	models.RarityEPIC,
	models.RarityLEGENDARY,
	models.RarityMYTHICAL,
	models.RarityROLESPECIFIC,
	models.RarityUNIQUE,
}

func rarityIndex(r models.Rarity) int {
	for i, rr := range rollRarityOrder {
		if rr == r {
			return i
		}
	}
	return -1
}

type RollTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
	Q  *models.Queries
}

func (s *RollTestSuite) SetupTest() {
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")

	pools, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	s.Require().NoError(err, "failed to connect to local test database (DATABASE_URL)")
	s.DB = pools
	s.Q = models.New(pools)
}

func (s *RollTestSuite) TearDownTest() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func (s *RollTestSuite) seedAnyAbility(name string, rarity models.Rarity) models.AbilityInfo {
	ability, err := s.Q.CreateAbilityInfo(context.Background(), models.CreateAbilityInfoParams{
		Name:           name,
		Description:    "roll test ability",
		DefaultCharges: 1,
		AnyAbility:     true,
		Rarity:         rarity,
	})
	s.Require().NoError(err)
	return ability
}

// TestGetRandomAnyAbilityByRarity exercises the exact-rarity roll. The query
// previously used `==` (invalid Postgres) so any invocation failed at runtime.
func (s *RollTestSuite) TestGetRandomAnyAbilityByRarity() {
	ctx := context.Background()
	created := s.seedAnyAbility("roll-test-exact-rare", models.RarityRARE)
	defer s.Q.DeleteAbilityInfo(ctx, created.ID)

	ability, err := s.Q.GetRandomAnyAbilityByRarity(ctx, models.RarityRARE)
	s.Require().NoError(err, "exact-rarity roll query must not error")
	s.Equal(models.RarityRARE, ability.Rarity)
	s.True(ability.AnyAbility)
}

// TestGetRandomAnyAbilityByMinimumRarity verifies the minimum-rarity roll only
// returns any-abilities at or above the requested rarity, excludes
// ROLE_SPECIFIC and UNIQUE (parity with the item roll), and is actually random.
func (s *RollTestSuite) TestGetRandomAnyAbilityByMinimumRarity() {
	ctx := context.Background()
	created := []models.AbilityInfo{}
	for i := 0; i < 3; i++ {
		created = append(created, s.seedAnyAbility("roll-test-min-rare", models.RarityRARE))
	}
	created = append(created, s.seedAnyAbility("roll-test-min-mythical", models.RarityMYTHICAL))
	// Must never be returned by the roll:
	created = append(created, s.seedAnyAbility("roll-test-min-common", models.RarityCOMMON))
	created = append(created, s.seedAnyAbility("roll-test-min-role-specific", models.RarityROLESPECIFIC))
	created = append(created, s.seedAnyAbility("roll-test-min-unique", models.RarityUNIQUE))
	defer func() {
		for _, ability := range created {
			s.Q.DeleteAbilityInfo(ctx, ability.ID)
		}
	}()

	minRarityIdx := rarityIndex(models.RarityRARE)
	seen := map[int32]bool{}
	for i := 0; i < 5; i++ {
		ability, err := s.Q.GetRandomAnyAbilityByMinimumRarity(ctx, models.RarityRARE)
		s.Require().NoError(err)
		s.True(ability.AnyAbility, "minimum-rarity roll returned a non-any-ability")
		s.GreaterOrEqual(rarityIndex(ability.Rarity), minRarityIdx, "roll returned a rarity below the minimum")
		s.NotEqual(models.RarityROLESPECIFIC, ability.Rarity)
		s.NotEqual(models.RarityUNIQUE, ability.Rarity)
		seen[ability.ID] = true
	}
	s.GreaterOrEqual(len(seen), 2, "minimum-rarity roll should return random results, not a fixed row")
}

func TestRollSuite(t *testing.T) {
	suite.Run(t, new(RollTestSuite))
}
