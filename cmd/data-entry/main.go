package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
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
	config config
	logger *log.Logger
	csv    [][]string
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

	db, err := pgxpool.New(context.Background(), cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}
	defer db.Close()

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if strings.Contains(file.Name(), "GOOD") {
		alignment := string(models.AlignmentGOOD)
		SyncRolesCsv(db, file, alignment)
	} else if strings.Contains(file.Name(), "EVIL") {
		alignment := string(models.AlignmentEVIL)
		SyncRolesCsv(db, file, alignment)
	} else if strings.Contains(file.Name(), "NEUTRAL") {
		alignment := string(models.AlignmentNEUTRAL)
		SyncRolesCsv(db, file, alignment)
	} else {
		log.Fatal("Invalid alignment")
	}
}

type TempCreateAbilityInfoParams struct {
	models.CreateAbilityInfoParams
	CategoryNames []string
}

func SyncRolesCsv(db *pgxpool.Pool, file *os.File, alignment string) error {
	reader := csv.NewReader(file)
	chunks := [][][]string{}
	currChunk := [][]string{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			chunks = append(chunks, currChunk)
			break
		}
		if err != nil {
			return err
		}
		if record[1] == "" {
			chunks = append(chunks, currChunk)
			currChunk = [][]string{}
		} else {
			currChunk = append(currChunk, record)
		}
	}
	chunks = chunks[1:]

	if len(chunks) < 1 {
		return errors.New("No records found")
	}

	type bulkRoleCreate struct {
		R models.CreateRoleParams
		A []TempCreateAbilityInfoParams
		P []models.CreatePerkInfoParams
	}

	bulkRoleCreateList := []bulkRoleCreate{}

	// TODO: Remove this hardcoded limit after testing
	for i := range chunks {
		roleParams, roleAbilityDetailParams, rolePassiveDetailParams, err := parseRoleChunk(chunks[i])
		if err != nil {
			log.Println("Error Parsing Roles CSV into chunks", err)
			return err
		}

		switch strings.ToUpper(alignment) {
		case string(models.AlignmentGOOD):
			roleParams.Alignment = models.AlignmentGOOD
		case string(models.AlignmentEVIL):
			roleParams.Alignment = models.AlignmentEVIL
		case string(models.AlignmentNEUTRAL):
			roleParams.Alignment = models.AlignmentNEUTRAL
		default:
			log.Println(alignment)
			return errors.New("Invalid alignment")
		}

		bulkEntry := bulkRoleCreate{
			R: roleParams,
			A: roleAbilityDetailParams,
			P: rolePassiveDetailParams,
		}
		bulkRoleCreateList = append(bulkRoleCreateList, bulkEntry)
	}

	q := models.New(db)

	// err := q.NukeRoles(context.Background())
	// if err != nil {
	// 	log.Println("Error Nuking Roles", err)
	// 	return err
	// }

	// NOTE: Need to create the role first before creating the ability/passive, otherwise the ability/passive will be created with the wrong role_id
	// hence why this is in its own loop
	roleIds := pq.Int32Array{}
	for _, roleParams := range bulkRoleCreateList {
		r, err := q.CreateRole(context.Background(), roleParams.R)
		if err != nil {
			log.Println("Error Creating Role", err)
			return err
		}
		roleIds = append(roleIds, r.ID)
	}

	realAbility := models.CreateAbilityInfoParams{}
	for i, roleParams := range bulkRoleCreateList {
		for _, a := range roleParams.A {
			roleID := roleIds[i]

			realAbility.Name = a.Name
			realAbility.Description = a.Description
			realAbility.DefaultCharges = a.DefaultCharges
			realAbility.Rarity = a.Rarity
			realAbility.AnyAbility = a.AnyAbility

			dbAbility, err := q.CreateAbilityInfo(context.Background(), realAbility)

			if err != nil {
				if util.ErrorContains(err, pgerrcode.UniqueViolation) {
					log.Println(a.Name, "already exists")
				} else {
					log.Println(err, roleParams.R.Name, a.Name)
					return err
				}
			}

			for _, categoryName := range a.CategoryNames {
				dbCategory, err := q.GetCategoryByFuzzy(context.Background(), strings.ToUpper(categoryName))
				if err != nil {
					log.Println("Error Getting Category ID", categoryName, err)
				}
				q.CreateAbilityCategoryJoin(context.Background(), models.CreateAbilityCategoryJoinParams{
					AbilityID:  dbAbility.ID,
					CategoryID: dbCategory.ID,
				})
			}

			err = q.CreateRoleAbilityJoin(context.Background(), models.CreateRoleAbilityJoinParams{RoleID: roleID, AbilityID: dbAbility.ID})
			if err != nil {
				log.Println(err, roleParams.R.Name, a.Name)
				return err
			}
		}

		for _, p := range roleParams.P {
			rId := roleIds[i]
			dbPerk, err := q.CreatePerkInfo(context.Background(), p)
			if err != nil {
				if !util.ErrorContains(err, "23505") {
					log.Println(err, roleParams.R.Name, p.Name)
					return err
				}
				// Passive already exists, so just grab it here before proceeding
				dbPerk, err = q.GetPerkInfoByFuzzy(context.Background(), p.Name)
				if err != nil {
					log.Println(err, roleParams.R.Name, p.Name)
					return err
				}
			}
			// insert entry into role_passives_join
			err = q.CreateRolePerkJoin(context.Background(), models.CreateRolePerkJoinParams{RoleID: rId, PerkID: dbPerk.ID})
			if err != nil {
				log.Println(err, roleParams.R.Name, p.Name)
				return err
			}
		}

	}

	return nil
}

func parseAbility(row []string) (TempCreateAbilityInfoParams, error) {
	abilityDetail := TempCreateAbilityInfoParams{}
	abilityDetail.Name = row[1]
	abilityDetail.Description = row[4]

	iCharge := int32(999999)
	if row[2] != "∞" {
		charge, err := strconv.Atoi(row[2])
		if err != nil {
			log.Println("ERR ON", abilityDetail.Name)
			return abilityDetail, err
		}
		iCharge = int32(charge)
	}

	abilityDetail.DefaultCharges = iCharge
	switch row[3] {
	case "*":
		abilityDetail.AnyAbility = true
		// abilityDetail.RoleSpecific = roleName
		switch models.Rarity(strings.TrimSpace(strings.ToUpper(row[6]))) {
		case models.RarityCOMMON:
			abilityDetail.Rarity = models.RarityCOMMON
		case models.RarityUNCOMMON:
			abilityDetail.Rarity = models.RarityUNCOMMON
		case models.RarityRARE:
			abilityDetail.Rarity = models.RarityRARE
		case models.RarityEPIC:
			abilityDetail.Rarity = models.RarityEPIC
		case models.RarityLEGENDARY:
			abilityDetail.Rarity = models.RarityLEGENDARY
		case models.RarityMYTHICAL:
			abilityDetail.Rarity = models.RarityMYTHICAL
		}
	case "^":
		abilityDetail.AnyAbility = true
		// abilityDetail.RoleSpecific = roleName
		abilityDetail.Rarity = models.RarityROLESPECIFIC
	case "":
		abilityDetail.AnyAbility = false
		// abilityDetail.RoleSpecific = roleName
		abilityDetail.Rarity = models.RarityROLESPECIFIC
	default:
		log.Printf("---------------- CANNOT PARSE '%s' AS ANY ABILITY DEFAULTING ROLE_SPECIFIC", row[3])
		abilityDetail.AnyAbility = false
		// abilityDetail.RoleSpecific = roleName
		abilityDetail.Rarity = models.RarityROLESPECIFIC
	}

	abilityDetail.CategoryNames = strings.Split(row[5], "/")
	return abilityDetail, nil
}

func parseRoleChunk(chunk [][]string) (models.CreateRoleParams, []TempCreateAbilityInfoParams, []models.CreatePerkInfoParams, error) {
	roleParams := models.CreateRoleParams{}
	tempRoleAbilityDetailParams := []TempCreateAbilityInfoParams{}
	rolePassiveDetailParams := []models.CreatePerkInfoParams{}
	roleParams.Name = chunk[1][1]

	abParseIndex := 3
	for chunk[abParseIndex][1] != "Perks:" {
		ab, err := parseAbility(chunk[abParseIndex])
		if err != nil {
			return roleParams, tempRoleAbilityDetailParams, rolePassiveDetailParams, err
		}

		tempRoleAbilityDetailParams = append(tempRoleAbilityDetailParams, ab)
		abParseIndex++
	}
	for _, p := range chunk[abParseIndex+1:] {
		createPassive := models.CreatePerkInfoParams{Name: p[1], Description: p[2]}
		rolePassiveDetailParams = append(rolePassiveDetailParams, createPassive)
	}

	return roleParams, tempRoleAbilityDetailParams, rolePassiveDetailParams, nil
}

func SyncItemsCVS(db *pgxpool.Pool, file *os.File) error {
	reader := csv.NewReader(file)
	chunks := [][][]string{}
	currChunk := [][]string{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			chunks = append(chunks, currChunk)
			break
		}
		if err != nil {
			return err
		}
		if record[1] == "" {
			chunks = append(chunks, currChunk)
			currChunk = [][]string{}
		} else {
			currChunk = append(currChunk, record)
		}
	}
}
