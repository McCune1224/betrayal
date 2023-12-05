package scheduler

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
)

func (bs *BetrayalScheduler) ScheduleEffect(effect string, inv *data.Inventory, duration time.Duration, s *discordgo.Session, update ...bool) error {
	log.Printf("Schedule Effect got duration of %d", duration)
	now := time.Now()
	jobData := &data.InventoryCronJob{
		InventoryID: inv.ID,
		PlayerID:    inv.DiscordID,
		ChannelID:   inv.UserPinChannel,
		Category:    "effect",
		ActionType:  "remove",
		Value:       effect,
		StartTime:   now.Unix(),
		InvokeTime:  now.Add(duration).Unix(),
	}
	log.Printf("Scheduled %s to trigger at %s", jobData.MakeJobID(), util.GetEstTimeStampFromDuration(duration))

	remove := func() {
		inv, err := bs.m.Inventories.GetByDiscordID(inv.DiscordID)
		if err != nil {
			log.Println(err)
			s.ChannelMessageSend(jobData.ChannelID, "Failed to find inventory for effect expiration")
			return
		}
		handler := inventory.InitInventoryHandler(bs.m, inv)
		best, err := handler.RemoveEffect(jobData.Value)
		if err != nil {
			if errors.Is(err, inventory.ErrEffectNotFound) {
				s.ChannelMessageSend(jobData.ChannelID, fmt.Sprintf("Effect %s not found", jobData.Value))
				return
			}
			log.Println(err)
			s.ChannelMessageSend(jobData.ChannelID, fmt.Sprintf("Failed to remove timed effect %s", jobData.Value))
			return
		}
		start := util.GetEstTimeStampFromDuration(time.Until(time.Unix(jobData.StartTime, 0)))
		msg := discordgo.MessageEmbed{
			Title:       "Effect Expired",
			Description: fmt.Sprintf("Effect %s has expired", best),
			Fields: []*discordgo.MessageEmbedField{
				{
					Value: fmt.Sprintf("Timer started at %s", start),
				},
				{
					Value: fmt.Sprintf("Timer ended at %s", util.GetEstTimeStamp()),
				},
			},
			Color:     discord.ColorThemeOrange,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err = s.ChannelMessageSendEmbed(jobData.ChannelID, &msg)
		if err != nil {
			log.Println(err)
		}
	}

	// Cases where we need to update the job instead of inserting a new one (e.g. bot restart)
	err := bs.InsertJob(jobData, remove)
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
		invokeTime := time.Unix(job.InvokeTime, 0)
		if invokeTime.Before(time.Now()) {
			log.Println("TODO: invoke job that is already past invoke time", job.MakeJobID())
			continue
		}
		inv, err := bs.m.Inventories.GetByDiscordID(job.PlayerID)
		if err != nil {
			log.Println("JOB ERR,", err.Error())
		}
		invokeDuration := time.Until(invokeTime)

		log.Println("Scheduling job for", util.GetEstTimeStampFromDuration(invokeDuration))
		err = bs.ScheduleEffect(job.Value, inv, invokeDuration, session, true)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return nil
}
