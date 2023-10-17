package data

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrRecordAlreadyExists = errors.New("record already exists")
)

// All interested models to be used in the application
type Models struct {
	Roles          RoleModel
	Insults        InsultModel
	Abilities      AbilityModel
	Perks          PerkModel
	Statuses       StatusModel
	Items          ItemModel
	Players        PlayerModel
	Inventories    InventoryModel
	Whitelists     WhitelistModel
	Actions        ActionModel
	FunnelChannels FunnelChannelModel
	Duels          DuelModel
}

// NewModels creates a new instance of the Models struct and attaches the database connection to it.
func NewModels(db *sqlx.DB) Models {
	ModelHandlers := Models{
		Roles:          RoleModel{DB: db},
		Insults:        InsultModel{DB: db},
		Abilities:      AbilityModel{DB: db},
		Perks:          PerkModel{DB: db},
		Statuses:       StatusModel{DB: db},
		Items:          ItemModel{DB: db},
		Players:        PlayerModel{DB: db},
		Inventories:    InventoryModel{DB: db},
		Whitelists:     WhitelistModel{DB: db},
		Actions:        ActionModel{DB: db},
		FunnelChannels: FunnelChannelModel{DB: db},
		Duels:          DuelModel{DB: db},
	}
	return ModelHandlers
}

// Helper to automatically generate the PSQL query for the keys of a struct
func SqlGenKeys(model interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(model))
	var query []string
	for i := 0; i < v.NumField(); i++ {
		columnName := v.Type().Field(i).Tag.Get("db")

		switch t := v.Field(i).Interface().(type) {
		case string:
			if t != "" {
				query = append(query, fmt.Sprintf("%s=$%s", columnName, columnName))
			}
		case int:
			if t != 0 {
				query = append(query, fmt.Sprintf("%s=$%s", columnName, columnName))
			}
		default:
			if reflect.ValueOf(t).Kind() == reflect.Ptr {
				if reflect.Indirect(reflect.ValueOf(t)) != reflect.ValueOf(nil) {
					query = append(query, fmt.Sprintf("%s=$%s", columnName, columnName))
				}
			} else {
				if reflect.ValueOf(t) != reflect.ValueOf(nil) {
					query = append(query, fmt.Sprintf("%s=$%s", columnName, columnName))
				}
			}
		}
	}
	return strings.Join(query, ", ")
}

// Function to generate an SQLX insert statement with placeholders
// i.e for query INSERT INTO table (column1, column2...) VALUES (:column1, :column2...)
// Will check if default field value is not nil/empty before adding to query
func PSQLGeneratedInsert(model interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(model))
	var columns []string
	var tally int

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("db")
		switch t := v.Field(i).Interface().(type) {
		case string:
			if t != "" {
				columns = append(columns, tag)
				tally += 1
			}
		case int:
			if t != 0 {
				columns = append(columns, tag)
				tally += 1
			}
		case int64:
			if t != 0 {
				columns = append(columns, tag)
				tally += 1
			}
		default:
			if reflect.ValueOf(t).Kind() == reflect.Ptr {
				if reflect.Indirect(reflect.ValueOf(t)) != reflect.ValueOf(nil) {
					columns = append(columns, tag)
					tally += 1
				}
			} else {
				if reflect.ValueOf(t) != reflect.ValueOf(nil) {
					columns = append(columns, tag)
					tally += 1
				}
			}
		}
	}
	left := ""
	right := ""
	for _, v := range columns {
		left = left + v + ", "
		right = right + ":" + v + ", "
	}
	left = strings.TrimSuffix(left, ", ")
	right = strings.TrimSuffix(right, ", ")

	left = "(" + left + ")"
	right = "VALUES (" + right + ")"

	return fmt.Sprintf("%s %s", left, right)
}

func PSQLGeneratedUpdate(model interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(model))
	var query []string
	for i := 0; i < v.NumField(); i++ {
		columnName := v.Type().Field(i).Tag.Get("db")

		switch t := v.Field(i).Interface().(type) {
		case string:
			if t != "" {
				query = append(query, fmt.Sprintf("%s = :%s", columnName, columnName))
			}
		case int:
			if t != 0 {
				query = append(query, fmt.Sprintf("%s = :%s", columnName, columnName))
			}
		case pq.StringArray:
			if len(t) != 0 {
				query = append(query, fmt.Sprintf("%s = :%s", columnName, columnName))
			}
		}
	}

	ret := strings.Join(query, ", ")
	return ret
}

// Idk where tf to put this thing so its getting thrown in here for now
func RemoveSliceItem[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}
