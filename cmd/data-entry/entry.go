package main

import (
	"encoding/csv"
	"errors"
	"os"
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

type ability struct {
	Name        string
	Description string
	AbilityType string
	Charges     int
}
type perk struct {
	Name        string
	Description string
}
type csvRole struct {
	Name            string
	Description     string
	AbilitiesString []string
	PerksString     []string
}

/*
	Convert Ability string chunks to Ability struct

Examples lines to parse include:
Solve [x0]* (Investigation/Positive/Non-visiting) - Figure out any piece of information of your choice about a player. Gain a charge for this every even day if you are Detective.
Soul Seer [∞]^ (Investigation/Neutral/Non-visiting) - Select a player, upon their death, you can see their role, their items, money and their last actions. You gain a charge upon the selected player dying. If used on yourself and you die, you may continue to make chats with the living.
*/
func (c *csvRole) SanitizeAbilities() ([]ability, error) {
	name := ""
	//name is up until we hit the first [
	for _, char := range c.AbilitiesString[0] {
		if char == '[' {
			break
		}
		name += string(char)
	}
	return nil, nil
}

// Convert Perk string chunks to Perk struct
func (c *csvRole) SanitizePerks() ([]perk, error) {
	splitPerks := []perk{}
	for _, perkString := range c.PerksString {
		split := strings.Split(perkString, "- ")
		if len(split) != 2 {
			return nil, errors.New("Failed to split perk string")
		}
		splitPerks = append(splitPerks, perk{
			Name:        split[0],
			Description: split[1],
		})
	}
	return nil, nil
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
		if len(line) == 0 {
			continue
		}
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
			currRole.AbilitiesString = abilities
		case "Perks:": // ,Perks:,
			perks := []string{}

			for j := i + 1; j < len(a.csv); j++ {
				if a.csv[j][1] == "" {
					break
				}
				perks = append(perks, a.csv[j][1])
			}
			currRole.PerksString = perks
		case "": // ,,
			// fmt.Println("NEW ROLE ===============================")
			roleList = append(roleList, currRole)
		default: // Empty line
			continue
		}
	}
	return roleList, nil
}

func (a *application) InsertAbility(ability data.Ability) {

}

func (a *application) InsertPerk(perk data.Perk) {
}
