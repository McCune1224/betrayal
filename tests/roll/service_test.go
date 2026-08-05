package roll

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	rollsvc "github.com/mccune1224/betrayal/internal/services/roll"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/suite"
)

// RollServiceSuite exercises the DB-backed roll service draws against the
// LOCAL database (see testutil for the production guard).
type RollServiceSuite struct {
	suite.Suite
	DB *pgxpool.Pool
	Q  *models.Queries
}

func (s *RollServiceSuite) SetupSuite() {
	s.DB = testutil.NewTestPool(s.T())
	s.Q = models.New(s.DB)
}

func (s *RollServiceSuite) SetupTest() {
	testutil.TruncateAll(s.T(), s.DB)
	ctx := context.Background()

	// One item per rarity so draws are unambiguous.
	for _, r := range []models.Rarity{models.RarityCOMMON, models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY} {
		_, err := s.Q.CreateItem(ctx, models.CreateItemParams{
			Name: "Item " + string(r), Description: "item", Rarity: r, Cost: 10,
		})
		s.Require().NoError(err)
	}

	// One any-ability (RARE) plus one role-specific ability tied to a role.
	_, err := s.Q.CreateAbilityInfo(ctx, models.CreateAbilityInfoParams{
		Name: "Any Power", Description: "any ability", DefaultCharges: 1,
		AnyAbility: true, Rarity: models.RarityRARE,
	})
	s.Require().NoError(err)

	role, err := s.Q.CreateRole(ctx, models.CreateRoleParams{
		Name: "Mafia", Description: "boss", Alignment: models.AlignmentEVIL,
	})
	s.Require().NoError(err)
	roleSpecific, err := s.Q.CreateAbilityInfo(ctx, models.CreateAbilityInfoParams{
		Name: "Mafia Hit", Description: "role specific", DefaultCharges: 1,
		AnyAbility: true, Rarity: models.RarityRARE,
	})
	s.Require().NoError(err)
	err = s.Q.CreateRoleAbilityJoin(ctx, models.CreateRoleAbilityJoinParams{
		RoleID: role.ID, AbilityID: roleSpecific.ID,
	})
	s.Require().NoError(err)
}

func (s *RollServiceSuite) TestRollItemByRarity() {
	ctx := context.Background()
	svc := rollsvc.New(s.DB)

	for _, r := range []models.Rarity{models.RarityCOMMON, models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY} {
		item, err := svc.RollItemByRarity(ctx, r)
		s.Require().NoError(err)
		s.Equal(r, item.Rarity)
	}
}

func (s *RollServiceSuite) TestRollItemAtMinimumRarity() {
	ctx := context.Background()
	svc := rollsvc.New(s.DB)

	item, err := svc.RollItemAtMinimumRarity(ctx, models.RarityRARE)
	s.Require().NoError(err)
	s.Contains([]models.Rarity{models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY}, item.Rarity)

	// Minimum rarity EXCLUSIVE behavior: min COMMON must not return UNIQUE.
	item, err = svc.RollItemAtMinimumRarity(ctx, models.RarityCOMMON)
	s.Require().NoError(err)
	s.NotEqual(models.RarityUNIQUE, item.Rarity)
}

func (s *RollServiceSuite) TestRollAnyAbilityIncludingRoleSpecific() {
	ctx := context.Background()
	svc := rollsvc.New(s.DB)

	// NOTE: the query INNER JOINs role_ability, so only role-tied
	// any-abilities are ever eligible — a latent quirk (WT-5 candidate).
	// "Mafia Hit" is the only any-ability joined to a role in the fixtures.
	ability, err := svc.RollAnyAbilityIncludingRoleSpecific(ctx, models.RarityRARE, 0)
	s.Require().NoError(err)
	s.Equal("Mafia Hit", ability.Name)

	role, err := s.Q.GetRoleByFuzzy(ctx, "Mafia")
	s.Require().NoError(err)
	ability, err = svc.RollAnyAbilityIncludingRoleSpecific(ctx, models.RarityRARE, role.ID)
	s.Require().NoError(err)
	s.Equal("Mafia Hit", ability.Name)
}

// TestRollAnyAbilityByRarityPinsKnownBug documents the WT-5 B2 regression:
// the SQL uses `rarity == $1` (invalid in Postgres), so the query always
// errors. When WT-5 fixes the query, this test must be updated to assert the
// returned ability instead of the error.
func (s *RollServiceSuite) TestRollAnyAbilityByRarityPinsKnownBug() {
	ctx := context.Background()
	svc := rollsvc.New(s.DB)

	_, err := svc.RollAnyAbilityByRarity(ctx, models.RarityRARE)
	s.Require().Error(err)
}

func TestRollServiceSuite(t *testing.T) {
	suite.Run(t, new(RollServiceSuite))
}
