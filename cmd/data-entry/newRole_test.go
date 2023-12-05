package main

import (
	"encoding/csv"
	"os"
	"testing"
)

func TestBuildNewRoleCSV(t *testing.T) {
	file, err := os.Open("./fat-dumpy/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var b csvBuilder
	_, err = b.BuildNewRoleCSV(records, "GOOD")
	if err != nil {
		t.Fatal(err)
	}
}
