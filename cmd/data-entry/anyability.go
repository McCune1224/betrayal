package main

import (
	"log"
	"strings"

	"github.com/mccune1224/betrayal/pkg/data"
)

func (*csvBuilder) BuildAnyAbilityCSV(csv [][]string) ([]data.AnyAbility, error) {
	var anyAbilities []data.AnyAbility
	for i, entry := range csv {
		var aa data.AnyAbility
		if i == 0 || i == 1 {
			continue
		}
		aa.Rarity = entry[1]
		// check for [] in name and if so parse it out
		if strings.Contains(entry[2], "[") {
			// grab string inside [] and assign to RoleSpecific
			aa.RoleSpecific = strings.Split(entry[2], "[")[1]
			// drop the ] at the end
			aa.RoleSpecific = strings.Split(aa.RoleSpecific, "]")[0]
			aa.Name = strings.Split(entry[2], " [")[0]
		} else {
			aa.Name = entry[2]
			aa.RoleSpecific = ""
		}
		aa.Description = entry[3]

		log.Println(aa.Name)

		anyAbilities = append(anyAbilities, aa)
	}

	return anyAbilities, nil
}
