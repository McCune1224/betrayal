package cron

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
)

func (bs *BetrayalScheduler) ScheduleEffect(effect string, inv *data.Inventory, duration time.Duration, job interface{}) error {
	jobData := data.InventoryCronJob{
		InventoryID:       inv.UserPinChannel,
		InventoryCategory: "effect",
		InventoryAction:   "remove",
		InventoryValue:    effect,
		StartTime:         time.Now().Unix(),
		InvokeTime:        time.Now().Add(duration).Unix(),
	}
	jobID := jobData.GenerateJobID()
	jobData.JobID = jobID
	err := bs.dbJobs.InventoryCronJobs.Insert(&jobData)
	if err != nil {
		return err
	}
	err = bs.UpsertJob(jobID, duration, job)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BetrayalScheduler) QueueScheduleJobs(session *discordgo.Session) error {
	jobs, err := bs.dbJobs.InventoryCronJobs.GetByCategory("effect")
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Found %d jobs to queue", len(jobs))
	for _, job := range jobs {
		inventory, err := bs.dbJobs.Inventories.GetByDiscordID(job.InventoryID)
		if err != nil {
			log.Printf("Failed to queue job \n%v\n%s", job, err.Error())
			continue
		}
		jobID := job.GenerateJobID()
		err = bs.UpsertJob(jobID, time.Duration(job.InvokeTime-job.StartTime)*time.Second, func() {
			err = bs.ScheduleEffect(job.InventoryValue, inventory, time.Duration(job.InvokeTime), func() {
				for k, v := range inventory.Effects {
					if strings.EqualFold(job.InventoryValue, v) {
						inventory.Effects = append(inventory.Effects[:k], inventory.Effects[k+1:]...)
						err = bs.dbJobs.Inventories.UpdateEffects(inventory)
						if err != nil {
							log.Println(err)
							msg := &discordgo.MessageEmbed{
								Title:       "Failed to remove effect",
								Description: fmt.Sprintf("Failed to remove effect %s from inventory", job.InventoryValue),
								Fields: []*discordgo.MessageEmbedField{
									{
										Name:  "Error",
										Value: err.Error(),
									},
								},
							}
							_, err := session.ChannelMessageSendEmbed(job.InventoryID, msg)
							if err != nil {
								log.Println(err)
								return
							}
						}
					}
				}

				msg := discordgo.MessageEmbed{
					Title:       "Effect Expired",
					Description: fmt.Sprintf("Effect %s has expired", job.InventoryValue),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Target Inventory",
							Value: fmt.Sprintf("<@%s>", inventory.DiscordID),
						},
						{
							Name:  "Target Channel",
							Value: fmt.Sprintf("<#%s>", inventory.UserPinChannel),
						},
					},
					Color:     0x00ff00,
					Timestamp: time.Now().Format(time.RFC3339),
				}
				_, err := session.ChannelMessageSendEmbed(job.InventoryID, &msg)
				if err != nil {
					log.Println(err)
				}
			})
		})
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("Queued job %s to execute at %s", jobID, time.Unix(job.InvokeTime, 0).Format(time.RFC3339))
	}
	return nil
}
