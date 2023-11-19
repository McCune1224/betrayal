package data

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Metadata is a struct that contains all the metadata for a cron job
type InventoryCronJob struct {
	ID int `db:"id"`
	// foreign key to Inventory ID
	InventoryID int64  `db:"inventory_id"`
	ChannelID   string `db:"channel_id"`
	PlayerID    string `db:"player_id"`
	Category    string `db:"category"`
	ActionType  string `db:"action_type"`
	Value       string `db:"value"`
	StartTime   int64  `db:"start_time"`
	InvokeTime  int64  `db:"invoke_time"`
}

// MakeJobID returns a unique job ID for the cron job based on the inventory ID, category, action type, and value
func (icj *InventoryCronJob) MakeJobID() string {
	return fmt.Sprintf("%s-%s-%s-%s", icj.PlayerID, icj.Category, icj.ActionType, icj.Value)
}

type InventoryCronJobModel struct {
	DB *sqlx.DB
}

func (icjm *InventoryCronJobModel) Insert(icj *InventoryCronJob) error {
	query := `INSERT INTO inventory_cron_jobs ` + PSQLGeneratedInsert(icj) + ` RETURNING id`
	_, err := icjm.DB.NamedExec(query, &icj)
	if err != nil {
		return err
	}
	return nil
}

func (icjm *InventoryCronJobModel) GetByJobID(jobID string) error {
	properties := strings.Split(jobID, "-")
	query := `SELECT * FROM inventory_cron_jobs WHERE player_id=$1 AND category=$2 AND action_type=$3 AND value=$4`
	var cronJob InventoryCronJob
	err := icjm.DB.Get(&cronJob, query, properties[0], properties[1], properties[2], properties[3])
	if err != nil {
		return err
	}
	return nil
}

func (icjm *InventoryCronJobModel) GetByInventoryID(inventoryID string) ([]InventoryCronJob, error) {
	query := `SELECT * FROM inventory_cron_jobs WHERE inventory_id=$1`
	var inventoryCronJobs []InventoryCronJob
	err := icjm.DB.Select(&inventoryCronJobs, query, inventoryID)
	if err != nil {
		return nil, err
	}
	return inventoryCronJobs, nil
}

func (cjm *InventoryCronJobModel) GetByCategory(category string) ([]InventoryCronJob, error) {
	query := `SELECT * FROM inventory_cron_jobs WHERE category=$1`
	cronJobs := []InventoryCronJob{}
	err := cjm.DB.Select(&cronJobs, query, category)
	if err != nil {
		return nil, err
	}
	return cronJobs, nil
}

func (cjm *InventoryCronJobModel) GetAll() (*InventoryCronJob, error) {
	query := `SELECT * FROM inventory_cron_jobs`
	var cronJob InventoryCronJob
	err := cjm.DB.Get(&cronJob, query)
	if err != nil {
		return nil, err
	}
	return &cronJob, nil
}

func (icjm *InventoryCronJobModel) ExtendInvokeTime(time int64) error {
	query := `UPDATE inventory_cron_jobs SET invoke_time=$1`
	_, err := icjm.DB.Exec(query, time)
	if err != nil {
		return err
	}
	return nil
}

func (icjm *InventoryCronJobModel) DeleteByInventoryID(inventoryID string) error {
	query := `DELETE FROM inventory_cron_jobs WHERE inventory_id=$1`
	_, err := icjm.DB.Exec(query, inventoryID)
	if err != nil {
		return err
	}
	return nil
}

func (icjm *InventoryCronJobModel) DeletebyJobID(jobID string) error {
	properties := strings.Split(jobID, "-")
	query := `DELETE FROM inventory_cron_jobs WHERE player_id=$1 AND category=$2 AND action_type=$3 AND value=$4`
	_, err := icjm.DB.Exec(query, properties[0], properties[1], properties[2], properties[3])
	if err != nil {
		return err
	}
	log.Println("Deleted JobID: ", jobID)
	return nil
}

func (icjm *InventoryCronJobModel) DeleteByID(id int) error {
	query := `DELETE FROM inventory_cron_jobs WHERE id=$1`
	_, err := icjm.DB.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
