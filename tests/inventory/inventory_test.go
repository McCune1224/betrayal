package inventory

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/suite"
)

// InventoryServiceSuite tests the inventory service (internal/services/inventory)
// against the LOCAL database. The old suite's /tmp/.s.PGSQL.5432 socket skip is
// gone: testutil.NewTestPool fails loudly if DATABASE_URL is unreachable, and
// every test starts from a truncated schema via testutil.TruncateAll.
type InventoryServiceSuite struct {
	suite.Suite
	DB     *pgxpool.Pool
	Q      *models.Queries
	player models.Player
}

func (s *InventoryServiceSuite) SetupSuite() {
	s.DB = testutil.NewTestPool(s.T())
	s.Q = models.New(s.DB)
}

// SetupTest truncates all tables and seeds one role, one player, one
// any-ability and one item — the fixture set the service tests mutate.
func (s *InventoryServiceSuite) SetupTest() {
	testutil.TruncateAll(s.T(), s.DB)
	ctx := context.Background()

	role, err := s.Q.CreateRole(ctx, models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	s.Require().NoError(err)

	player, err := s.Q.CreatePlayer(ctx, models.CreatePlayerParams{
		ID:        100000000000000001,
		RoleID:    pgtype.Int4{Int32: role.ID, Valid: true},
		Alive:     true,
		Coins:     200,
		CoinBonus: pgtype.Numeric{}, // NULL -> column DEFAULT 0
		Luck:      0,
		Alignment: models.AlignmentEVIL,
	})
	s.Require().NoError(err)
	s.player = player

	_, err = s.Q.CreateAbilityInfo(ctx, models.CreateAbilityInfoParams{
		Name: "Shadow Step", Description: "Teleport in the dark",
		DefaultCharges: 2, AnyAbility: true, Rarity: models.RarityRARE,
	})
	s.Require().NoError(err)

	_, err = s.Q.CreateItem(ctx, models.CreateItemParams{
		Name: "Silver Dagger", Description: "Stabby", Rarity: models.RarityRARE, Cost: 50,
	})
	s.Require().NoError(err)
}

func (s *InventoryServiceSuite) handler() *inventory.InventoryHandler {
	// Jank() builds a handler without a ken context (documented test path).
	return inventory.Jank(s.player, s.DB)
}

func (s *InventoryServiceSuite) TestAddAbility() {
	ctx := context.Background()
	handler := s.handler()

	added, err := handler.AddAbility("Shadow Step", 0)
	s.Require().NoError(err)
	s.Equal("Shadow Step", added.Name)
	s.Equal(int32(2), added.DefaultCharges)

	// Duplicate add is rejected.
	_, err = handler.AddAbility("Shadow Step", 1)
	s.Require().Error(err)
	s.Equal("ability already added", err.Error())

	// Persisted via the join table (default charges when quantity is 0).
	rows, err := s.Q.ListPlayerAbilityJoin(ctx, s.player.ID)
	s.Require().NoError(err)
	s.Require().Len(rows, 1)
	s.Equal(added.ID, rows[0].AbilityID)
	s.Equal(int32(2), rows[0].Quantity)
}

func (s *InventoryServiceSuite) TestUpdateAbility() {
	handler := s.handler()
	added, err := handler.AddAbility("Shadow Step", 1)
	s.Require().NoError(err)

	updated, err := handler.UpdateAbility("Shadow Step", 5)
	s.Require().NoError(err)
	s.Equal(added.ID, updated.ID)

	rows, err := s.Q.ListPlayerAbilityJoin(context.Background(), s.player.ID)
	s.Require().NoError(err)
	s.Require().Len(rows, 1)
	s.Equal(int32(5), rows[0].Quantity)
}

func (s *InventoryServiceSuite) TestRemoveAbility() {
	handler := s.handler()
	_, err := handler.AddAbility("Shadow Step", 1)
	s.Require().NoError(err)

	removed, err := handler.RemoveAbility("Shadow Step")
	s.Require().NoError(err)
	s.Equal("Shadow Step", removed.Name)

	rows, err := s.Q.ListPlayerAbilityJoin(context.Background(), s.player.ID)
	s.Require().NoError(err)
	s.Empty(rows)
}

func (s *InventoryServiceSuite) TestAddItemUpsertsQuantity() {
	ctx := context.Background()
	handler := s.handler()

	_, err := handler.AddItem("Silver Dagger", 1)
	s.Require().NoError(err)
	_, err = handler.AddItem("Silver Dagger", 1)
	s.Require().NoError(err)

	items, err := s.Q.ListPlayerItemInventory(ctx, s.player.ID)
	s.Require().NoError(err)
	s.Require().Len(items, 1)
	s.Equal("Silver Dagger", items[0].Name)
	s.Equal(int32(2), items[0].Quantity)
}

func (s *InventoryServiceSuite) TestRemoveItemDecrementsThenDeletes() {
	handler := s.handler()
	_, err := handler.AddItem("Silver Dagger", 2)
	s.Require().NoError(err)

	// First remove decrements to 1.
	_, err = handler.RemoveItem("Silver Dagger", 1)
	s.Require().NoError(err)
	items, err := s.Q.ListPlayerItemInventory(context.Background(), s.player.ID)
	s.Require().NoError(err)
	s.Require().Len(items, 1)
	s.Equal(int32(1), items[0].Quantity)

	// Second remove deletes the join row entirely.
	_, err = handler.RemoveItem("Silver Dagger", 1)
	s.Require().NoError(err)
	items, err = s.Q.ListPlayerItemInventory(context.Background(), s.player.ID)
	s.Require().NoError(err)
	s.Empty(items)
}

func (s *InventoryServiceSuite) TestFetchInventory() {
	handler := s.handler()
	_, err := handler.AddAbility("Shadow Step", 1)
	s.Require().NoError(err)
	_, err = handler.AddItem("Silver Dagger", 1)
	s.Require().NoError(err)

	inv, err := handler.FetchInventory()
	s.Require().NoError(err)
	s.Equal(s.player.ID, inv.Player.ID)
	s.Equal("Mafia", inv.Role.Name)
	s.Require().Len(inv.Abilities, 1)
	s.Equal("Shadow Step", inv.Abilities[0].Name)
	s.Require().Len(inv.Items, 1)
	s.Equal("Silver Dagger", inv.Items[0].Name)
}

func TestInventoryServiceSuite(t *testing.T) {
	suite.Run(t, new(InventoryServiceSuite))
}
