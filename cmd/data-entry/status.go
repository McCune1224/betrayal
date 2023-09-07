package main

import (
	"encoding/csv"
	"os"

	"github.com/mccune1224/betrayal/internal/data"
)

// Load in csv contents and append to app struct
func (app *application) ParsePerkCsv(filepath string) error {
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

func GetStatuses(csv [][]string) []data.Status {
	var statuses []data.Status
	for i, entry := range csv {
		// Lil janky but line 0 is empty and last line is notes, should
		// probably just remove those lines from the csv, Too Bad!
		if i == 0 || (i == len(csv)-1) {
			continue
		}
		status := data.Status{
			Name:        entry[1],
			Description: entry[2],
		}
		statuses = append(statuses, status)

	}
	return statuses
}
