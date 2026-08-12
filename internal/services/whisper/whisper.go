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
	ErrIncompleteGroup  = errors.New("linked whisper group is incomplete")
	ErrNoRecipients     = errors.New("whisper has no recipients")
	ErrInvalidMessage   = errors.New("whisper message is invalid")
	ErrNoWarningMessage = errors.New("no suspicion messages are available")
)

type PlayerConfessional struct {
	PlayerID  int64
	ChannelID string
	GroupID   string
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
// excludes the sender's own confessional from delivery.
func ResolveSenderRecipients(senderID int64, confessionals []PlayerConfessional) ([]string, error) {
	var sender *PlayerConfessional
	for i := range confessionals {
		if confessionals[i].PlayerID == senderID {
			sender = &confessionals[i]
			break
		}
	}
	if sender == nil || sender.GroupID == "" {
		return nil, ErrNoRecipients
	}

	members := make([]PlayerConfessional, 0, len(confessionals))
	for _, conf := range confessionals {
		if conf.GroupID == sender.GroupID {
			members = append(members, conf)
		}
	}
	if len(members) < 2 {
		return nil, ErrIncompleteGroup
	}
	sort.Slice(members, func(i, j int) bool { return members[i].PlayerID < members[j].PlayerID })
	channels := make([]string, 0, len(members)-1)
	seen := make(map[string]struct{}, len(members))
	for _, member := range members {
		if member.ChannelID == "" {
			return nil, ErrIncompleteGroup
		}
		if member.PlayerID == senderID {
			continue
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
}

type DeliveryResult struct {
	WarningSent                bool
	DeliveredRecipientChannels []string
}

// Deliver sends the original or doubt-replaced message to all preflighted
// recipients, then sends the sender receipt. The sender receipt is deliberately
// not included in the recipient fan-out and never reveals the original message
// when doubt replaces it.
func Deliver(req DeliveryRequest, sender Sender, roller Roller, warningPool []string) (DeliveryResult, error) {
	if req.SenderChannelID == "" || len(req.RecipientChannelIDs) == 0 || req.Message == "" {
		return DeliveryResult{}, ErrInvalidMessage
	}
	result := DeliveryResult{DeliveredRecipientChannels: make([]string, 0, len(req.RecipientChannelIDs))}
	groupLabel := relationshipLabel(len(req.RecipientChannelIDs) + 1)
	warningSent := roller != nil && len(warningPool) > 0 && roller.Hit(SuspicionChance)
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
	if result.WarningSent {
		receipt = "Whisper sent.\n\nSomething blurred between intention and arrival. The mirrors did not carry your words as spoken."
	}
	if err := sender.Send(req.SenderChannelID, receipt); err != nil {
		return result, fmt.Errorf("send whisper receipt: %w", err)
	}
	return result, nil
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
