package main

import (
	"flag"
	"log"
	"os"

	"github.com/mccune1224/betrayal/internal/data"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Flags for CLI app
var (
	file = flag.String("file", "", "File to read from")
)

type config struct {
	database struct {
		dns string
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

	var cfg config

	flag.Parse()
	if *file == "" {
		log.Fatal("file is required")
	}
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	app := &application{
		config: cfg,
		logger: logger,
	}

	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	cfg.database.dns = viper.GetString("DATABASE_URL")
	if cfg.database.dns == "" {
		app.logger.Fatal("DATABASE_URL is required")
	}
	db, err := gorm.Open(postgres.Open(cfg.database.dns), &gorm.Config{})
	if err != nil {
		log.Fatal("error opening database,", err)
	}
	roleEntry := data.RoleModel{DB: db}
	// abilityEntry := data.AbilityModel{DB: db}
	// perkEntry := data.PerkModel{DB: db}

	//Cascade delet tables and recreate them

	err = app.ParseCsv("./fat-dumpy/good_roles.csv")
	if err != nil {
		logger.Fatal(err)
	}
	roles, err := app.SplitRoles("role")
	if err != nil {
		logger.Fatal(err)
	}

	// make tables just in case

	db.AutoMigrate(
		&data.Role{},
		&data.Ability{},
		&data.Perk{},
		&data.Category{},
	)
	for _, role := range roles {

		abilities, err := role.SanitizeAbilities()
		perks, err := role.SanitizePerks()
		role := data.Role{
			Name:        role.Name,
			Description: role.Description,
			Abilities:   abilities,
			Perks:       perks,
			Alignment:   "GOOD",
		}
		err = roleEntry.DB.Create(&role).Error

		if err != nil {
			logger.Println(err)
		}
	}

}
