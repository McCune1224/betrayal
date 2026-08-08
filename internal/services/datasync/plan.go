package datasync

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/mccune1224/betrayal/internal/models"
)

// PlanRoles diffs parsed role documents against the database (read-only).
// Matching is by exact name. Returns a plan the UI can render and ApplyRoles
// can execute.
func PlanRoles(ctx context.Context, q *models.Queries, alignment models.Alignment, docs []RoleDoc) (*RoleSourcePlan, error) {
	plan := &RoleSourcePlan{
		Alignment: alignment,
		Counts:    map[Action]int{},
	}

	for _, doc := range docs {
		rp := RolePlan{
			Doc:       doc,
			Abilities: []AbilityPlan{},
			Perks:     []PerkPlan{},
		}

		existing, err := q.GetRoleByName(ctx, doc.Name)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			rp.Action = ActionCreate
			rp.Changes = append(rp.Changes, "new role")
		case err != nil:
			return nil, fmt.Errorf("lookup role %q: %w", doc.Name, err)
		default:
			rp.OldDesc = existing.Description
			rp.Changes = diffString(existing.Description, doc.Description, "description", rp.Changes)
			if existing.Alignment != doc.Alignment {
				rp.Changes = append(rp.Changes,
					fmt.Sprintf("alignment: %s → %s", existing.Alignment, doc.Alignment))
			}
			rp.Action = actionFor(rp.Changes)
		}
		plan.Counts[rp.Action]++

		// Abilities.
		for _, a := range doc.Abilities {
			ap := AbilityPlan{Doc: a}
			existingAbility, err := q.GetAbilityInfoByName(ctx, a.Name)
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				ap.Action = ActionCreate
				ap.Changes = append(ap.Changes, "new ability")
			case err != nil:
				return nil, fmt.Errorf("lookup ability %q: %w", a.Name, err)
			default:
				ap.OldDesc = existingAbility.Description
				ap.Changes = diffString(existingAbility.Description, a.Description, "description", ap.Changes)
				if existingAbility.DefaultCharges != a.DefaultCharges {
					ap.Changes = append(ap.Changes, fmt.Sprintf("charges: %s → %s",
						ChargesDisplay(existingAbility.DefaultCharges), ChargesDisplay(a.DefaultCharges)))
				}
				if existingAbility.Rarity != a.Rarity {
					ap.Changes = append(ap.Changes, fmt.Sprintf("rarity: %s → %s", existingAbility.Rarity, a.Rarity))
				}
				if existingAbility.AnyAbility != a.AnyAbility {
					ap.Changes = append(ap.Changes, fmt.Sprintf("any-ability: %t → %t", existingAbility.AnyAbility, a.AnyAbility))
				}
				ap.Action = actionFor(ap.Changes)
			}
			plan.Counts[ap.Action]++
			rp.Abilities = append(rp.Abilities, ap)

			// Flag categories that don't exist yet so the UI can warn before apply.
			for _, cat := range a.Categories {
				if _, err := q.GetCategoryByName(ctx, cat); errors.Is(err, pgx.ErrNoRows) {
					plan.Warnings = append(plan.Warnings,
						fmt.Sprintf("ability %q: category %q not found — link will be skipped", a.Name, cat))
				} else if err != nil {
					return nil, fmt.Errorf("lookup category %q: %w", cat, err)
				}
			}
		}

		// Perks.
		for _, p := range doc.Perks {
			pp := PerkPlan{Doc: p}
			existingPerk, err := q.GetPerkInfoByName(ctx, p.Name)
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				pp.Action = ActionCreate
				pp.Changes = append(pp.Changes, "new passive")
			case err != nil:
				return nil, fmt.Errorf("lookup perk %q: %w", p.Name, err)
			default:
				pp.OldDesc = existingPerk.Description
				pp.Changes = diffString(existingPerk.Description, p.Description, "description", pp.Changes)
				pp.Action = actionFor(pp.Changes)
			}
			plan.Counts[pp.Action]++
			rp.Perks = append(rp.Perks, pp)
		}

		plan.Roles = append(plan.Roles, rp)
	}
	return plan, nil
}

// PlanItems diffs parsed item documents against the database (read-only).
func PlanItems(ctx context.Context, q *models.Queries, docs []ItemDoc) (*ItemSourcePlan, error) {
	plan := &ItemSourcePlan{Counts: map[Action]int{}}

	for _, doc := range docs {
		ip := ItemPlan{Doc: doc}
		existing, err := q.GetItemByName(ctx, doc.Name)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			ip.Action = ActionCreate
			ip.Changes = append(ip.Changes, "new item")
		case err != nil:
			return nil, fmt.Errorf("lookup item %q: %w", doc.Name, err)
		default:
			ip.OldDesc = existing.Description
			ip.Changes = diffString(existing.Description, doc.Description, "description", ip.Changes)
			if existing.Rarity != doc.Rarity {
				ip.Changes = append(ip.Changes, fmt.Sprintf("rarity: %s → %s", existing.Rarity, doc.Rarity))
			}
			if existing.Cost != doc.Cost {
				ip.Changes = append(ip.Changes, fmt.Sprintf("cost: %d → %d", existing.Cost, doc.Cost))
			}
			ip.Action = actionFor(ip.Changes)
		}
		plan.Counts[ip.Action]++
		plan.Items = append(plan.Items, ip)

		for _, cat := range doc.Categories {
			if _, err := q.GetCategoryByName(ctx, cat); errors.Is(err, pgx.ErrNoRows) {
				plan.Warnings = append(plan.Warnings,
					fmt.Sprintf("item %q: category %q not found — link will be skipped", doc.Name, cat))
			} else if err != nil {
				return nil, fmt.Errorf("lookup category %q: %w", cat, err)
			}
		}
	}
	return plan, nil
}

// diffString appends a "field: old → new" change when the values differ.
func diffString(oldV, newV, field string, changes []string) []string {
	if oldV != newV {
		changes = append(changes, fmt.Sprintf("%s: %q → %q", field, oldV, newV))
	}
	return changes
}

// actionFor classifies a diff: any change → update, otherwise skip.
func actionFor(changes []string) Action {
	if len(changes) > 0 {
		return ActionUpdate
	}
	return ActionSkip
}
