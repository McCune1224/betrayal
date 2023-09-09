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
	file = flag.String("file", "", "File to read from")
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
	defer db.Close()

	app.models = data.NewModels(db)

	err = app.ParseAnyAbilityCsv(*file)
	if err != nil {
		app.logger.Fatal(err)
	}

	app.InsertAnyAbilities()
}

func (a *application) InsertStatuses(db *sqlx.DB) {
	statusEntry := data.StatusModel{DB: db}
	for _, status := range GetStatuses(a.csv) {
		id, err := statusEntry.Insert(&status)
		if err != nil {
			a.logger.Fatal(err)
		}
		a.logger.Println(id, status.Name)

	}

}

// Catch all for entering InsertRoleJoins into daatbase
func (a *application) InsertRoleJoins(db *sqlx.DB) error {

	roleEntry := data.RoleModel{DB: db}
	abilityEntry := data.AbilityModel{DB: db}
	perkEntry := data.PerkModel{DB: db}
	a.logger.Println(perkEntry, abilityEntry, roleEntry)

	err := a.ParseRoleCsv(*file)
	if err != nil {
		a.logger.Fatal(err)
	}

	roles, err := a.SplitRoles("role")
	if err != nil {
		a.logger.Fatal(err)
	}

	for i, role := range roles {
		if i == 0 {
			continue
		}

		dbRole, err := roleEntry.GetByName(role.Name)
		if err != nil {
			a.logger.Fatal(err)
		}
		if dbRole.ID == -1 {
			a.logger.Fatal("Ability not found")
		}

		abilities, err := role.GetAbilities()
		if err != nil {
			a.logger.Fatal(err)
		}
		perks, err := role.GetPerks()
		if err != nil {
			a.logger.Fatal(err)
		}

		a.logger.Println("JOINING ABILITIES")
		for _, ability := range abilities {
			a.logger.Println(ability.Name)
			dbAbl, err := abilityEntry.GetByName(ability.Name)

			if err != nil {
				a.logger.Fatal(err)
			}

			if dbAbl.ID == -1 {
				a.logger.Fatal("Ability not found")
			}

			err = roleEntry.InsertJoinAbility(dbRole.ID, dbAbl.ID)
			if err != nil {
				a.logger.Fatal(err)
			}
		}
		a.logger.Println("JOINING PERKS")
		for _, perk := range perks {
			a.logger.Println(perk.Name)
			dbPerk, err := perkEntry.GetByName(perk.Name)
			if err != nil {
				a.logger.Fatal(err)
			}
			if dbPerk.ID == -1 {
				a.logger.Fatal("Perk not found")
			}
			err = roleEntry.InsertJoinPerk(dbRole.ID, dbPerk.ID)
			if err != nil {
				a.logger.Fatal(err)
			}

		}
	}
	return nil

}

func (app *application) InsertItems() {
	itemEntry := app.models.Items
	parsedItems, err := GetItems(app.csv)
	if err != nil {
		app.logger.Fatal(err)
	}

	for _, item := range parsedItems {
		fmt.Println("ITEM:", item.Name, item.Rarity)
		id, err := itemEntry.Insert(&item)
		if err != nil {
			app.logger.Fatal(err)
		}
		app.logger.Println(id, item.Name)

	}
}

func (app *application) InsertAnyAbilities() {
	// abilityEntry := app.models.Abilities
	parsedAnyAbilities, err := GetAnyAbilities(app.csv)
	if err != nil {
		app.logger.Fatal(err)
	}
	fmt.Println(len(parsedAnyAbilities))
}
