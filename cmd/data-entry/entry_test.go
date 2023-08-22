package main

import (
	"testing"

	"github.com/mccune1224/betrayal/internal/data"
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

func TestParseRoles(t *testing.T) {
	testApp := &application{}
	err := testApp.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		t.Fatal(err)
	}

	roles, err := testApp.ParseRoles("role")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(roles)
}

// ,Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,
func TestParseAbilityLine(t *testing.T) {
	testApp := &application{}
	err := testApp.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input         string
		output        data.Ability
		expectedError error
	}{
		{
			",Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,",
			data.Ability{
				Name:           "Follow",
				Effect:         "You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.",
				Charges:        3,
				Categories:     []string{"Investigation", "Neutral", "Visiting"},
				IsRoleSpecific: false,
				IsAnyAbility:   false,
			},
			nil,
		}, {
			",Pandemonium [x2]* (Debuff/Negative/Visiting) - Make a target player mad about being a role. You choose the player and the role. The target player must make a concerted effort to portray themselves as that role to other players for a 48 hour period. Breaking madness leads to insta-death, bypassing everything.,",
			data.Ability{
				Name:           "Pandemonium",
				Effect:         "Make a target player mad about being a role. You choose the player and the role. The target player must make a concerted effort to portray themselves as that role to other players for a 48 hour period. Breaking madness leads to insta-death, bypassing everything.",
				Charges:        2,
				Categories:     []string{"Debuff", "Negative", "Visiting"},
				IsRoleSpecific: false,
				IsAnyAbility:   false,
			},
			nil,
		},
	}
	// ability := ",Follow [x3]* (Investigation/Neutral/Visiting) - You will know everything your target does and who to for 24 hours from use. Does not include perk-based actions.,"
	// ability2 := ",Pandemonium [x2]* (Debuff/Negative/Visiting) - Make a target player mad about being a role. You choose the player and the role. The target player must make a concerted effort to portray themselves as that role to other players for a 48 hour period. Breaking madness leads to insta-death, bypassing everything.,"
	for _, test := range tests {
		parsedAbility, err := testApp.parseAbilityLine(test.input)
		if err != nil {
			t.Error(err.Error())
		}
		if parsedAbility.Name != test.output.Name {
			t.Errorf("Expected name to be %s, got %s", test.output.Name, parsedAbility.Name)
		}
		if parsedAbility.Effect != test.output.Effect {
			t.Errorf("Expected effect to be %s, got %s", test.output.Effect, parsedAbility.Effect)
		}
		if parsedAbility.Charges != test.output.Charges {
			t.Errorf(
				"Expected charges to be %d, got %d",
				test.output.Charges,
				parsedAbility.Charges,
			)
		}
		if len(parsedAbility.Categories) != len(test.output.Categories) {
			t.Errorf(
				"Expected categories to be %v, got %v",
				test.output.Categories,
				parsedAbility.Categories,
			)
		}
		if parsedAbility.IsAnyAbility != test.output.IsAnyAbility {
			t.Errorf(
				"Expected isAnyAbility to be %v, got %v",
				test.output.IsAnyAbility,
				parsedAbility.IsAnyAbility,
			)
		}
		if parsedAbility.IsRoleSpecific != test.output.IsRoleSpecific {
			t.Errorf(
				"Expected isRoleSpecific to be %v, got %v",
				test.output.IsRoleSpecific,
				parsedAbility.IsRoleSpecific,
			)
		}
		t.Log("Test passed", test.input, parsedAbility)
	}
}
