package data

import (
	"errors"

	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/util"
)

// Ability that can be rolled in AA events
type AnyAbility struct {
	ID           int64          `db:"id"`
	Name         string         `db:"name"`
	Description  string         `db:"description"`
	Categories   pq.StringArray `db:"categories"`
	Rarity       string         `db:"rarity"`
	RoleSpecific string         `db:"role_specific"`
}

func (am *AbilityModel) InsertAnyAbility(aa *AnyAbility) error {
	query := `INSERT INTO any_abilities (name, description, categories, rarity, role_specific)
    VALUES ($1, $2, $3, $4, $5)`
	_, err := am.DB.Exec(
		query,
		aa.Name,
		aa.Description,
		aa.Categories,
		aa.Rarity,
		aa.RoleSpecific,
	)
	if err != nil {
		return err
	}
	return nil
}

func (am *AbilityModel) GetAnyAbility(id int64) (*AnyAbility, error) {
	var aa AnyAbility
	err := am.DB.Get(&aa, "SELECT * FROM any_abilities WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &aa, nil
}

func (am *AbilityModel) GetAllAnyAbilities() ([]AnyAbility, error) {
	var anyAbilities []AnyAbility
	err := am.DB.Select(&anyAbilities, "SELECT * FROM any_abilities")
	if err != nil {
		return nil, err
	}
	return anyAbilities, nil
}

func (am *AbilityModel) GetAnyAbilityByName(name string) (*AnyAbility, error) {
	var aa AnyAbility
	err := am.DB.Get(&aa, "SELECT * FROM any_abilities WHERE name ILIKE $1", name)
	if err != nil {
		return nil, err
	}
	return &aa, nil
}

func (am *AbilityModel) GetAnyAbilitybyFuzzy(name string) (*AnyAbility, error) {
	var aa AnyAbility
	am.DB.Get(&aa, "SELECT * FROM any_abilities WHERE name ILIKE $1", name)
	if aa.Name != "" {
		return &aa, nil
	}

	var aaChoices []AnyAbility
	if len(name) < 2 {
		return nil, errors.New("search term must be at least 2 characters")
	}
	err := am.DB.Select(&aaChoices, "SELECT * FROM any_abilities")
	if err != nil {
		return nil, err
	}
	strChoices := make([]string, len(aaChoices))
	for i := range aaChoices {
		strChoices[i] = aaChoices[i].Name
	}
	best, _ := util.FuzzyFind(name, strChoices)
	for _, a := range aaChoices {
		if a.Name == best {
			aa = a
		}
	}
	return &aa, nil
}

func (am *AbilityModel) GetAnyAbilityByCategory(category string) ([]AnyAbility, error) {
	var anyAbilities []AnyAbility
	err := am.DB.Select(&anyAbilities, "SELECT * FROM any_abilities WHERE categories ILIKE $1", category)
	if err != nil {
		return nil, err
	}
	return anyAbilities, nil
}

func (am *AbilityModel) GetAnyAbilityByRarity(rarity string) ([]AnyAbility, error) {
	var anyAbilities []AnyAbility
	err := am.DB.Select(&anyAbilities, "SELECT * FROM any_abilities WHERE rarity ILIKE $1", rarity)
	if err != nil {
		return nil, err
	}
	return anyAbilities, nil
}

func (am *AbilityModel) GetRandomAnyAbilityByRarity(rarity string) (*AnyAbility, error) {
	var aa AnyAbility
	err := am.DB.Get(
		&aa,
		"SELECT * FROM any_abilities WHERE rarity ILIKE $1 ORDER BY RANDOM() LIMIT 1",
		rarity,
	)
	if err != nil {
		return nil, err
	}
	return &aa, nil
}

func (am *AbilityModel) UpdateAnyAbility(aa *AnyAbility) error {
	query := `UPDATE any_abilities SET name = $1, description = $2, categories = $3, rarity = $4, role_specific = $5 WHERE id = $6`
	_, err := am.DB.Exec(
		query,
		aa.Name,
		aa.Description,
		aa.Categories,
		aa.Rarity,
		aa.RoleSpecific,
		aa.ID,
	)
	if err != nil {
		return err
	}
	return nil
}
