package main

import "github.com/mccune1224/betrayal/internal/data"

func (*csvBuilder) BuildStatusCSV(csv [][]string) ([]data.Status, error) {
	var statuses []data.Status

	for i, entry := range csv {
		if i == 0 || i == 1 || i == len(csv)-1 {
			continue
		}

		status := data.Status{
			Name:        entry[1],
			Description: entry[2],
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}
