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

// Simple Interface to write to databse to not worry about differnt model types
type DBWriter interface {
	Insert(modelType interface{}, jsonPayload []byte)
}

// Load in csv and append to app struct
func (app *application) ParseCsv(filepath string) error {
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

// example csv line:
// ,Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,
func (a *application) parseAbilityLine(line string) (data.Ability, error) {
	name := ""
	description := ""
	charges := ""
	abilityType := ""
	categories := []string{}

	if len(line) == 0 {
		return data.Ability{}, errors.New("Line is empty")
	}
	// Parsing:
	// 1. Drop the commas
	// 2. Split on the first dash
	// 3. Clean up spaces
	// 4. Description is everything after the dash
	// 5. Left half contains name, charges, and categories
	//      a. Name is first word
	//      b. Charges is second word parsed [x3] -> 3
	//      c. Categories is in parens and split on slash (Investigation/Neutral/Visiting) -> [Investigation, Neutral, Visiting]

	// 1. Drop the commas
	line = line[1 : len(line)-1]

	// 2. Split on the first dash
	dashParse := strings.Split(line, "-")

	// 3. Clean up spaces
	dashParse[0] = strings.TrimSpace(dashParse[0])
	dashParse[1] = strings.TrimSpace(dashParse[1])

	// 4. Description is everything after the dash
	description = dashParse[1]

	// 5. Left half contains name, charges, and categories
	left := strings.Split(dashParse[0], " ")

	// 5a. Name is first word
	name = left[0]

	// 5b. Clean up charges
	charges = left[1]
	//ability type is indicated by the last character(s) of the charges
	// * = Any Ability
	// ** = Declare as Undercover
	// ^ = Role Specific
	abilityType = strings.Split(charges, "]")[1]
	// 5b. Charges is second word parsed [x3] -> 3
	charges = strings.Split(charges, "[")[1]
	charges = strings.Split(charges, "]")[0]
	charges = strings.Replace(charges, "x", "", -1)
	if charges == "∞" {
		charges = "-1"
	}

	// 5c. Categories is in parens and split on slash (Investigation/Neutral/Visiting) -> [Investigation, Neutral, Visiting]
	rawCategories := left[2]
	// Drop parens
	rawCategories = rawCategories[1 : len(rawCategories)-1]
	// parse on slash
	categories = strings.Split(rawCategories, "/")

	intCharges, err := strconv.Atoi(charges)
	if err != nil {
		return data.Ability{}, err
	}
	roleSpecific := false
	anyAbility := true
	switch abilityType {
	case "*":
		roleSpecific = false
		anyAbility = true

	case "**":
		roleSpecific = false
		anyAbility = false
	case "^":
		roleSpecific = true
		anyAbility = false

	}
	return data.Ability{
		Name:           name,
		Effect:         description,
		Charges:        intCharges,
		Categories:     categories,
		IsRoleSpecific: roleSpecific,
		IsAnyAbility:   anyAbility,
	}, nil
}

func (a *application) ParseRoles(roleType string) ([]data.Role, error) {
	roles := []data.Role{}
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

	currRole := data.Role{}
	for i, line := range a.csv {
		if len(line) == 0 {
			continue
		}
		fmt.Println(line[1 : len(line)-1])
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
					for _, ability := range abilities {
						// TODO: Parse ability
						// parsedAbility, err := parseAbilityLine(ability)
						fmt.Println(ability)

					}
					break
				}
				abilities = append(abilities, a.csv[j][1])
			}
		case "Perks:": // ,Perks:,
			perks := []string{}

			for j := i + 1; j < len(a.csv); j++ {
				if a.csv[j][1] == "" {
					break
				}
				perks = append(perks, a.csv[j][1])
			}
		case "": // ,,
			fmt.Println("NEW ROLE ===============================")
		default: // Empty line
			continue
		}
	}
	return roles, nil
}

func (a *application) InsertAbility(ability data.Ability) {

}

func (a *application) InsertPerk(perk data.Perk) {
}
