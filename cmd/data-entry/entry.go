package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"

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
