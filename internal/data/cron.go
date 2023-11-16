package data

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Metadata is a struct that contains all the metadata for a cron job
type InventoryCronJob struct {
	ID                int    `json:"id"`
	InventoryID       string `json:"inventory_id"`
	InventoryCategory string `json:"inventory_category"`
	InventoryAction   string `json:"inventory_action"`
	InventoryValue    string `json:"inventory_value"`
	StartTime         int64  `json:"start_time"`
	InvokeTime        int64  `json:"invoke_time"`
}

// FIXME: THIS RIGHT IS IS NASTY BUT IDK WHAT ELSE TO DO RN
func (icj *InventoryCronJob) GenerateJobID() string {
	return fmt.Sprintf("%s-%s-%s-%s", icj.InventoryID, icj.InventoryCategory, icj.InventoryAction, icj.InventoryValue)
}

type InventoryCronJobModel struct {
	DB *sqlx.DB
}

func (icjm *InventoryCronJobModel) Insert(icj *InventoryCronJob) error {
	query := `INSERT INTO inventory_cron_jobs (inventory_id, inventory_category, inventory_action, inventory_value, start_time, invoke_time) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	row := icjm.DB.QueryRow(query, icj.InventoryID, icj.InventoryCategory, icj.InventoryAction, icj.InventoryValue, icj.StartTime, icj.InvokeTime)
	err := row.Scan(&icj.ID)
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
	query := `SELECT * FROM inventory_cron_jobs WHERE inventory_category=$1`
	var cronJobs []InventoryCronJob
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

func (icjm *InventoryCronJobModel) DeleteByID(id int) error {
	query := `DELETE FROM inventory_cron_jobs WHERE id=$1`
	_, err := icjm.DB.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
