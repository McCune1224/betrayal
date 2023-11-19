package scheduler

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/util"
)

func (bs *BetrayalScheduler) ScheduleEffect(effect string, inv *data.Inventory, duration time.Duration, job func()) error {
	log.Printf("Schedule Effect got duration of %d", duration)
	jobData := &data.InventoryCronJob{
		InventoryID: inv.ID,
		PlayerID:    inv.DiscordID,
		ChannelID:   inv.UserPinChannel,
		Category:    "effect",
		ActionType:  "remove",
		Value:       effect,
		StartTime:   time.Now().Unix(),
		InvokeTime:  time.Now().Add(duration).Unix(),
	}
	log.Printf("Scheduled job to trigger at %s", util.GetEstTimeStampFromDuration(duration))
	err := bs.InsertJob(jobData, job)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BetrayalScheduler) QueueScheduleJobs(session *discordgo.Session) error {
	jobs, err := bs.m.InventoryCronJobs.GetByCategory("effect")
	if err != nil {
		log.Println(err)
		return err
	}

	for _, job := range jobs {
		if time.Now().Unix() > job.InvokeTime {
			err := bs.m.InventoryCronJobs.DeletebyJobID(job.MakeJobID())
			if err != nil {
				log.Printf("Failed to delete job %s, %s", job.MakeJobID(), err.Error())
			}
			continue
		}
		dur := time.Duration(job.InvokeTime - time.Now().Unix())
		log.Printf("TODO: schedule job for %s to trigger at %s", job.MakeJobID(), util.GetEstTimeStampFromDuration(dur))
	}
	return nil
}
