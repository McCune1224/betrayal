package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/data"
)

// example csv line:
// ,Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,

// Charge can look like any of the following:
// [x3]*
// (x3)*
// [x3]**
func parseChargeAndAbilityType(
	charge string,
) (parsedCharge int, parsedAbilityType string, err error) {
	//Split on the closing symbol of either ] or )
	//The first half is the charge, the second half is the ability type
	//If the ability type is not present, it is assumed to be *
	parsedCharge = 0
	closingIndex := 0
	if charge[0] == '[' {
		closingIndex = strings.Index(charge, "]")
	} else {
		closingIndex = strings.Index(charge, ")")
	}
	if closingIndex == -1 {
		return 0, "", errors.New("Failed to parse () or [] in charge, " + charge)
	}
	// slice off the opening symbol and the x
	stringCharge := charge[2:closingIndex]

	if stringCharge == "∞" {
		parsedCharge = -1
	} else {

		parsedCharge, err = strconv.Atoi(stringCharge)
	}

	//convert charge to int
	//If the charge is ∞, it is represented by -1
	return parsedCharge, charge[closingIndex+1:], nil
}

func (a *application) parseAbilityLine(line string) (data.Ability, error) {
	name := ""
	description := ""
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

	charge, abilityType, err := parseChargeAndAbilityType(left[1])
	if err != nil {
		return data.Ability{}, err
	}

	// 5c. Categories is in parens and split on slash (Investigation/Neutral/Visiting) -> [Investigation, Neutral, Visiting]
	rawCategories := left[2]
	// Drop parens
	rawCategories = rawCategories[1 : len(rawCategories)-1]
	// parse on slash
	categories = strings.Split(rawCategories, "/")

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
		Charges:        charge,
		Categories:     categories,
		IsRoleSpecific: roleSpecific,
		IsAnyAbility:   anyAbility,
	}, nil
}
