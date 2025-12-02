package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

type DataSyncJob struct {
	Name      string
	URL       string
	Alignment string
	SyncFunc  func(context.Context, *pgxpool.Pool, io.Reader, string) error
}

// logInfo prints an info message with cyan color
func logInfo(format string, args ...interface{}) {
	fmt.Printf("%sℹ %s%s\n", colorCyan, fmt.Sprintf(format, args...), colorReset)
}

// logSuccess prints a success message with green color
func logSuccess(format string, args ...interface{}) {
	fmt.Printf("%s✓ %s%s\n", colorGreen, fmt.Sprintf(format, args...), colorReset)
}

// logError prints an error message with red color
func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s✗ Error: %s%s\n", colorRed, fmt.Sprintf(format, args...), colorReset)
}

// logWarning prints a warning message with yellow color
func logWarning(format string, args ...interface{}) {
	fmt.Printf("%s⚠ %s%s\n", colorYellow, fmt.Sprintf(format, args...), colorReset)
}

// logTask prints a task header
func logTask(title string) {
	fmt.Printf("\n%s%s%s\n", colorBold+colorBlue, title, colorReset)
	fmt.Println(strings.Repeat("─", len(title)))
}

// Syncs roles and items from Google Sheets CSV exports to the database
func main() {
	logTask("Betrayal Data Entry Tool - Syncing from Google Sheets")

	// Load environment
	dsn := os.Getenv("DATABASE_POOLER_URL")
	if dsn == "" {
		logError("DATABASE_POOLER_URL environment variable not set")
		fmt.Fprintf(os.Stderr, "Set it with: export DATABASE_POOLER_URL='postgresql://user:pass@host:port/db'\n")
		os.Exit(1)
	}

	logInfo("Connecting to database...")
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logError("Failed to create database pool: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		logError("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	logSuccess("Connected to database")

	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Define data sync jobs
	jobs := []DataSyncJob{
		{
			Name:      "Good Roles",
			URL:       os.Getenv("GOOD_ROLES_CSV"),
			Alignment: string(models.AlignmentGOOD),
			SyncFunc:  SyncRolesCsv,
		},
		{
			Name:      "Evil Roles",
			URL:       os.Getenv("EVIL_ROLES_CSV"),
			Alignment: string(models.AlignmentEVIL),
			SyncFunc:  SyncRolesCsv,
		},
		{
			Name:      "Neutral Roles",
			URL:       os.Getenv("NEUTRAL_ROLES_CSV"),
			Alignment: string(models.AlignmentNEUTRAL),
			SyncFunc:  SyncRolesCsv,
		},
		{
			Name:      "Items",
			URL:       os.Getenv("ITEM_CSV"),
			Alignment: "",
			SyncFunc:  SyncItemsCsv,
		},
	}

	// Validate all URLs are set
	logInfo("Validating required CSV URLs...")
	for _, job := range jobs {
		if job.URL == "" {
			logWarning("Missing CSV URL for %s - skipping", job.Name)
			continue
		}
	}

	// Execute sync jobs in parallel
	logInfo("Starting data synchronization (parallel execution)...")
	wg.Add(len(jobs))

	for _, job := range jobs {
		go executeDataSync(ctx, db, job, &wg)
	}

	wg.Wait()
	logSuccess("All data synchronization tasks completed!")
}

// executeDataSync runs a single data sync job with error handling and timing
func executeDataSync(ctx context.Context, db *pgxpool.Pool, job DataSyncJob, wg *sync.WaitGroup) {
	defer wg.Done()

	if job.URL == "" {
		logWarning("Skipping %s - no URL provided", job.Name)
		return
	}

	start := time.Now()
	logInfo("Starting sync for %s...", job.Name)

	// Fetch CSV from URL
	resp, err := http.Get(job.URL)
	if err != nil {
		logError("Failed to fetch %s: %v", job.Name, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logError("Failed to read response body for %s: %v", job.Name, err)
		return
	}

	// Execute sync function
	alignmentStr := ""
	if job.Alignment != "" {
		alignmentStr = string(job.Alignment)
	}

	err = job.SyncFunc(ctx, db, strings.NewReader(string(body)), alignmentStr)
	if err != nil {
		logError("Failed to sync %s: %v", job.Name, err)
		return
	}

	duration := time.Since(start)
	logSuccess("Completed %s in %s", job.Name, duration.String())
}

type TempCreateAbilityInfoParams struct {
	models.CreateAbilityInfoParams
	CategoryNames []string
}

func SyncRolesCsv(ctx context.Context, db *pgxpool.Pool, r io.Reader, alignment string) error {
	logInfo("Parsing CSV data...")
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
			logError("Failed to read CSV: %v", err)
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
		logWarning("No role records found in CSV")
		return errors.New("no records found")
	}

	logInfo("Found %d roles to sync", len(chunks))

	type bulkRoleCreate struct {
		R models.CreateRoleParams
		A []TempCreateAbilityInfoParams
		P []models.CreatePerkInfoParams
	}

	bulkRoleCreateList := []bulkRoleCreate{}

	// Parse all roles from chunks
	for i := range chunks {
		roleParams, roleAbilityDetailParams, rolePassiveDetailParams, err := parseRoleChunk(chunks[i])
		if err != nil {
			logError("Failed to parse role chunk %d: %v", i, err)
			return err
		}

		// Set alignment
		switch strings.ToUpper(alignment) {
		case string(models.AlignmentGOOD):
			roleParams.Alignment = models.AlignmentGOOD
		case string(models.AlignmentEVIL):
			roleParams.Alignment = models.AlignmentEVIL
		case string(models.AlignmentNEUTRAL):
			roleParams.Alignment = models.AlignmentNEUTRAL
		default:
			logError("Invalid alignment: %s", alignment)
			return errors.New("invalid alignment")
		}

		bulkEntry := bulkRoleCreate{
			R: roleParams,
			A: roleAbilityDetailParams,
			P: rolePassiveDetailParams,
		}
		bulkRoleCreateList = append(bulkRoleCreateList, bulkEntry)
	}

	q := models.New(db)

	// Create all roles first
	logInfo("Creating %d roles...", len(bulkRoleCreateList))
	roleIds := pq.Int32Array{}
	for i, roleParams := range bulkRoleCreateList {
		r, err := q.CreateRole(context.Background(), roleParams.R)
		if err != nil {
			logError("Failed to create role '%s': %v", roleParams.R.Name, err)
			return err
		}
		roleIds = append(roleIds, r.ID)
		if (i+1)%10 == 0 {
			logInfo("  Created %d/%d roles", i+1, len(bulkRoleCreateList))
		}
	}
	logSuccess("Created all %d roles", len(roleIds))

	// Create abilities and perks for each role
	logInfo("Creating abilities and perks...")
	abilityCount := 0
	perkCount := 0

	for i, roleParams := range bulkRoleCreateList {
		roleID := roleIds[i]

		// Create abilities
		for _, a := range roleParams.A {
			realAbility := models.CreateAbilityInfoParams{
				Name:           a.Name,
				Description:    a.Description,
				DefaultCharges: a.DefaultCharges,
				Rarity:         a.Rarity,
				AnyAbility:     a.AnyAbility,
			}

			dbAbility, err := q.CreateAbilityInfo(context.Background(), realAbility)

			if err != nil {
				if util.ErrorContains(err, pgerrcode.UniqueViolation) {
					logWarning("Ability '%s' already exists, linking to role", a.Name)
					// Try to get existing ability
					dbAbility, err = q.GetAbilityInfoByFuzzy(context.Background(), a.Name)
					if err != nil {
						logError("Failed to find existing ability '%s': %v", a.Name, err)
						continue
					}
				} else {
					logError("Failed to create ability '%s' for role '%s': %v", a.Name, roleParams.R.Name, err)
					return err
				}
			}
			abilityCount++

			// Link to categories
			for _, categoryName := range a.CategoryNames {
				dbCategory, err := q.GetCategoryByFuzzy(context.Background(), strings.ToUpper(categoryName))
				if err != nil {
					logWarning("Could not find category '%s' for ability '%s'", categoryName, a.Name)
					continue
				}
				q.CreateAbilityCategoryJoin(context.Background(), models.CreateAbilityCategoryJoinParams{
					AbilityID:  dbAbility.ID,
					CategoryID: dbCategory.ID,
				})
			}

			// Link ability to role
			err = q.CreateRoleAbilityJoin(context.Background(), models.CreateRoleAbilityJoinParams{RoleID: roleID, AbilityID: dbAbility.ID})
			if err != nil {
				logError("Failed to link ability '%s' to role '%s': %v", a.Name, roleParams.R.Name, err)
				return err
			}
		}

		// Create perks
		for _, p := range roleParams.P {
			dbPerk, err := q.CreatePerkInfo(context.Background(), p)
			if err != nil {
				if !util.ErrorContains(err, "23505") {
					logError("Failed to create perk '%s' for role '%s': %v", p.Name, roleParams.R.Name, err)
					return err
				}
				// Perk already exists, get it
				dbPerk, err = q.GetPerkInfoByFuzzy(context.Background(), p.Name)
				if err != nil {
					logError("Failed to find existing perk '%s': %v", p.Name, err)
					return err
				}
			}
			perkCount++

			// Link perk to role
			err = q.CreateRolePerkJoin(context.Background(), models.CreateRolePerkJoinParams{RoleID: roleID, PerkID: dbPerk.ID})
			if err != nil {
				logError("Failed to link perk '%s' to role '%s': %v", p.Name, roleParams.R.Name, err)
				return err
			}
		}
	}

	logSuccess("Created %d abilities and %d perks", abilityCount, perkCount)
	return nil
}

func parseAbility(row []string) (TempCreateAbilityInfoParams, error) {
	abilityDetail := TempCreateAbilityInfoParams{}
	abilityDetail.Name = row[1]
	abilityDetail.Description = row[4]

	// Parse charges
	iCharge := int32(999999)
	if row[2] != "∞" {
		charge, err := strconv.Atoi(row[2])
		if err != nil {
			logError("Invalid charge count '%s' for ability '%s'", row[2], abilityDetail.Name)
			return abilityDetail, err
		}
		iCharge = int32(charge)
	}

	abilityDetail.DefaultCharges = iCharge

	// Parse rarity and type
	switch row[3] {
	case "*":
		abilityDetail.AnyAbility = true
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
		abilityDetail.Rarity = models.RarityROLESPECIFIC
	case "":
		abilityDetail.AnyAbility = false
		abilityDetail.Rarity = models.RarityROLESPECIFIC
	default:
		logWarning("Unknown ability type '%s', defaulting to ROLE_SPECIFIC", row[3])
		abilityDetail.AnyAbility = false
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

func SyncItemsCsv(ctx context.Context, db *pgxpool.Pool, r io.Reader, _ string) error {
	logInfo("Parsing CSV data...")
	reader := csv.NewReader(r)
	csv, err := reader.ReadAll()
	if err != nil {
		logError("Failed to read CSV: %v", err)
		return err
	}

	// Skip header rows and count valid items
	validItems := 0
	for i := range csv {
		if i == 0 || i == 1 || len(csv) == i-1 {
			continue
		}
		validItems++
	}

	logInfo("Found %d items to sync", validItems)

	q := models.New(db)
	createdCount := 0

	for i, entry := range csv {
		// Skip header rows
		if i == 0 || i == 1 || len(csv) == i-1 {
			continue
		}

		item := models.CreateItemParams{
			Name:        entry[2],
			Description: entry[5],
		}

		// Parse rarity
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
		default:
			logWarning("Unknown rarity '%s' for item '%s', skipping", entry[1], item.Name)
			continue
		}

		// Parse cost
		strCost := entry[3]
		if strCost == "X" {
			item.Cost = 0
		} else {
			cost, err := strconv.ParseInt(strCost, 10, 64)
			if err != nil {
				logError("Invalid cost '%s' for item '%s': %v", strCost, item.Name, err)
				continue
			}
			item.Cost = int32(cost)
		}

		// Create item
		dbItem, err := q.CreateItem(context.Background(), item)
		if err != nil {
			logError("Failed to create item '%s': %v", item.Name, err)
			continue
		}
		createdCount++

		// Link to categories
		categories := entry[4]
		parsedCategories := strings.Split(categories, "/")
		for j, category := range parsedCategories {
			parsedCategories[j] = strings.TrimSpace(category)
		}

		for _, category := range parsedCategories {
			dbCategory, err := q.GetCategoryByFuzzy(context.Background(), strings.ToUpper(category))
			if err != nil {
				logWarning("Could not find category '%s' for item '%s'", category, item.Name)
				continue
			}
			q.CreateItemCategoryJoin(context.Background(), models.CreateItemCategoryJoinParams{
				ItemID:     dbItem.ID,
				CategoryID: dbCategory.ID,
			})
		}

		if (createdCount)%10 == 0 {
			logInfo("  Created %d/%d items", createdCount, validItems)
		}
	}

	logSuccess("Created %d items", createdCount)
	return nil
}
