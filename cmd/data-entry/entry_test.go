package main

import (
	"testing"
)

//
// func TestParseCsv(t *testing.T) {
// 	testApp := &application{}
// 	err := testApp.ParseCsv("./fat-dumpy/good_roles.csv")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log("Test Passed")
//
// 	for _, line := range testApp.csv {
// 		t.Log(line)
// 	}
// }

// func TestParseRoles(t *testing.T) {
// 	testApp := &application{}
// 	err := testApp.ParseCsv("./fat-dumpy/good_roles.csv")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	roles, err := testApp.ParseRoles("role")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log(roles)
// }

// ,Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,
func TestParseAbilityLine(t *testing.T) {
	testApp := &application{}
	err := testApp.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		t.Fatal(err)
	}

	// ability := ",Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,"
	ability2 := ",Pandemonium [x2]* (Debuff/Negative/Visiting) - Make a target player mad about being a role. You choose the player and the role. The target player must make a concerted effort to portray themselves as that role to other players for a 48 hour period. Breaking madness leads to insta-death, bypassing everything.,"
	parsedAbility, err := testApp.parseAbilityLine(ability2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(parsedAbility.Name)
	t.Log(parsedAbility.Effect)
	t.Log(parsedAbility.Charges)
	t.Log(parsedAbility.Categories)
	t.Log(parsedAbility.IsRoleSpecific)
	t.Log(parsedAbility.IsAnyAbility)

}
