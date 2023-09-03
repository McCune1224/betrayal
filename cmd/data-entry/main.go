package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/spf13/viper"
)

// Flags for CLI app
var (
	entryType = flag.String("type", "", "Type of entry to create")
	file      = flag.String("file", "", "File to read from")
)

type config struct {
	database struct {
		dsn string
	}
}

type application struct {
	config   config
	models   data.Models
	logger   *log.Logger
	modelMap map[string]data.Models
	csv      [][]string
}

// Really just here pull in json data and populate the databse with it.
func main() {

	flag.Parse()
	if *file == "" {
		log.Fatal("file is required")
	}

	var cfg config
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	cfg.database.dsn = viper.GetString("DATABASE_URL")
	if cfg.database.dsn == "" {
		app.logger.Fatal("DATABASE_URL is required")
	}

	db, err := sqlx.Connect("postgres", cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}

	roleEntry := data.RoleModel{DB: db}
	abilityEntry := data.AbilityModel{DB: db}
	perkEntry := data.PerkModel{DB: db}
	fmt.Println(perkEntry, abilityEntry, roleEntry)

	err = app.ParseCsv(*file)
	if err != nil {
		logger.Fatal(err)
	}

	roles, err := app.SplitRoles("role")
	if err != nil {
		logger.Fatal(err)
	}

	for i, role := range roles {
		if i == 0 {
			continue
		}
		fmt.Println(role.Name)
		abilities, err := role.GetAbilities()
		if err != nil {
			app.logger.Fatal(err)
		}
		for _, ability := range abilities {
			fmt.Println(ability)
			abilityID, err := abilityEntry.Insert(&ability)
			if err != nil {
				if !strings.Contains(err.Error(), "duplicate key value violate") {
					app.logger.Fatal(err)
				}
			}
			fmt.Println(abilityID)
		}

		perks, err := role.GetPerks()
		if err != nil {
			app.logger.Fatal(err)
		}

		for _, perk := range perks {
			fmt.Println(perk)
			fmt.Println(perk.Name, perk.Description)
			perkID, err := perkEntry.Insert(&perk)
			if err != nil {
				if !strings.Contains(err.Error(), "duplicate key value violate") {
					app.logger.Fatal(err)
				}
			}
			fmt.Println(perkID)
		}

	}
}
