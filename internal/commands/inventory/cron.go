package inventory

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

func (i *Inventory) ScheduleStatus(ctx ken.SubCommandContext, status *data.Status, inv *data.Inventory, duration int64) error {
	// only certain statuses can be scheduled so return early if the status is not one of them

	job := data.InventoryCronJob{
		InventoryID:       inv.DiscordID,
		InventoryCategory: "status",
		InventoryAction:   "remove",
		InventoryValue:    status.Name,
		StartTime:         time.Now().Unix(),
		InvokeTime:        time.Now().Unix() + duration,
	}

	err := i.models.InventoryCronJobs.Insert(&job)
	if err != nil {
		return err
	}
	jobID := fmt.Sprintf("%s-%s-%s-%s", job.InventoryID, job.InventoryCategory, job.InventoryAction, job.InventoryValue)
	err = i.scheduler.UpsertJob(jobID, time.Duration(duration)*time.Second, func() {
		for k, v := range inv.Statuses {
			if strings.EqualFold(v, job.InventoryValue) {
				inv.Statuses = append(inv.Statuses[:k], inv.Statuses[k+1:]...)
				err = i.models.Inventories.UpdateStatuses(inv)
				if err != nil {
					log.Println(err)
					return
				}
				err = i.updateInventoryMessage(ctx, inv)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func (i *Inventory) ScheduleEffect(ctx ken.SubCommandContext, effect string, inv *data.Inventory, duration time.Duration) error {
	jobData := data.InventoryCronJob{
		InventoryID:       inv.DiscordID,
		InventoryCategory: "effect",
		InventoryAction:   "remove",
		InventoryValue:    effect,
		StartTime:         time.Now().Unix(),
		InvokeTime:        time.Now().Add(duration).Unix(),
	}

	jobID := jobData.GenerateJobID()

	i.scheduler.UpsertJob(jobID, duration, func() {
		for k, v := range inv.Effects {
			if strings.EqualFold(v, effect) {
				inv.Effects = append(inv.Effects[:k], inv.Effects[k+1:]...)
				err := i.models.Inventories.UpdateEffects(inv)
				if err != nil {
					log.Println(err)
					return
				}
				err = i.updateInventoryMessage(ctx, inv)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
		return
	})
	return nil
}
