package datasync_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/require"
)

// TestMain bootstraps through testutil: production guard, advisory lock with
// the other DB suites, migrations applied, tables truncated between tests.
func TestMain(m *testing.M) {
	os.Exit(testutil.Bootstrap(m))
}

func mustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testutil.NewTestPool(t)
	t.Cleanup(pool.Close)
	testutil.TruncateAll(t, pool)
	return pool
}

func seedRole(t *testing.T, pool *pgxpool.Pool, name, desc string, alignment models.Alignment) models.Role {
	t.Helper()
	r, err := models.New(pool).CreateRole(context.Background(), models.CreateRoleParams{
		Name: name, Description: desc, Alignment: alignment,
	})
	require.NoError(t, err)
	return r
}

func TestPlanRolesDiff(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()
	q := models.New(pool)

	// Existing: RoleA (GOOD, old desc). Sheet changes desc → update; RoleB is new.
	seedRole(t, pool, "RoleA", "old description", models.AlignmentGOOD)

	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(roleCSV), models.AlignmentGOOD)
	require.NoError(t, err)
	require.Empty(t, warnings)

	plan, err := datasync.PlanRoles(ctx, q, models.AlignmentGOOD, docs)
	require.NoError(t, err)
	require.Len(t, plan.Roles, 2)
	// RoleA exists with a stale description → update; everything else (RoleA's
	// abilities/perks + the whole RoleB chunk) is new.
	require.Equal(t, 6, plan.Counts[datasync.ActionCreate])
	require.Equal(t, 1, plan.Counts[datasync.ActionUpdate])

	a := plan.Roles[0]
	require.Equal(t, "RoleA", a.Doc.Name)
	require.Equal(t, datasync.ActionUpdate, a.Action)
	require.Contains(t, strings.Join(a.Changes, " "), "description")

	b := plan.Roles[1]
	require.Equal(t, "RoleB", b.Doc.Name)
	require.Equal(t, datasync.ActionCreate, b.Action)
	require.Equal(t, datasync.ActionCreate, b.Abilities[0].Action)
	require.Equal(t, datasync.ActionCreate, b.Perks[0].Action)
}

func TestPlanRolesSkipWhenUnchanged(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()

	// Apply the sheet once so the DB matches it exactly, then re-plan:
	// every entry must come back as skip (upsert idempotency).
	docs, _, err := datasync.ParseRolesCSV(strings.NewReader(roleCSV), models.AlignmentGOOD)
	require.NoError(t, err)
	plan, err := datasync.PlanRoles(ctx, models.New(pool), models.AlignmentGOOD, docs)
	require.NoError(t, err)
	require.NoError(t, datasync.ApplyRoles(ctx, pool, plan))

	again, err := datasync.PlanRoles(ctx, models.New(pool), models.AlignmentGOOD, docs)
	require.NoError(t, err)
	require.Equal(t, 7, again.Counts[datasync.ActionSkip], "all roles+abilities+perks unchanged")
	require.Zero(t, again.Counts[datasync.ActionCreate])
	require.Zero(t, again.Counts[datasync.ActionUpdate])
}

func TestPlanItemsDiff(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()
	q := models.New(pool)

	// Existing item with the old description → update. Other rows create.
	_, err := models.New(pool).CreateItem(ctx, models.CreateItemParams{
		Name: "Sword of Testing", Description: "old sword desc",
		Rarity: models.RarityRARE, Cost: 50,
	})
	require.NoError(t, err)

	docs, warnings, err := datasync.ParseItemsCSV(strings.NewReader(itemCSV))
	require.NoError(t, err)
	require.Len(t, warnings, 2)

	plan, err := datasync.PlanItems(ctx, q, docs)
	require.NoError(t, err)
	require.Len(t, plan.Items, 2)
	require.Equal(t, 1, plan.Counts[datasync.ActionCreate]) // Free Thing
	require.Equal(t, 1, plan.Counts[datasync.ActionUpdate]) // Sword of Testing
	require.Contains(t, strings.Join(plan.Items[0].Changes, " "), "description")
}

func TestApplyRolesCreatesAndLinks(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()

	docs, _, err := datasync.ParseRolesCSV(strings.NewReader(roleCSV), models.AlignmentGOOD)
	require.NoError(t, err)
	plan, err := datasync.PlanRoles(ctx, models.New(pool), models.AlignmentGOOD, docs)
	require.NoError(t, err)

	require.NoError(t, datasync.ApplyRoles(ctx, pool, plan))

	q := models.New(pool)
	a, err := q.GetRoleByName(ctx, "RoleA")
	require.NoError(t, err)
	require.Equal(t, "Role description A", a.Description)
	require.Equal(t, models.AlignmentGOOD, a.Alignment)

	b, err := q.GetRoleByName(ctx, "RoleB")
	require.NoError(t, err)
	require.Equal(t, "Role description B", b.Description)

	// Ability + perk joined to RoleB.
	roleAbilities, err := q.ListRoleAbilityForRole(ctx, b.ID)
	require.NoError(t, err)
	require.Len(t, roleAbilities, 1)
	require.Equal(t, "Ability Three", roleAbilities[0].Name)

	rolePerks, err := q.ListRolePerkForRole(ctx, b.ID)
	require.NoError(t, err)
	require.Len(t, rolePerks, 1)
	require.Equal(t, "Passive Two", rolePerks[0].Name)

	// Category join landed for the COMMON ability (Stealth/Combat categories
	// don't exist in the schema yet, so those links are skipped — but the
	// ability itself exists with ROLE_SPECIFIC rarity from the sheet's ^).
	one, err := q.GetAbilityInfoByName(ctx, "Ability One")
	require.NoError(t, err)
	require.Equal(t, models.RarityCOMMON, one.Rarity)
}

func TestApplyIsIdempotent(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()

	docs, _, err := datasync.ParseItemsCSV(strings.NewReader(itemCSV))
	require.NoError(t, err)
	plan, err := datasync.PlanItems(ctx, models.New(pool), docs)
	require.NoError(t, err)

	require.NoError(t, datasync.ApplyItems(ctx, pool, plan))
	require.NoError(t, datasync.ApplyItems(ctx, pool, plan), "second apply is a no-op, not an error")

	q := models.New(pool)
	items, err := q.ListItem(ctx)
	require.NoError(t, err)
	require.Len(t, items, 2, "no duplicate items after double apply")
}

func TestSeedSourcesAndPreserveEditedURL(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()

	svc := datasync.New(pool, map[string]string{
		"good_roles": "https://sheets.example/good.csv",
		"items":      "https://sheets.example/items.csv",
	})
	require.NoError(t, svc.SeedSources(ctx))

	sources, err := svc.ListSources(ctx)
	require.NoError(t, err)
	require.Len(t, sources, 4)
	byName := map[string]models.SyncSource{}
	for _, s := range sources {
		byName[s.Name] = s
	}
	require.Equal(t, "https://sheets.example/good.csv", byName["good_roles"].Url)
	require.True(t, byName["good_roles"].Enabled)

	// A user edits the URL in the panel; re-seeding must not clobber it.
	edited, err := svc.UpdateSource(ctx, byName["good_roles"].ID, "https://sheets.example/edited.csv", true)
	require.NoError(t, err)
	require.Equal(t, "https://sheets.example/edited.csv", edited.Url)

	require.NoError(t, svc.SeedSources(ctx))
	after, err := svc.ListSources(ctx)
	require.NoError(t, err)
	for _, s := range after {
		if s.Name == "good_roles" {
			require.Equal(t, "https://sheets.example/edited.csv", s.Url, "re-seed keeps the edited URL")
		}
	}
}

func TestRecordRunAndLastRuns(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()

	svc := datasync.New(pool, map[string]string{"good_roles": "https://sheets.example/good.csv"})
	require.NoError(t, svc.SeedSources(ctx))
	sources, err := svc.ListSources(ctx)
	require.NoError(t, err)
	var src models.SyncSource
	for _, s := range sources {
		if s.Name == "good_roles" {
			src = s
		}
	}
	require.NotZero(t, src.ID)

	require.NoError(t, svc.RecordRun(ctx, &src.ID, "good_roles", datasync.RunStatusApplied, "web",
		"", map[datasync.Action]int{datasync.ActionCreate: 3, datasync.ActionUpdate: 1, datasync.ActionSkip: 40}))

	runs, err := svc.LastRuns(ctx, 5)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	run := runs[0]
	require.Equal(t, "good_roles", run.SourceName)
	require.Equal(t, datasync.RunStatusApplied, run.Status)
	require.Equal(t, "web", run.RunBy)

	var counts map[string]int
	require.NoError(t, json.Unmarshal(run.ActionCounts, &counts))
	require.Equal(t, 3, counts["create"])
	require.Equal(t, 1, counts["update"])
}
