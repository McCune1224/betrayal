package main

import (
	"testing"
)

func TestParseCsv(t *testing.T) {
	app := &application{}
	err := app.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		t.Errorf("Error parsing csv: %v", err)
	}
}

func TestSplitRoles(t *testing.T) {
	app := &application{}
	err := app.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		t.Errorf("Error parsing csv: %v", err)
	}
	_, err = app.SplitRoles("role")
	if err != nil {
		t.Errorf("Error splitting roles: %v", err)
	}
	// for _, role := range roles {
	// 	fmt.Println("===============================================")
	// 	t.Log("___NAME___: ", role.Name)
	// 	t.Log("___DESCRIPTION___: ", role.Description)
	// 	t.Log("___ABILITIES___: ", role.AbilitiesString)
	// 	t.Log("___PERKS___ : ", role.PerksString)
	// 	fmt.Println(len(role.AbilitiesString), len(role.PerksString))
	// 	fmt.Println("===============================================")
	// 	fmt.Println()
	// }
}

// func TestSanitizeAbilities(t *testing.T) {
// 	app := &application{}
// 	err := app.ParseCsv("./fat-dumpy/good_roles.csv")
// 	if err != nil {
// 		t.Errorf("Error parsing csv: %v", err)
// 	}
// 	roles, err := app.SplitRoles("role")
// 	if err != nil {
// 		t.Errorf("Error splitting roles: %v", err)
// 	}
// 	for _, role := range roles {
// 		_, err := role.SanitizeAbilities()
// 		if err != nil {
// 			t.Errorf("Error sanitizing abilities: %v", err)
// 		}
// 	}
// }

func TestSanitizePerks(t *testing.T) {
	app := &application{}
	err := app.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		t.Errorf("Error parsing csv: %v", err)
	}
	roles, err := app.SplitRoles("role")
	if err != nil {
		t.Errorf("Error splitting roles: %v", err)
	}
	for _, role := range roles {
		_, err := role.SanitizePerks()
		if err != nil {
			t.Errorf("Error sanitizing perks: %v", err)
		}
	}
}
