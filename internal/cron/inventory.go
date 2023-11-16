package cron

import (
	"log"
	"time"

	"github.com/mccune1224/betrayal/internal/data"
)

func (bs *BetrayalScheduler) ScheduleEffect(effect string, inv *data.Inventory, duration time.Duration, job interface{}) error {
	jobData := data.InventoryCronJob{
		InventoryID:       inv.DiscordID,
		InventoryCategory: "effect",
		InventoryAction:   "remove",
		InventoryValue:    effect,
		StartTime:         time.Now().Unix(),
		InvokeTime:        time.Now().Add(duration).Unix(),
	}

	jobID := jobData.GenerateJobID()
	err := bs.dbJobs.Insert(&jobData)
	if err != nil {
		return err
	}
	err = bs.UpsertJob(jobID, duration, job)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BetrayalScheduler) QueueScheduleJobs() error {
	jobs, err := bs.dbJobs.GetByCategory("effect")
	if err != nil {
		return err
	}
	for _, job := range jobs {
		jobID := job.GenerateJobID()
		err := bs.UpsertJob(jobID, time.Duration(job.InvokeTime-job.StartTime)*time.Second, func() {
			log.Println("FOUND JOB", jobID)
			// print the jobData pulled from the DB to the console
			log.Printf("%+v\n", job)
			log.Println("--------------------")
			log.Println("--------------------")
		})
		if err != nil {
			return err
		}
	}

	return nil
}
