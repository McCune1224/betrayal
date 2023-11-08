package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
)

// Flags for CLI app
var (
	fileName = flag.String("file", "", "File to read from")
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

type csvBuilder struct{}

// Really just here pull in json data and populate the databse with it.
func main() {
	flag.Parse()
	// if *file == "" {
	// 	log.Fatal("file is required")
	// }
	//
	var cfg config
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	cfg.database.dsn = os.Getenv("DATABASE_URL")
	if cfg.database.dsn == "" {
		app.logger.Fatal("DATABASE_URL is required")
	}

	db, err := sqlx.Connect("postgres", cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}
	defer db.Close()
	app.models = data.NewModels(db)

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// CreateRolls(file, app, "EVIL")
	// CreateAnyAbilities(file, app)
	// CreateItems(file, app)
	CreateStatuses(file, app)
}

func CreateRolls(file *os.File, app *application, alignment string) {
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	var b csvBuilder

	csvRoles, err := b.BuildRoleCSV(records)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(len(csvRoles))

	for i, csvRole := range csvRoles {
		if i == 0 {
			continue
		}
		fmt.Println("----------------------------------------------------------")
		role, err := csvRole.ToDBEntry(alignment)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Inserting role", role.Name)
		abilities, err := csvRole.GetAbilities()
		if err != nil {
			log.Fatal(err)
		}

		perks, err := csvRole.GetPerks()
		if err != nil {
			log.Fatal(err)
		}

		rID, err := app.models.Roles.Insert(&role)
		if err != nil {
			log.Println("FAILED TO INSERT ROLE", role.Name)
			log.Fatal(err)
		}

		for _, ability := range abilities {
			aID, err := app.models.Abilities.Insert(&ability)
			if err != nil {
				log.Println("FAILED TO INSERT ABILITY", ability.Name)
				log.Fatal(err)
			}
			err = app.models.Roles.InsertJoinAbility(rID, aID)
			if err != nil {
				log.Fatal(err)
			}
		}

		for _, perk := range perks {
			pID, err := app.models.Perks.Insert(&perk)
			if err != nil {
				log.Println("FAILED TO INSERT PERK", perk.Name)
				log.Fatal(err)
			}
			err = app.models.Roles.InsertJoinPerk(rID, pID)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// func CreateItems(file *os.File, app *application) {
// 	csvReader := csv.NewReader(file)
// 	records, err := csvReader.ReadAll()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var b csvBuilder
// 	items, err := b.BuildItemCSV(records)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	for _, item := range items {
// 		_, err := app.models.Items.Insert(&item)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// }

func CreateAnyAbilities(file *os.File, app *application) {
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	var b csvBuilder
	anyAbilities, err := b.BuildAnyAbilityCSV(records)
	if err != nil {
		log.Fatal(err)
	}

	for _, aa := range anyAbilities {
		err := app.models.Abilities.InsertAnyAbility(&aa)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func CreateItems(file *os.File, app *application) {
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	var b csvBuilder
	items, err := b.BuildItemCSV(records)
	if err != nil {
		log.Fatal(err)
	}
	tx := app.models.Items.DB.MustBegin()
	for _, item := range items {
		_, err := app.models.Items.Insert(&item)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	// commit transaction here
}

func CreateStatuses(file *os.File, app *application) {
	csvReader := csv.NewReader(file)
	record, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	var b csvBuilder
	statuses, err := b.BuildStatusCSV(record)
	if err != nil {
		log.Fatal(err)
	}

	for _, status := range statuses {
		_, err := app.models.Statuses.Insert(&status)
		if err != nil {
			log.Fatal(err)
		}
	}
}
