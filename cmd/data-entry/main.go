package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
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

	cfg.database.dsn = os.Getenv("DATABASE_POOLER_URL")
	if cfg.database.dsn == "" {
		app.logger.Fatal("DATABASE_URL is required")
	}

	db, err := pgxpool.New(context.Background(), cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}
	defer db.Close()

	wg := sync.WaitGroup{}
	wg.Add(4)

	dbCtx := context.Background()

	lazy := []struct {
		URL       string
		alignment models.Alignment
	}{
		{URL: os.Getenv("GOOD_ROLES_CSV"), alignment: models.AlignmentGOOD},
		{URL: os.Getenv("EVIL_ROLES_CSV"), alignment: models.AlignmentEVIL},
		{URL: os.Getenv("NEUTRAL_ROLES_CSV"), alignment: models.AlignmentNEUTRAL},
	}

	go func() {
		start := time.Now()
		csvUrl := lazy[0].URL
		httpClient := &http.Client{}
		resp, err := httpClient.Get(csvUrl)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		SyncRolesCsv(dbCtx, db, strings.NewReader(string(body)), string(lazy[0].alignment))
		fmt.Println("~~ Good Roles Done %s ~~", time.Since(start))
		wg.Done()
	}()

	go func() {
		start := time.Now()
		csvUrl := lazy[1].URL
		httpClient := &http.Client{}
		resp, err := httpClient.Get(csvUrl)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		SyncRolesCsv(dbCtx, db, strings.NewReader(string(body)), string(lazy[1].alignment))
		fmt.Println("~~ Evil Roles Done %s ~~", time.Since(start))
		wg.Done()
	}()

	go func() {
		start := time.Now()
		csvUrl := lazy[2].URL
		httpClient := &http.Client{}
		resp, err := httpClient.Get(csvUrl)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		SyncRolesCsv(dbCtx, db, strings.NewReader(string(body)), string(lazy[2].alignment))
		fmt.Println("~~ Neutral Roles Done %s ~~", time.Since(start))
		wg.Done()
	}()

	go func() {
		start := time.Now()
		item_CSV_URL := os.Getenv("ITEM_CSV")
		httpClient := &http.Client{}
		resp, err := httpClient.Get(item_CSV_URL)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		SyncItemsCsv(dbCtx, db, strings.NewReader(string(body)))
		fmt.Println("~~ Items Done %s ~~", time.Since(start))
		wg.Done()
	}()

	wg.Wait()
	// file, err := os.Open(*fileName)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if strings.Contains(file.Name(), "GOOD") {
	// 	alignment := string(models.AlignmentGOOD)
	// 	SyncRolesCsv(db, file, alignment)
	// } else if strings.Contains(file.Name(), "EVIL") {
	// 	alignment := string(models.AlignmentEVIL)
	// 	SyncRolesCsv(db, file, alignment)
	// } else if strings.Contains(file.Name(), "NEUTRAL") {
	// 	alignment := string(models.AlignmentNEUTRAL)
	// 	SyncRolesCsv(db, file, alignment)
	// } else {
	// 	log.Fatal("Invalid alignment")
	// }
	// file.Close()

}

type TempCreateAbilityInfoParams struct {
	models.CreateAbilityInfoParams
	CategoryNames []string
}

func SyncRolesCsv(ctx context.Context, db *pgxpool.Pool, r io.Reader, alignment string) error {
	reader := csv.NewReader(r)
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
	roleParams.Description = chunk[1][2]

	abParseIndex := 3
	for chunk[abParseIndex][1] != "Passives:" {
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

func SyncItemsCsv(ctx context.Context, db *pgxpool.Pool, r io.Reader) error {
	reader := csv.NewReader(r)
	csv, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for i, entry := range csv {
		if i == 0 || i == 1 || len(csv) == i-1 {
			continue
		}

		item := models.CreateItemParams{
			// Rarity:      entry[1],
			Name:        entry[2],
			Description: entry[5],
		}
		switch strings.ToUpper(entry[1]) {
		case "COMMON":
			item.Rarity = models.RarityCOMMON
		case "UNCOMMON":
			item.Rarity = models.RarityUNCOMMON
		case "RARE":
			item.Rarity = models.RarityRARE
		case "EPIC":
			item.Rarity = models.RarityEPIC
		case "LEGENDARY":
			item.Rarity = models.RarityLEGENDARY
		case "MYTHICAL":
			item.Rarity = models.RarityMYTHICAL
		case "UNIQUE":
			item.Rarity = models.RarityUNIQUE
		}

		// FIXME: This is stinky and very specific to the item csv, Too Bad!
		strCost := entry[3]
		if strCost == "X" {
			item.Cost = 0
		} else {
			cost, err := strconv.ParseInt(strCost, 10, 64)
			if err != nil {
				return err
			}
			item.Cost = int32(cost)
		}

		categories := entry[4]
		parsedCategories := strings.Split(categories, "/")
		for i, category := range parsedCategories {
			parsedCategories[i] = strings.TrimSpace(category)
		}

		q := models.New(db)

		dbItem, err := q.CreateItem(context.Background(), item)
		if err != nil {
			log.Println("Error Creating Item", err)
			return err
		}

		for _, category := range parsedCategories {
			dbCategory, err := q.GetCategoryByFuzzy(context.Background(), strings.ToUpper(category))
			if err != nil {
				log.Println("Error Getting Category ID", category, err)
			}
			q.CreateItemCategoryJoin(context.Background(), models.CreateItemCategoryJoinParams{
				ItemID:     dbItem.ID,
				CategoryID: dbCategory.ID,
			})
		}
	}
	return nil
}
