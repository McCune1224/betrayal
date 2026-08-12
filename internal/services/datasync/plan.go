package datasync

import (
	"context"
	"fmt"

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
	roles, err := q.Listrole(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	abilities, err := q.ListAbilityInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("list abilities: %w", err)
	}
	perks, err := q.ListPerkInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("list perks: %w", err)
	}
	categories, err := q.ListCategory(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	rolesByName := make(map[string]models.Role, len(roles))
	for _, role := range roles {
		rolesByName[role.Name] = role
	}
	abilitiesByName := make(map[string]models.AbilityInfo, len(abilities))
	for _, ability := range abilities {
		abilitiesByName[ability.Name] = ability
	}
	perksByName := make(map[string]models.PerkInfo, len(perks))
	for _, perk := range perks {
		perksByName[perk.Name] = perk
	}
	categoriesByName := make(map[string]models.Category, len(categories))
	for _, category := range categories {
		categoriesByName[category.Name] = category
	}

	for _, doc := range docs {
		rp := RolePlan{
			Doc:       doc,
			Abilities: []AbilityPlan{},
			Perks:     []PerkPlan{},
		}

		existing, exists := rolesByName[doc.Name]
		if !exists {
			rp.Action = ActionCreate
			rp.Changes = append(rp.Changes, "new role")
		} else {
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
			existingAbility, exists := abilitiesByName[a.Name]
			if !exists {
				ap.Action = ActionCreate
				ap.Changes = append(ap.Changes, "new ability")
			} else {
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
				if _, exists := categoriesByName[cat]; !exists {
					plan.Warnings = append(plan.Warnings,
						fmt.Sprintf("ability %q: category %q not found — link will be skipped", a.Name, cat))
				}
			}
		}

		// Perks.
		for _, p := range doc.Perks {
			pp := PerkPlan{Doc: p}
			existingPerk, exists := perksByName[p.Name]
			if !exists {
				pp.Action = ActionCreate
				pp.Changes = append(pp.Changes, "new passive")
			} else {
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
	items, err := q.ListItem(ctx)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	categories, err := q.ListCategory(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	itemsByName := make(map[string]models.Item, len(items))
	for _, item := range items {
		itemsByName[item.Name] = item
	}
	categoriesByName := make(map[string]models.Category, len(categories))
	for _, category := range categories {
		categoriesByName[category.Name] = category
	}

	for _, doc := range docs {
		ip := ItemPlan{Doc: doc}
		existing, exists := itemsByName[doc.Name]
		if !exists {
			ip.Action = ActionCreate
			ip.Changes = append(ip.Changes, "new item")
		} else {
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
			if _, exists := categoriesByName[cat]; !exists {
				plan.Warnings = append(plan.Warnings,
					fmt.Sprintf("item %q: category %q not found — link will be skipped", doc.Name, cat))
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
