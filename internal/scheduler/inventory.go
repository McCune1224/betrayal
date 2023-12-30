package scheduler

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/pkg/data"
)

func (bs *BetrayalScheduler) ScheduleEffect(effect string, inv *data.Inventory, duration time.Duration, s *discordgo.Session, expired ...bool) error {
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
				errMsg := errorScheduleMessage(jobData)
				s.ChannelMessageSendEmbed(jobData.ChannelID, errMsg)
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
	if len(expired) > 0 && expired[0] {
		err := bs.InvokeJob(jobData.MakeJobID(), remove)
		if err != nil {
			return err
		}
	}

	err := bs.InsertJob(jobData, remove)
	if err != nil {
		return err
	}

	return nil
}

func (bs *BetrayalScheduler) ScheduleStatus(status string, inv *data.Inventory, duration time.Duration, s *discordgo.Session, expired ...bool) error {
	log.Printf("Schedule Status got duration of %d", duration)
	now := time.Now()
	jobData := &data.InventoryCronJob{
		InventoryID: inv.ID,
		PlayerID:    inv.DiscordID,
		ChannelID:   inv.UserPinChannel,
		Category:    "status",
		ActionType:  "remove",
		Value:       status,
		StartTime:   now.Unix(),
		InvokeTime:  now.Add(duration).Unix(),
	}
	log.Printf("Scheduled %s to trigger at %s", jobData.MakeJobID(), util.GetEstTimeStampFromDuration(duration))
	remove := func() {
		inv, err := bs.m.Inventories.GetByDiscordID(inv.DiscordID)
		if err != nil {
			log.Println(err)
			s.ChannelMessageSend(jobData.ChannelID, "Failed to find inventory for status expiration")
			return
		}
		handler := inventory.InitInventoryHandler(bs.m, inv)
		best, err := handler.RemoveStatus(jobData.Value)
		if err != nil {
			if errors.Is(err, inventory.ErrStatusNotFound) {
				errMsg := errorScheduleMessage(jobData)
				s.ChannelMessageSendEmbed(jobData.ChannelID, errMsg)
				return
			}
			log.Println(err)
			s.ChannelMessageSend(jobData.ChannelID, fmt.Sprintf("Failed to remove timed status %s", jobData.Value))
			return
		}
		start := util.GetEstTimeStampFromDuration(time.Until(time.Unix(jobData.StartTime, 0)))
		msg := discordgo.MessageEmbed{
			Title:       "Status Expired",
			Description: fmt.Sprintf("Status %s has expired", best),
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
	if len(expired) > 0 && expired[0] {
		err := bs.InvokeJob(jobData.MakeJobID(), remove)
		if err != nil {
			return err
		}
		return nil
	}

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

		inv, err := bs.m.Inventories.GetByDiscordID(job.PlayerID)
		if err != nil {
			log.Println("JOB ERR,", err.Error())
		}

		invokeTime := time.Unix(job.InvokeTime, 0)
		isExpired := invokeTime.Before(time.Now())
		invokeDuration := time.Until(invokeTime)

		log.Println("Scheduling job for", util.GetEstTimeStampFromDuration(invokeDuration))

		switch job.Category {
		case "effect":
			err = bs.ScheduleEffect(job.Value, inv, invokeDuration, session, isExpired)
			if err != nil {
				log.Println(err)
				continue
			}
		case "status":
			err = bs.ScheduleStatus(job.Value, inv, invokeDuration, session, isExpired)
			if err != nil {
				log.Println(err)
				continue
			}
		}

	}
	return nil
}

func errorScheduleMessage(jobData *data.InventoryCronJob) *discordgo.MessageEmbed {
	msg := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Failed to remove %s %s %s", discord.EmojiError, jobData.Category, jobData.Value, discord.EmojiError),
		Description: "Unable to remove due to an error (likely item already removed)",
		Color:       discord.ColorThemeRed,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return &msg
}
