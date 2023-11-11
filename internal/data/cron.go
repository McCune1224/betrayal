package data

import "github.com/jmoiron/sqlx"

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

type InventoryCronJobModel struct {
	DB *sqlx.DB
}

func (icjm *InventoryCronJobModel) Insert(icj *InventoryCronJob) error {
	query := `INSERT INTO inventory_cron_jobs ` + PSQLGeneratedInsert(icj) + ` RETURNING id`
	_, err := icjm.DB.NamedExec(query, icj)
	if err != nil {
		return err
	}
	return nil
}

func (icjm *InventoryCronJobModel) GetByInventoryID(inventoryID string) ([]*InventoryCronJob, error) {
	query := `SELECT * FROM inventory_cron_jobs WHERE inventory_id=$1`
	var inventoryCronJobs []*InventoryCronJob
	err := icjm.DB.Select(&inventoryCronJobs, query, inventoryID)
	if err != nil {
		return nil, err
	}
	return inventoryCronJobs, nil
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
