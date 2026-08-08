// Command data-entry syncs the game catalog from Google Sheets CSV exports.
//
// ARCHIVED (2026-08): the web panel's /sync page is the primary surface for
// this workflow (fetch → preview diff → validate → apply). This CLI remains as
// a thin wrapper over the shared internal/services/datasync engine for
// scripting: same env vars (GOOD_ROLES_CSV / EVIL_ROLES_CSV /
// NEUTRAL_ROLES_CSV / ITEM_CSV), same behavior — upsert by name, additive,
// nothing deleted.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
)

func main() {
	dsn := os.Getenv("DATABASE_POOLER_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_POOLER_URL environment variable not set")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create database pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	svc := datasync.New(pool, envURLs())
	if err := svc.SeedSources(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "seed sync sources: %v\n", err)
		os.Exit(1)
	}

	runAll(ctx, pool, svc)
}

// envURLs builds the source-name → CSV URL map from the environment.
func envURLs() map[string]string {
	out := map[string]string{}
	for _, src := range datasync.CanonicalSources {
		key := map[string]string{
			"good_roles":    "GOOD_ROLES_CSV",
			"evil_roles":    "EVIL_ROLES_CSV",
			"neutral_roles": "NEUTRAL_ROLES_CSV",
			"items":         "ITEM_CSV",
		}[src.Name]
		out[src.Name] = os.Getenv(key)
	}
	return out
}

// runAll fetches, diffs, and applies every enabled source in sequence,
// printing the same style of summary the original tool produced.
func runAll(ctx context.Context, pool *pgxpool.Pool, svc *datasync.Service) {
	q := models.New(pool)
	sources, err := svc.ListSources(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list sources: %v\n", err)
		os.Exit(1)
	}

	failed := false
	recordFailed := func(src models.SyncSource, err error) {
		_ = svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusFailed, "cli", err.Error(), nil)
	}
	for _, src := range sources {
		if !src.Enabled {
			fmt.Printf("⚠ skipping %s (disabled)\n", src.Name)
			continue
		}
		if src.Url == "" {
			fmt.Printf("⚠ skipping %s (no CSV URL configured)\n", src.Name)
			continue
		}

		start := time.Now()
		fmt.Printf("\n== %s ==\n", src.Name)

		switch src.Kind {
		case "roles":
			body, err := svc.Fetch(ctx, src)
			if err != nil {
				recordFailed(src, err)
				fail(src.Name, err)
				failed = true
				continue
			}
			docs, warnings, err := datasync.ParseRolesCSV(body, models.Alignment(src.Alignment))
			body.Close()
			if err != nil {
				recordFailed(src, err)
				fail(src.Name, err)
				failed = true
				continue
			}
			plan, err := datasync.PlanRoles(ctx, q, models.Alignment(src.Alignment), docs)
			if err != nil {
				recordFailed(src, err)
				fail(src.Name, err)
				failed = true
				continue
			}
			plan.Warnings = append(plan.Warnings, warnings...)
			for _, w := range plan.Warnings {
				fmt.Printf("  ⚠ %s\n", w)
			}
			if err := datasync.ApplyRoles(ctx, pool, plan); err != nil {
				_ = svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusFailed, "cli", err.Error(), plan.Counts)
				fail(src.Name, err)
				failed = true
				continue
			}
			_ = svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusApplied, "cli", "", plan.Counts)
			fmt.Printf("  ✓ %d roles applied in %s (%d created, %d updated, %d unchanged)\n",
				len(plan.Roles), time.Since(start).Round(time.Millisecond),
				plan.Counts[datasync.ActionCreate], plan.Counts[datasync.ActionUpdate], plan.Counts[datasync.ActionSkip])

		case "items":
			body, err := svc.Fetch(ctx, src)
			if err != nil {
				recordFailed(src, err)
				fail(src.Name, err)
				failed = true
				continue
			}
			docs, warnings, err := datasync.ParseItemsCSV(body)
			body.Close()
			if err != nil {
				recordFailed(src, err)
				fail(src.Name, err)
				failed = true
				continue
			}
			plan, err := datasync.PlanItems(ctx, q, docs)
			if err != nil {
				recordFailed(src, err)
				fail(src.Name, err)
				failed = true
				continue
			}
			plan.Warnings = append(plan.Warnings, warnings...)
			for _, w := range plan.Warnings {
				fmt.Printf("  ⚠ %s\n", w)
			}
			if err := datasync.ApplyItems(ctx, pool, plan); err != nil {
				_ = svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusFailed, "cli", err.Error(), plan.Counts)
				fail(src.Name, err)
				failed = true
				continue
			}
			_ = svc.RecordRun(ctx, &src.ID, src.Name, datasync.RunStatusApplied, "cli", "", plan.Counts)
			fmt.Printf("  ✓ %d items applied in %s (%d created, %d updated, %d unchanged)\n",
				len(plan.Items), time.Since(start).Round(time.Millisecond),
				plan.Counts[datasync.ActionCreate], plan.Counts[datasync.ActionUpdate], plan.Counts[datasync.ActionSkip])
		}
	}

	if failed {
		os.Exit(1)
	}
	fmt.Println("\nAll data synchronization tasks completed!")
}

func fail(name string, err error) {
	fmt.Fprintf(os.Stderr, "✗ %s failed: %v\n", name, err)
}
