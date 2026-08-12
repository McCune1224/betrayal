package whisper

import (
	"errors"
	"fmt"
	"sort"
)

const SuspicionChance = 0.02

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
	Timestamp           string
}

type DeliveryResult struct {
	WarningSent                bool
	DeliveredRecipientChannels []string
}

// Deliver sends to all preflighted recipients, then sends the sender receipt.
// The sender receipt is deliberately not included in the suspicion fan-out.
func Deliver(req DeliveryRequest, sender Sender, roller Roller, warningPool []string) (DeliveryResult, error) {
	if req.SenderChannelID == "" || len(req.RecipientChannelIDs) == 0 || req.Message == "" {
		return DeliveryResult{}, ErrInvalidMessage
	}
	result := DeliveryResult{DeliveredRecipientChannels: make([]string, 0, len(req.RecipientChannelIDs))}
	primary := fmt.Sprintf("%s\n\n%s", req.Timestamp, req.Message)
	for _, channelID := range req.RecipientChannelIDs {
		if err := sender.Send(channelID, primary); err != nil {
			return result, fmt.Errorf("send whisper to %s: %w", channelID, err)
		}
		result.DeliveredRecipientChannels = append(result.DeliveredRecipientChannels, channelID)
	}

	if roller != nil && len(warningPool) > 0 && roller.Hit(SuspicionChance) {
		warning := warningPool[roller.Intn(len(warningPool))]
		for _, channelID := range result.DeliveredRecipientChannels {
			if err := sender.Send(channelID, warning); err != nil {
				return result, fmt.Errorf("send whisper suspicion to %s: %w", channelID, err)
			}
		}
		result.WarningSent = true
	}

	receipt := fmt.Sprintf("Whisper sent at %s\n\n%s", req.Timestamp, req.Message)
	if err := sender.Send(req.SenderChannelID, receipt); err != nil {
		return result, fmt.Errorf("send whisper receipt: %w", err)
	}
	return result, nil
}
