package main

import (
	"encoding/csv"
	"os"
	"strings"
)

// Take in a filepath and attach CSV data to app struct
func (app *application) ParseCsv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	entries, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	app.csv = entries

	return nil

}

type AnyAbilityLine struct {
	Rarity      string
	Name        string
	Description string
}

// Update the database with new AA's from CSV
func (app *application) UpdateAbilityRarites() []AnyAbilityLine {
	dbAbilityUpdater := app.models.Abilities
	updates := []AnyAbilityLine{}
	for i, line := range app.csv {
		var aal AnyAbilityLine
		// skip row 0 and 1
		if i == 0 || i == 1 {
			continue
		}

		aal.Rarity = strings.TrimSpace(line[1])
		aal.Name = strings.TrimSpace(line[2])
		aal.Description = strings.TrimSpace(line[3])

		if aal.Rarity == "" || aal.Name == "" || aal.Description == "" {
			app.logger.Println("!! FAILED TO PARSE LINE", i)
			continue
		}

		currAA, err := dbAbilityUpdater.GetByName(aal.Name)
		if err != nil {
			app.logger.Printf("Error updating '%s', Skipping", aal.Name)
			continue
		}
		currAA.Rarity = aal.Rarity
		err = dbAbilityUpdater.Update(currAA)
		if err != nil {
			app.logger.Printf("Error updating '%s', Skipping", aal.Name)
			continue
		}
	}
	return updates
}
