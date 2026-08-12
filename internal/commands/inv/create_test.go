package inv

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.Bootstrap(m))
}

// knownStatusNames is the set of statuses seeded by migration 000008. Every
// immunity/status referenced by the role ops map must exist in it — a typo'd
// name previously inserted status id 0 and failed `/inv create` with a foreign
// key violation.
var knownStatusNames = map[string]bool{
	"Cursed":       true,
	"Death Cursed": true,
	"Frozen":       true,
	"Paralyzed":    true,
	"Burned":       true,
	"Empowered":    true,
	"Drunk":        true,
	"Restrained":   true,
	"Disabled":     true,
	"Blackmailed":  true,
	"Despaired":    true,
	"Madness":      true,
	"Lucky":        true,
	"Unlucky":      true,
}

// TestRoleOpsData guards the per-role creation-time adjustments: every role
// that had a switch arm is present, all referenced statuses exist, and the
// roles fixed during the switch→map conversion behave correctly.
func TestRoleOpsData(t *testing.T) {
	for role, ops := range roleOpsByRole {
		for _, immunity := range ops.immunities {
			assert.True(t, knownStatusNames[immunity], "role %q references unknown immunity status %q", role, immunity)
		}
		for _, status := range ops.statuses {
			assert.True(t, knownStatusNames[status], "role %q references unknown status %q", role, status)
		}
	}

	// Every role covered by the old switch must still be covered.
	for _, role := range []string{
		"cerberus", "detective", "fisherman", "hero", "nurse", "terminal",
		"wizard", "yeti", "cyborg", "entertainer", "magician", "masochist",
		"succubus", "arsonist", "cultist", "director", "gatekeeper", "hacker",
		"highwayman", "imp", "threatener",
	} {
		_, ok := roleOpsByRole[role]
		assert.True(t, ok, "missing role ops for %q", role)
	}

	// Regressions from the switch→map conversion:
	// - magician's "Lucky" was mislabeled as an immunity; it should be a status
	//   (like entertainer, same perk).
	assert.Contains(t, roleOpsByRole["magician"].statuses, "Lucky")
	assert.Contains(t, roleOpsByRole["magician"].immunities, "Unlucky")
	// - succubus referenced "Blackmail" (no such status); must be "Blackmailed".
	assert.Contains(t, roleOpsByRole["succubus"].immunities, "Blackmailed")
	assert.NotContains(t, roleOpsByRole["succubus"].immunities, "Blackmail")
	// - cultist referenced "Curse" (no such status); must be "Cursed".
	assert.Contains(t, roleOpsByRole["cultist"].immunities, "Cursed")
	assert.NotContains(t, roleOpsByRole["cultist"].immunities, "Curse")

	// Item limit overrides preserved.
	require.NotNil(t, roleOpsByRole["fisherman"].itemLimit)
	assert.Equal(t, int32(8), *roleOpsByRole["fisherman"].itemLimit)
	require.NotNil(t, roleOpsByRole["threatener"].itemLimit)
	assert.Equal(t, int32(6), *roleOpsByRole["threatener"].itemLimit)
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	godotenv.Load("../../../.env")
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	require.NoError(t, err, "failed to connect to local test database (DATABASE_URL)")
	return pool
}

// TestGameConfigInt verifies the /inv create defaults: rows from game_config
// win, missing or unparseable rows fall back to the current defaults.
func TestGameConfigInt(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	q := models.New(pool)
	ctx := context.Background()
	const key = "unit_test_default_coins"

	require.NoError(t, q.DeleteGameConfig(ctx, key))
	defer q.DeleteGameConfig(ctx, key)

	// Missing row -> fallback.
	assert.Equal(t, int32(200), gameConfigInt(ctx, q, key, 200))

	// Valid row -> configured value.
	_, err := q.UpsertGameConfig(ctx, models.UpsertGameConfigParams{Key: key, Value: "250"})
	require.NoError(t, err)
	assert.Equal(t, int32(250), gameConfigInt(ctx, q, key, 200))

	// Garbage row -> fallback.
	_, err = q.UpsertGameConfig(ctx, models.UpsertGameConfigParams{Key: key, Value: "lots"})
	require.NoError(t, err)
	assert.Equal(t, int32(200), gameConfigInt(ctx, q, key, 200))
}
