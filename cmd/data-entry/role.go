package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/data"
)

// Load in csv and append to app struct
func (app *application) ParseRoleCsv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}
	app.csv = records

	return nil
}

type csvRole struct {
	Name             string
	Description      string
	AbilitiesStrings []string
	PerksStrings     []string
}

/*
	Convert Ability string chunks to Ability struct

Examples lines to parse include:
Solve [x0]* (Investigation/Positive/Non-visiting) - Figure out any piece of information of your choice about a player. Gain a charge for this every even day if you are Detective.
Soul Seer [∞]^ (Investigation/Neutral/Non-visiting) - Select a player, upon their death, you can see their role, their items, money and their last actions. You gain a charge upon the selected player dying. If used on yourself and you die, you may continue to make chats with the living.
*/
func (c *csvRole) GetAbilities() ([]data.Ability, error) {

	abilities := []data.Ability{}
	for i, currAbilityString := range c.AbilitiesStrings {
		// fmt.Println(currAbilityString)

		name := ""
		currAbility := data.Ability{}
		for _, char := range currAbilityString {
			if char == '[' {
				break
			}
			name += string(char)
		}
		//Everything past the name, since this is easier to deal with without the name
		snipString := currAbilityString[len(name):]

		chargeIndex := strings.Index(snipString, " ")
		if chargeIndex == -1 {
			return nil, errors.New(
				fmt.Sprintf("FAILED LINE %d: FAILED CHARGE PARSE %s", i, currAbilityString),
			)
		}
		charge := snipString[:chargeIndex]

		abilityTypeIndex := strings.Index(charge, "]") + 1
		if abilityTypeIndex == 0 {
			return nil, errors.New(
				fmt.Sprintf("FAILED LINE %d: FAILED ABILITY TYPE PARSE %s", i, currAbilityString),
			)
		}

		abilityType := charge[abilityTypeIndex:]
		switch abilityType {
		case "*":
			currAbility.AnyAbility = true
		default:
			currAbility.AnyAbility = false
		}

		description := strings.Split(currAbilityString, "- ")[1]

		// Get number inside of [ ]
		foo := strings.Index(charge, "[")
		if foo == -1 {
			return nil, errors.New(
				fmt.Sprintf("FAILED LINE %d: FAILED CHARGE PARSE %s", i, currAbilityString),
			)
		}

		bar := strings.Index(charge, "]")
		if bar == -1 {
			return nil, errors.New(
				fmt.Sprintf("FAILED LINE %d: FAILED CHARGE PARSE %s", i, currAbilityString),
			)
		}

		charge = charge[foo+1 : bar]
		if charge == "∞" {
			charge = "-1"
		} else {
			charge = charge[1:]
		}

		chargeInt, err := strconv.Atoi(charge)
		if err != nil {
			return nil, errors.New(
				fmt.Sprintf("FAILED LINE %d: FAILED CHARGE INT CONVERT %s", i, currAbilityString),
			)
		}

		categoriesOpenIndex := strings.Index(snipString, "(")
		categoriesCloseIndex := strings.Index(snipString, ")")

		if categoriesOpenIndex == -1 || categoriesCloseIndex == -1 {
			return nil, errors.New(
				fmt.Sprintf("FAILED LINE %d: FAILED CATEGORY PARSE %s", i, currAbilityString),
			)
		}
		categoriesString := snipString[categoriesOpenIndex+1 : categoriesCloseIndex]
		categoriesParse := strings.Split(categoriesString, "/")
		categories := []string{}
		for _, category := range categoriesParse {
			category = strings.TrimSpace(category)
			categories = append(categories, category)
		}

		currAbility.Name = name
		currAbility.Description = description
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
			return nil, errors.New("Failed to split perk string:\n " + perkString + "\n")
		}

		if len(split) > 2 {
			for i := 2; i < len(split); i++ {
				split[1] += split[i]
			}
		}

		splitPerks = append(splitPerks, data.Perk{

			Name:        split[0],
			Description: split[1],
		})
	}
	return splitPerks, nil
}

// Parse Roles in CSV's to string chunks
func (a *application) SplitRoles(roleType string) ([]csvRole, error) {
	roleList := []csvRole{}
	if len(a.csv) == 0 {
		return nil, errors.New("csv is empty")
	}

	// Each entry in csv for a role will look like this:
	// ,Name ,Description
	// ,Agent,Perhaps we're asking the wrong questions.
	// ,Abilities:,
	// ,Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,
	// ,Spy [x3]* (Investigation/Neutral/Visiting) - You will know everything that happens to your target and who is doing what to them for 24 hours on use.,
	// ,"Hidden [x2]* (Redirection/Neutral/Non-visiting) - For 24 hours on use, anything done on you will be reflected back to the user.",
	// ,Perks:,
	// ,"Hawkeye - If someone steals from you, you will know who did it.",
	// ,"Organised - Your vote cannot be stolen, blocked or tampered with in any way.",
	// ,"Tracker - If anyone in your alliance does a positive action to someone outside of your alliance, you will know who gave who, but not what it was.",

	currRole := csvRole{}
	for i, line := range a.csv {
		// fmt.Println(line[1 : len(line)-1])
		switch line[1] {
		case "Name ": // ,Name ,Description
			roleName := a.csv[i+1][1]
			roleDescription := a.csv[i+1][2]

			currRole.Name = roleName
			currRole.Description = roleDescription

		case "Abilities:": // ,Abilities:,
			abilities := []string{}
			for j := i + 1; j < len(a.csv); j++ {
				// If we hit Perk: then we are done with abilities
				if a.csv[j][1] == "Perks:" {
					// Go through each ability and parse it
					break
				}
				abilities = append(abilities, a.csv[j][1])
			}
			currRole.AbilitiesStrings = abilities
		case "Perks:": // ,Perks:,
			perks := []string{}

			for j := i + 1; j < len(a.csv); j++ {
				if a.csv[j][1] == "" {
					break
				}
				perks = append(perks, a.csv[j][1])
			}
			currRole.PerksStrings = perks
		case "": // ,,
			// fmt.Println("NEW ROLE ===============================")
			roleList = append(roleList, currRole)
		default: // Empty line
			continue
		}
	}
	return roleList, nil
}
