package datasync

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

// ApplyRoles executes a role source plan inside ONE transaction. Sheet edits
// propagate (upsert by name), new rows are created, and nothing is deleted.
// Missing categories are skipped (they were already flagged as warnings in
// the preview plan).
func ApplyRoles(ctx context.Context, pool *pgxpool.Pool, plan *RoleSourcePlan) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := models.New(tx)

	for _, rp := range plan.Roles {
		roleID, err := upsertRole(ctx, q, rp)
		if err != nil {
			return err
		}

		for _, ap := range rp.Abilities {
			if ap.Action == ActionSkip {
				continue
			}
			abilityID, err := upsertAbility(ctx, q, ap.Doc)
			if err != nil {
				return err
			}
			if err := q.CreateRoleAbilityJoin(ctx, models.CreateRoleAbilityJoinParams{
				RoleID: roleID, AbilityID: abilityID,
			}); err != nil {
				return fmt.Errorf("link ability %q to role %q: %w", ap.Doc.Name, rp.Doc.Name, err)
			}
			if err := linkCategories(ctx, q, ap.Doc.Categories, func(catID int32) error {
				return q.CreateAbilityCategoryJoin(ctx, models.CreateAbilityCategoryJoinParams{
					AbilityID: abilityID, CategoryID: catID,
				})
			}); err != nil {
				return fmt.Errorf("link categories for ability %q: %w", ap.Doc.Name, err)
			}
		}

		for _, pp := range rp.Perks {
			if pp.Action == ActionSkip {
				continue
			}
			perkID, err := upsertPerk(ctx, q, pp.Doc)
			if err != nil {
				return err
			}
			if err := q.CreateRolePerkJoin(ctx, models.CreateRolePerkJoinParams{
				RoleID: roleID, PerkID: perkID,
			}); err != nil {
				return fmt.Errorf("link perk %q to role %q: %w", pp.Doc.Name, rp.Doc.Name, err)
			}
		}
	}
	return tx.Commit(ctx)
}

// ApplyItems executes an item source plan inside ONE transaction.
func ApplyItems(ctx context.Context, pool *pgxpool.Pool, plan *ItemSourcePlan) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := models.New(tx)

	for _, ip := range plan.Items {
		if ip.Action == ActionSkip {
			continue
		}
		itemID, err := upsertItem(ctx, q, ip.Doc)
		if err != nil {
			return err
		}
		if err := linkCategories(ctx, q, ip.Doc.Categories, func(catID int32) error {
			return q.CreateItemCategoryJoin(ctx, models.CreateItemCategoryJoinParams{
				ItemID: itemID, CategoryID: catID,
			})
		}); err != nil {
			return fmt.Errorf("link categories for item %q: %w", ip.Doc.Name, err)
		}
	}
	return tx.Commit(ctx)
}

func upsertRole(ctx context.Context, q *models.Queries, rp RolePlan) (int32, error) {
	existing, err := q.GetRoleByName(ctx, rp.Doc.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		r, err := q.CreateRole(ctx, models.CreateRoleParams{
			Name: rp.Doc.Name, Description: rp.Doc.Description, Alignment: rp.Doc.Alignment,
		})
		if err != nil {
			return 0, fmt.Errorf("create role %q: %w", rp.Doc.Name, err)
		}
		return r.ID, nil
	case err != nil:
		return 0, fmt.Errorf("lookup role %q: %w", rp.Doc.Name, err)
	default:
		r, err := q.UpdateRole(ctx, models.UpdateRoleParams{
			ID: existing.ID, Name: rp.Doc.Name, Description: rp.Doc.Description, Alignment: rp.Doc.Alignment,
		})
		if err != nil {
			return 0, fmt.Errorf("update role %q: %w", rp.Doc.Name, err)
		}
		return r.ID, nil
	}
}

func upsertAbility(ctx context.Context, q *models.Queries, doc AbilityDoc) (int32, error) {
	existing, err := q.GetAbilityInfoByName(ctx, doc.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		a, err := q.CreateAbilityInfo(ctx, models.CreateAbilityInfoParams{
			Name: doc.Name, Description: doc.Description,
			DefaultCharges: doc.DefaultCharges, Rarity: doc.Rarity, AnyAbility: doc.AnyAbility,
		})
		if err != nil {
			return 0, fmt.Errorf("create ability %q: %w", doc.Name, err)
		}
		return a.ID, nil
	case err != nil:
		return 0, fmt.Errorf("lookup ability %q: %w", doc.Name, err)
	default:
		a, err := q.UpdateAbilityInfo(ctx, models.UpdateAbilityInfoParams{
			ID: existing.ID, Name: doc.Name, Description: doc.Description,
			DefaultCharges: doc.DefaultCharges, Rarity: doc.Rarity, AnyAbility: doc.AnyAbility,
		})
		if err != nil {
			return 0, fmt.Errorf("update ability %q: %w", doc.Name, err)
		}
		return a.ID, nil
	}
}

func upsertPerk(ctx context.Context, q *models.Queries, doc PerkDoc) (int32, error) {
	existing, err := q.GetPerkInfoByName(ctx, doc.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		p, err := q.CreatePerkInfo(ctx, models.CreatePerkInfoParams{
			Name: doc.Name, Description: doc.Description,
		})
		if err != nil {
			return 0, fmt.Errorf("create perk %q: %w", doc.Name, err)
		}
		return p.ID, nil
	case err != nil:
		return 0, fmt.Errorf("lookup perk %q: %w", doc.Name, err)
	default:
		p, err := q.UpdatePerkInfo(ctx, models.UpdatePerkInfoParams{
			ID: existing.ID, Name: doc.Name, Description: doc.Description,
		})
		if err != nil {
			return 0, fmt.Errorf("update perk %q: %w", doc.Name, err)
		}
		return p.ID, nil
	}
}

func upsertItem(ctx context.Context, q *models.Queries, doc ItemDoc) (int32, error) {
	existing, err := q.GetItemByName(ctx, doc.Name)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		it, err := q.CreateItem(ctx, models.CreateItemParams{
			Name: doc.Name, Description: doc.Description, Rarity: doc.Rarity, Cost: doc.Cost,
		})
		if err != nil {
			return 0, fmt.Errorf("create item %q: %w", doc.Name, err)
		}
		return it.ID, nil
	case err != nil:
		return 0, fmt.Errorf("lookup item %q: %w", doc.Name, err)
	default:
		it, err := q.UpdateItem(ctx, models.UpdateItemParams{
			ID: existing.ID, Name: doc.Name, Description: doc.Description, Rarity: doc.Rarity, Cost: doc.Cost,
		})
		if err != nil {
			return 0, fmt.Errorf("update item %q: %w", doc.Name, err)
		}
		return it.ID, nil
	}
}

// linkCategories resolves category names to IDs and invokes fn per ID.
// Unknown categories are skipped silently (they are surfaced as warnings in
// the preview plan).
func linkCategories(ctx context.Context, q *models.Queries, names []string, fn func(catID int32) error) error {
	for _, name := range names {
		cat, err := q.GetCategoryByName(ctx, name)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if err := fn(cat.ID); err != nil {
			return err
		}
	}
	return nil
}
