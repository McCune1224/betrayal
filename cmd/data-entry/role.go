package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/data"
)

type csvRole struct {
	Name             string
	Description      string
	AbilitiesStrings []string
	PerksStrings     []string
}

// Name returns the name of the role (with tidy up of spaces)
func (c *csvRole) GetName() string {
	// snip any whitespaces from the name
	return strings.TrimSpace(c.Name)
}

func (c *csvRole) GetDescription() string {
	return strings.TrimSpace(c.Description)
}

func (c *csvRole) GetAbilities() ([]data.Ability, error) {
	abilities := []data.Ability{}
	for i, currAbilityString := range c.AbilitiesStrings {

		name := ""
		currAbility := data.Ability{}
		for _, char := range currAbilityString {
			if char == '[' {
				break
			}
			name += string(char)
		}
		// Everything past the name, since this is easier to deal with without the name
		snipString := currAbilityString[len(name):]

		chargeIndex := strings.Index(snipString, " ")
		if chargeIndex == -1 {
			return nil, fmt.Errorf("failed line %d: failed charge parse %s", i, currAbilityString)
		}
		charge := snipString[:chargeIndex]

		abilityTypeIndex := strings.Index(charge, "]") + 1
		if abilityTypeIndex == 0 {
			return nil, fmt.Errorf("failed line %d: failed ability type parse %s", i, currAbilityString)
		}

		abilityType := charge[abilityTypeIndex:]
		switch abilityType {
		case "*":
			currAbility.AnyAbility = true
		default:
			currAbility.AnyAbility = false
		}

		dashSplit := strings.Split(currAbilityString, "- ")[1:]
		description := strings.Join(dashSplit, "- ")

		// Get number inside of [ ]
		foo := strings.Index(charge, "[")
		if foo == -1 {
			return nil, fmt.Errorf("failed line %d: failed charge parse %s", i, currAbilityString)
		}

		bar := strings.Index(charge, "]")
		if bar == -1 {
			return nil, fmt.Errorf("failed line %d: failed charge parse %s", i, currAbilityString)
		}

		charge = charge[foo+1 : bar]
		if charge == "∞" {
			charge = "-1"
		} else {
			charge = charge[1:]
		}

		chargeInt, err := strconv.Atoi(charge)
		if err != nil {
			return nil, fmt.Errorf("failed line %d: failed charge int convERT %s", i, currAbilityString)
		}

		categoriesOpenIndex := strings.Index(snipString, "(")
		categoriesCloseIndex := strings.Index(snipString, ")")

		if categoriesOpenIndex == -1 || categoriesCloseIndex == -1 {
			return nil, fmt.Errorf("failed line %d: failed category parse %s", i, currAbilityString)
		}
		categoriesString := snipString[categoriesOpenIndex+1 : categoriesCloseIndex]
		categoriesParse := strings.Split(categoriesString, "/")
		categories := []string{}
		for _, category := range categoriesParse {
			category = strings.TrimSpace(category)
			categories = append(categories, category)
		}

		currAbility.Name = strings.TrimSpace(name)
		currAbility.Description = strings.TrimSpace(description)
		currAbility.Charges = chargeInt
		currAbility.Categories = categories

		abilities = append(abilities, currAbility)
	}
	return abilities, nil
}

func (c *csvRole) GetPerks() ([]data.Perk, error) {
	splitPerks := []data.Perk{}
	for _, perkString := range c.PerksStrings {
		split := strings.Split(perkString, "- ")
		if len(split) < 2 {
			return nil, errors.New("failed to split perk string:\n " + perkString + "\n")
		}

		if len(split) > 2 {
			for i := 2; i < len(split); i++ {
				split[1] += split[i]
			}
		}

		splitPerks = append(splitPerks, data.Perk{
			Name:        strings.TrimSpace(split[0]),
			Description: strings.TrimSpace(split[1]),
		})
	}
	return splitPerks, nil
}

func (c *csvRole) ToDBEntry(alignment string) (data.Role, error) {
	name := c.GetName()
	description := c.GetDescription()
	return data.Role{
		Name:        name,
		Description: description,
		Alignment:   alignment,
	}, nil
}

func (*csvBuilder) BuildRoleCSV(csv [][]string) ([]csvRole, error) {
	roleList := []csvRole{}
	if len(csv) == 0 {
		return nil, errors.New("csv is empty")
	}

	currRole := csvRole{}
	for i, line := range csv {
		switch line[1] {
		case "Name ": // ,Name ,Description
			roleName := csv[i+1][1]
			roleDescription := csv[i+1][2]

			currRole.Name = roleName
			currRole.Description = roleDescription

		case "Abilities:": // ,Abilities:,
			abilities := []string{}
			for j := i + 1; j < len(csv); j++ {
				// If we hit Perk: then we are done with abilities
				if csv[j][1] == "Perks:" {
					// Go through each ability and parse it
					break
				}
				abilities = append(abilities, csv[j][1])
			}
			currRole.AbilitiesStrings = abilities
		case "Perks:": // ,Perks:,
			perks := []string{}

			for j := i + 1; j < len(csv); j++ {
				if csv[j][1] == "" {
					break
				}
				perks = append(perks, csv[j][1])
			}
			currRole.PerksStrings = perks
		case "": // ,,
			roleList = append(roleList, currRole)
		default: // Empty line
			if i == len(csv)-1 {
				roleList = append(roleList, currRole)
			}
			continue
		}
	}
	return roleList, nil
}
