package whisper

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const SuspicionChance = 0.05

var (
	ErrSelfTarget       = errors.New("cannot whisper to yourself")
	ErrDeadSender       = errors.New("dead players cannot whisper")
	ErrIncompleteGroup  = errors.New("linked whisper group is incomplete")
	ErrNoRecipients     = errors.New("whisper has no recipients")
	ErrInvalidMessage   = errors.New("whisper message is invalid")
	ErrNoWarningMessage = errors.New("no suspicion messages are available")
)

type PlayerConfessional struct {
	PlayerID  int64
	ChannelID string
	GroupID   string
	Alive     bool
}

type SenderDelivery struct {
	ChannelIDs      []string
	GroupSize       int
	AliveRecipients int
	DeadRecipients  int
}

// ResolveRecipients resolves the target player to the complete linked group.
// An empty GroupID means the target is a singleton. The sender is never added
// implicitly; the caller must reject self-targeting before delivery.
func ResolveRecipients(senderID, targetID int64, confessionals []PlayerConfessional) ([]string, error) {
	if senderID == targetID {
		return nil, ErrSelfTarget
	}
	var target *PlayerConfessional
	for i := range confessionals {
		if confessionals[i].PlayerID == targetID {
			target = &confessionals[i]
			break
		}
	}
	if target == nil || target.ChannelID == "" {
		return nil, ErrNoRecipients
	}

	members := make([]PlayerConfessional, 0, len(confessionals))
	if target.GroupID == "" {
		members = append(members, *target)
	} else {
		for _, conf := range confessionals {
			if conf.GroupID == target.GroupID {
				members = append(members, conf)
			}
		}
		if len(members) < 2 {
			return nil, ErrIncompleteGroup
		}
		// A linked group is complete only when every member has a current
		// confessional. Do this before any Discord send so triplets cannot
		// silently degrade into partial delivery.
		for _, member := range members {
			if member.ChannelID == "" {
				return nil, ErrIncompleteGroup
			}
		}
	}

	sort.Slice(members, func(i, j int) bool { return members[i].PlayerID < members[j].PlayerID })
	channels := make([]string, 0, len(members))
	seen := make(map[string]struct{}, len(members))
	for _, member := range members {
		if member.ChannelID == "" {
			return nil, ErrIncompleteGroup
		}
		if _, ok := seen[member.ChannelID]; ok {
			continue
		}
		seen[member.ChannelID] = struct{}{}
		channels = append(channels, member.ChannelID)
	}
	if len(channels) == 0 {
		return nil, ErrNoRecipients
	}
	return channels, nil
}

// ResolveSenderRecipients resolves the sender's complete linked group and
// excludes dead members and the sender's own confessional from delivery.
func ResolveSenderRecipients(senderID int64, confessionals []PlayerConfessional) ([]string, error) {
	delivery, err := ResolveSenderDelivery(senderID, confessionals)
	if err != nil {
		return nil, err
	}
	return delivery.ChannelIDs, nil
}

func ResolveSenderDelivery(senderID int64, confessionals []PlayerConfessional) (SenderDelivery, error) {
	var sender *PlayerConfessional
	for i := range confessionals {
		if confessionals[i].PlayerID == senderID {
			sender = &confessionals[i]
			break
		}
	}
	if sender == nil || sender.GroupID == "" {
		return SenderDelivery{}, ErrNoRecipients
	}
	if !sender.Alive {
		return SenderDelivery{}, ErrDeadSender
	}

	members := make([]PlayerConfessional, 0, len(confessionals))
	for _, conf := range confessionals {
		if conf.GroupID == sender.GroupID {
			members = append(members, conf)
		}
	}
	if len(members) < 2 {
		return SenderDelivery{}, ErrIncompleteGroup
	}
	sort.Slice(members, func(i, j int) bool { return members[i].PlayerID < members[j].PlayerID })
	delivery := SenderDelivery{ChannelIDs: make([]string, 0, len(members)-1), GroupSize: len(members)}
	seen := make(map[string]struct{}, len(members))
	for _, member := range members {
		if member.PlayerID == senderID {
			if member.ChannelID == "" {
				return SenderDelivery{}, ErrIncompleteGroup
			}
			continue
		}
		if !member.Alive {
			delivery.DeadRecipients++
			continue
		}
		if member.ChannelID == "" {
			return SenderDelivery{}, ErrIncompleteGroup
		}
		delivery.AliveRecipients++
		if _, ok := seen[member.ChannelID]; ok {
			continue
		}
		seen[member.ChannelID] = struct{}{}
		delivery.ChannelIDs = append(delivery.ChannelIDs, member.ChannelID)
	}
	return delivery, nil
}

type Sender interface {
	Send(channelID, content string) error
}

type Roller interface {
	Hit(chance float64) bool
	Intn(n int) int
}

type DeliveryRequest struct {
	SenderChannelID     string
	RecipientChannelIDs []string
	Message             string
	GroupSize           int
	AliveRecipients     int
	DeadRecipients      int
}

type DeliveryResult struct {
	WarningSent                bool
	DeliveredRecipientChannels []string
	SenderStatus               string
}

// Deliver sends the original or doubt-replaced message to all preflighted
// recipients, then sends the sender receipt. The sender receipt is deliberately
// not included in the recipient fan-out and never reveals the original message
// when doubt replaces it.
func Deliver(req DeliveryRequest, sender Sender, roller Roller, warningPool []string) (DeliveryResult, error) {
	if req.SenderChannelID == "" || req.Message == "" {
		return DeliveryResult{}, ErrInvalidMessage
	}
	result := DeliveryResult{DeliveredRecipientChannels: make([]string, 0, len(req.RecipientChannelIDs))}
	groupSize := req.GroupSize
	if groupSize == 0 {
		groupSize = len(req.RecipientChannelIDs) + 1
	}
	groupLabel := relationshipLabel(groupSize)
	warningSent := len(req.RecipientChannelIDs) > 0 && roller != nil && len(warningPool) > 0 && roller.Hit(SuspicionChance)
	primary := fmt.Sprintf("Your %s whispers:\n\n> %s\n\nA message passed quietly through the mirrors.", groupLabel, quoteMessage(req.Message))
	if warningSent {
		warning := warningPool[roller.Intn(len(warningPool))]
		primary = warning
	}
	for _, channelID := range req.RecipientChannelIDs {
		if err := sender.Send(channelID, primary); err != nil {
			return result, fmt.Errorf("send whisper to %s: %w", channelID, err)
		}
		result.DeliveredRecipientChannels = append(result.DeliveredRecipientChannels, channelID)
	}
	result.WarningSent = warningSent

	receipt := fmt.Sprintf("Whisper sent.\n\nYour message found its way to your %s.", groupLabel)
	if req.DeadRecipients > 0 {
		receipt = deadRecipientReceipt(groupLabel, req.AliveRecipients)
	}
	if result.WarningSent && req.DeadRecipients == 0 {
		receipt = "Whisper sent.\n\nSomething blurred between intention and arrival. The mirrors did not carry your words as spoken."
	}
	result.SenderStatus = strings.TrimPrefix(receipt, "Whisper sent.\n\n")
	return result, nil
}

func deadRecipientReceipt(groupLabel string, alive int) string {
	if alive == 0 {
		return fmt.Sprintf("Whisper sent.\n\nThe %s window has shattered. There is no living reflection left to receive your words.", groupLabel)
	}
	if groupLabel == "twin" {
		return "Whisper sent.\n\nA crack has spread across the twin’s mirror. Your words reached no living reflection."
	}
	return "Whisper sent.\n\nA crack runs through the triplet’s mirror. Your words reached the reflections that remain."
}

func quoteMessage(message string) string {
	return strings.ReplaceAll(message, "\n", "\n> ")
}

func relationshipLabel(groupSize int) string {
	switch groupSize {
	case 2:
		return "twin"
	case 3:
		return "triplet"
	default:
		return fmt.Sprintf("group of %d", groupSize)
	}
}
