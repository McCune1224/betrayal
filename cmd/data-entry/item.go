package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
)

func (app *application) ParseItemCsv(filepath string) error {
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

func GetItems(csv [][]string) ([]data.Item, error) {
	var items []data.Item

	for i, entry := range csv {

		if i == 0 || i == 1 || (i == len(csv)-1) {
			continue
		}

		item := data.Item{
			Rarity:      entry[1],
			Name:        entry[2],
			Description: entry[5],
		}

		strCost := entry[3]

		// FIXME: This is stinky and very specific to the item csv, Too Bad!

		if strCost == "X" {
			item.Cost = 0
		} else {
			cost, err := strconv.ParseInt(strCost, 10, 64)
			if err != nil {
				return nil, err
			}
			item.Cost = cost
		}

		categories := entry[4]
		parsedCategories := strings.Split(categories, "/")
		item.Categories = pq.StringArray(parsedCategories)

		items = append(items, item)
	}

	return items, nil
}
