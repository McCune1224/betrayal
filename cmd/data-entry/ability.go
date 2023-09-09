package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mccune1224/betrayal/internal/data"
)

func (app *application) ParseAnyAbilityCsv(filepath string) error {
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

func GetAnyAbilities(csv [][]string) ([]data.Ability, error) {
	var abilities []data.Ability

	for i, entry := range csv {
		if i == 0 {
			continue
		}

		ability := data.Ability{
			Rarity:      entry[1],
			Name:        entry[2],
			Description: entry[3],
		}
		fmt.Println("ABILITY: ", ability)
		abilities = append(abilities, ability)
	}

	return abilities, nil
}
