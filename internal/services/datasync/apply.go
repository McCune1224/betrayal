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
// Every role's abilities/perks are upserted and linked REGARDLESS of the
// plan's action: an unchanged ability referenced by a NEW role must still get
// its role_ability join, and the join queries are ON CONFLICT DO NOTHING so
// re-applies are idempotent. Missing categories are skipped (they were
// already flagged as warnings in the preview plan).
func ApplyRoles(ctx context.Context, pool *pgxpool.Pool, plan *RoleSourcePlan) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := ApplyRolesTx(ctx, tx, plan); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ApplyRolesTx applies a role plan using the caller's transaction. This is
// used by the all-or-nothing game reset flow as well as the normal sync page.
func ApplyRolesTx(ctx context.Context, tx models.DBTX, plan *RoleSourcePlan) error {
	q := models.New(tx)
	roleIDs := make(map[string]int32, len(plan.Roles))
	abilityIDs := make(map[string]int32)
	perkIDs := make(map[string]int32)
	categoryIDs := make(map[string]int32)

	for _, rp := range plan.Roles {
		roleID, ok := roleIDs[rp.Doc.Name]
		if !ok {
			var err error
			roleID, err = upsertRole(ctx, q, rp)
			if err != nil {
				return err
			}
			roleIDs[rp.Doc.Name] = roleID
		}

		for _, ap := range rp.Abilities {
			abilityID, ok := abilityIDs[ap.Doc.Name]
			if !ok {
				var err error
				abilityID, err = upsertAbility(ctx, q, ap.Doc)
				if err != nil {
					return err
				}
				abilityIDs[ap.Doc.Name] = abilityID
			}
			if err := q.CreateRoleAbilityJoin(ctx, models.CreateRoleAbilityJoinParams{
				RoleID: roleID, AbilityID: abilityID,
			}); err != nil {
				return fmt.Errorf("link ability %q to role %q: %w", ap.Doc.Name, rp.Doc.Name, err)
			}
			if err := linkCategories(ctx, q, ap.Doc.Categories, categoryIDs, func(catID int32) error {
				return q.CreateAbilityCategoryJoin(ctx, models.CreateAbilityCategoryJoinParams{
					AbilityID: abilityID, CategoryID: catID,
				})
			}); err != nil {
				return fmt.Errorf("link categories for ability %q: %w", ap.Doc.Name, err)
			}
		}

		for _, pp := range rp.Perks {
			perkID, ok := perkIDs[pp.Doc.Name]
			if !ok {
				var err error
				perkID, err = upsertPerk(ctx, q, pp.Doc)
				if err != nil {
					return err
				}
				perkIDs[pp.Doc.Name] = perkID
			}
			if err := q.CreateRolePerkJoin(ctx, models.CreateRolePerkJoinParams{
				RoleID: roleID, PerkID: perkID,
			}); err != nil {
				return fmt.Errorf("link perk %q to role %q: %w", pp.Doc.Name, rp.Doc.Name, err)
			}
		}
	}
	return nil
}

// ApplyItems executes an item source plan inside ONE transaction. Like
// ApplyRoles, items are upserted and category-linked regardless of the plan
// action so unchanged items still pick up new category links from the sheet.
func ApplyItems(ctx context.Context, pool *pgxpool.Pool, plan *ItemSourcePlan) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := ApplyItemsTx(ctx, tx, plan); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ApplyItemsTx applies an item plan using the caller's transaction.
func ApplyItemsTx(ctx context.Context, tx models.DBTX, plan *ItemSourcePlan) error {
	q := models.New(tx)
	itemIDs := make(map[string]int32, len(plan.Items))
	categoryIDs := make(map[string]int32)

	for _, ip := range plan.Items {
		itemID, ok := itemIDs[ip.Doc.Name]
		if !ok {
			var err error
			itemID, err = upsertItem(ctx, q, ip.Doc)
			if err != nil {
				return err
			}
			itemIDs[ip.Doc.Name] = itemID
		}
		if err := linkCategories(ctx, q, ip.Doc.Categories, categoryIDs, func(catID int32) error {
			return q.CreateItemCategoryJoin(ctx, models.CreateItemCategoryJoinParams{
				ItemID: itemID, CategoryID: catID,
			})
		}); err != nil {
			return fmt.Errorf("link categories for item %q: %w", ip.Doc.Name, err)
		}
	}
	return nil
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
func linkCategories(ctx context.Context, q *models.Queries, names []string, cache map[string]int32, fn func(catID int32) error) error {
	for _, name := range names {
		catID, ok := cache[name]
		if !ok {
			cat, err := q.GetCategoryByName(ctx, name)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			catID = cat.ID
			cache[name] = catID
		}
		if err := fn(catID); err != nil {
			return err
		}
	}
	return nil
}
