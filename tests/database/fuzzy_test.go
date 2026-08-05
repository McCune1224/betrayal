package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/suite"
)

// FuzzyTestSuite exercises the levenshtein-based fuzzy lookup queries used by
// /search and /list. Runs against the LOCAL database only (see testutil).
type FuzzyTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
	Q  *models.Queries
}

func (f *FuzzyTestSuite) SetupSuite() {
	f.DB = testutil.NewTestPool(f.T())
	f.Q = models.New(f.DB)
}

func (f *FuzzyTestSuite) SetupTest() {
	testutil.TruncateAll(f.T(), f.DB)
	f.seed()
}

// seed inserts the fixture set used by every fuzzy test.
func (f *FuzzyTestSuite) seed() {
	ctx := context.Background()
	abilities := []models.CreateAbilityInfoParams{
		{Name: "Shadow Step", Description: "Teleport in the dark", DefaultCharges: 2, AnyAbility: true, Rarity: models.RarityRARE},
		{Name: "Fireball", Description: "Throw a ball of fire", DefaultCharges: 3, AnyAbility: true, Rarity: models.RarityUNCOMMON},
		{Name: "Hexed Blood", Description: "Role specific curse", DefaultCharges: 1, AnyAbility: false, Rarity: models.RarityROLESPECIFIC},
	}
	for _, a := range abilities {
		_, err := f.Q.CreateAbilityInfo(ctx, a)
		f.Require().NoError(err)
	}

	items := []models.CreateItemParams{
		{Name: "Silver Dagger", Description: "Stabby", Rarity: models.RarityRARE, Cost: 50},
		{Name: "Healing Salve", Description: "Soothing", Rarity: models.RarityCOMMON, Cost: 10},
	}
	for _, i := range items {
		_, err := f.Q.CreateItem(ctx, i)
		f.Require().NoError(err)
	}

	_, err := f.Q.CreateRole(ctx, models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	f.Require().NoError(err)
}

func (f *FuzzyTestSuite) TestGetAbilityInfoByFuzzy() {
	ability, err := f.Q.GetAbilityInfoByFuzzy(context.Background(), "shadow")
	f.Require().NoError(err)
	f.Equal("Shadow Step", ability.Name)
}

func (f *FuzzyTestSuite) TestGetAnyAbilityByFuzzy() {
	// Only any_ability=true rows are candidates.
	ability, err := f.Q.GetAnyAbilityByFuzzy(context.Background(), "hexed")
	f.Require().NoError(err)
	f.NotEqual("Hexed Blood", ability.Name) // role-specific is excluded
}

func (f *FuzzyTestSuite) TestSearchAbilityByKeyword() {
	// NOTE: the query is a bare LIKE (no wildcards) — effectively
	// case-insensitive exact match. Documenting current behavior; fixing the
	// search UX is WT-5 territory.
	results, err := f.Q.SearchAbilityByKeyword(context.Background(), "Fireball")
	f.Require().NoError(err)
	f.Len(results, 1)
	f.Equal("Fireball", results[0].Name)

	results, err = f.Q.SearchAbilityByKeyword(context.Background(), "fire")
	f.Require().NoError(err)
	f.Empty(results)
}

func (f *FuzzyTestSuite) TestGetItemByFuzzy() {
	item, err := f.Q.GetItemByFuzzy(context.Background(), "dagger")
	f.Require().NoError(err)
	f.Equal("Silver Dagger", item.Name)
}

func (f *FuzzyTestSuite) TestSearchRoleByName() {
	roles, err := f.Q.SearchRoleByName(context.Background(), "mafia")
	f.Require().NoError(err)
	f.Require().Len(roles, 1)
	f.Equal("Mafia", roles[0].Name)
}

func TestFuzzySuite(t *testing.T) {
	suite.Run(t, new(FuzzyTestSuite))
}
