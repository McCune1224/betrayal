package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	if err != nil {
		app.logger.Fatal(err)
	}

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
		dbRole := data.Role{
			Name:        role.Name,
			Description: role.Description,
		}
		abilities, err := role.GetAbilities()
		if err != nil {
			app.logger.Fatal(err)
		}
		perks, err := role.GetPerks()
		if err != nil {
			app.logger.Fatal(err)
		}
		fmt.Println(len(abilities))
		fmt.Println(len(perks))
		// fmt.Println("---------------")
		// fmt.Println(role.Name)
		// fmt.Println(role.Description)
		// for _, ability := range abilities {
		// 	fmt.Println(ability.Name)
		// 	fmt.Println(ability.Description)
		// 	fmt.Println(ability.Charges)
		// 	fmt.Println(ability.Categories)
		// 	fmt.Println(ability.AnyAbility)
		// 	fmt.Println("---")
		// }
		// for _, perk := range perks {
		// 	fmt.Println(perk.Name)
		// 	fmt.Println(perk.Effect)
		// 	fmt.Println("---")
		// }
		// fmt.Println("---------------")

		dbRole.Name = role.Name
		dbRole.Description = role.Description
		dbRole.Alignment = "NEUTRAL"
		fmt.Println("|", dbRole.Name, dbRole.Description, "|")
		_, err = roleEntry.Insert(&dbRole)
		if err != nil {
			app.logger.Fatal(err)
		}
		fmt.Println()
	}

}
