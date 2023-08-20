package main

import (
	"flag"
	"log"
	"os"

	"github.com/mccune1224/betrayal/internal/data"
	"github.com/spf13/viper"
)

// Flags for CLI app
var (
	file  = flag.String("file", "", "File to read from")
	table = flag.String("table", "", "Table to insert into")
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
	switch *table {
	case "role":
	}
}
