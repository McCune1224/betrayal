package gamereset

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
)

// Summary describes the rows that will be cleared by a reset.
type Summary struct {
	Players   int64
	Votes     int64
	CycleRows int64
	Items     int64
	Roles     int64
	Abilities int64
	Perks     int64
	SyncRuns  int64
	AuditRows int64
	LogRows   int64
}

// Result reports the completed reset/import operation.
type Result struct {
	Summary Summary
	Sources []string
}

type Service struct {
	pool *pgxpool.Pool
	sync *datasync.Service
	mu   sync.Mutex
}

func New(pool *pgxpool.Pool, syncService *datasync.Service) *Service {
	return &Service{pool: pool, sync: syncService}
}

func (s *Service) ListSources(ctx context.Context) ([]models.SyncSource, error) {
	return s.sync.ListSources(ctx)
}

// Preview returns counts without changing the database.
func (s *Service) Preview(ctx context.Context) (Summary, error) {
	out := Summary{}
	var err error
	for table, dest := range map[string]*int64{
		"player":        &out.Players,
		"vote":          &out.Votes,
		"game_cycle":    &out.CycleRows,
		"item":          &out.Items,
		"role":          &out.Roles,
		"ability_info":  &out.Abilities,
		"perk_info":     &out.Perks,
		"sync_run":      &out.SyncRuns,
		"command_audit": &out.AuditRows,
		"logs":          &out.LogRows,
	} {
		err = s.pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(dest)
		if err != nil {
			return Summary{}, fmt.Errorf("count %s: %w", table, err)
		}
	}
	return out, nil
}

// Execute fetches and parses every configured source before opening the reset
// transaction. Once the transaction starts, clearing and re-importing are
// atomic: a bad source or failed insert rolls back the entire operation.
func (s *Service) Execute(ctx context.Context) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	before, err := s.Preview(ctx)
	if err != nil {
		return Result{}, err
	}
	sources, err := s.sync.ListSources(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("load sync sources: %w", err)
	}
	plans := make([]preparedSource, 0, len(sources))
	for _, src := range sources {
		if !src.Enabled || src.Url == "" {
			return Result{}, fmt.Errorf("source %q is disabled or has no URL", src.Name)
		}
		prepared, err := s.prepare(ctx, src)
		if err != nil {
			return Result{}, fmt.Errorf("prepare %s: %w", src.Name, err)
		}
		plans = append(plans, prepared)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, resetSQL); err != nil {
		return Result{}, fmt.Errorf("clear game data: %w", err)
	}
	for _, prepared := range plans {
		switch {
		case prepared.roles != nil:
			if err := datasync.ApplyRolesTx(ctx, tx, prepared.roles); err != nil {
				return Result{}, fmt.Errorf("import %s: %w", prepared.name, err)
			}
		case prepared.items != nil:
			if err := datasync.ApplyItemsTx(ctx, tx, prepared.items); err != nil {
				return Result{}, fmt.Errorf("import %s: %w", prepared.name, err)
			}
		}
	}
	if _, err := tx.Exec(ctx, "INSERT INTO game_cycle (is_elimination, day) VALUES (FALSE, 0)"); err != nil {
		return Result{}, fmt.Errorf("reset day state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit reset: %w", err)
	}

	names := make([]string, len(plans))
	for i, plan := range plans {
		names[i] = plan.name
	}
	return Result{Summary: before, Sources: names}, nil
}

type preparedSource struct {
	name  string
	roles *datasync.RoleSourcePlan
	items *datasync.ItemSourcePlan
}

func (s *Service) prepare(ctx context.Context, src models.SyncSource) (preparedSource, error) {
	body, err := s.sync.Fetch(ctx, src)
	if err != nil {
		return preparedSource{}, err
	}
	defer body.Close()
	q := models.New(s.pool)
	prepared := preparedSource{name: src.Name}
	switch src.Kind {
	case "roles":
		docs, _, err := datasync.ParseRolesCSV(body, models.Alignment(src.Alignment))
		if err != nil {
			return preparedSource{}, err
		}
		plan, err := datasync.PlanRoles(ctx, q, models.Alignment(src.Alignment), docs)
		if err != nil {
			return preparedSource{}, err
		}
		prepared.roles = plan
	case "items":
		docs, _, err := datasync.ParseItemsCSV(body)
		if err != nil {
			return preparedSource{}, err
		}
		plan, err := datasync.PlanItems(ctx, q, docs)
		if err != nil {
			return preparedSource{}, err
		}
		prepared.items = plan
	default:
		return preparedSource{}, fmt.Errorf("unsupported source kind %q", src.Kind)
	}
	return prepared, nil
}

// resetSQL deliberately preserves game configuration, Discord channel
// configuration, sync source URLs, built-in statuses, and categories. It
// clears players, ownership, votes, day state, audit/log history, sync history,
// and all catalog rows that are rebuilt from the four CSV sources.
const resetSQL = `
TRUNCATE TABLE
  player_confessional, player_immunity, player_note, player_item,
  player_status, player_perk, player_ability, vote, player,
  role_ability, role_perk, ability_category, item_category,
  ability_info, perk_info, item, role, game_cycle, sync_run,
  command_audit, logs
RESTART IDENTITY CASCADE;
`
