package alliance

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
)

var (
	ErrAlreadyOwner    = errors.New("user is already owner of an alliance")
	ErrAlreadyExists   = errors.New("alliance already exists")
	ErrRequestNotFound = errors.New("alliance request not found")
)

func (ah *AllianceHandler) CreateAllinaceRequest(allianceName string, requestorID string) error {
	req := &data.AllianceRequest{
		Name:        allianceName,
		RequesterID: requestorID,
	}
	// Check to make sure name doesn't already exists
	existingAllianceName, _ := ah.m.Alliances.GetByName(allianceName)
	if existingAllianceName.Name != "" {
		return ErrAlreadyExists
	}

	// Check to make sure owner doesn't already have an alliance
	existingOwnned, _ := ah.m.Alliances.GetByOwnerID(requestorID)
	if existingOwnned.Name != "" {
		return ErrAlreadyExists
	}

	// Check to make sure owner isn't already in an alliance
	existingMember, _ := ah.m.Alliances.GetByMemberID(requestorID)
	if existingMember.Name != "" {
		return ErrAlreadyExists
	}

	// Create the request
	err := ah.m.Alliances.InsertRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (ah *AllianceHandler) ApproveAllianceRequest(playerID string, s *discordgo.Session) (*data.Alliance, error) {
	pendingRequest, err := ah.m.Alliances.GetRequestByRequesterID(playerID)
	if err != nil {
		return nil, err
	}

	// make the channel to put alliance in
	channel, err := s.GuildChannelCreate(s.State.Application.GuildID, pendingRequest.Name, discordgo.ChannelTypeGuildText)
	if err != nil {
		return nil, err
	}

	newAlliance := &data.Alliance{
		Name:      pendingRequest.Name,
		OwnerID:   pendingRequest.RequesterID,
		ChannelID: channel.ID,
		MemberIDs: pq.StringArray{},
	}
	err = ah.m.Alliances.Insert(newAlliance)
	if err != nil {
		return nil, err
	}

	// Delete the request
	err = ah.m.Alliances.DeleteRequest(pendingRequest)
	if err != nil {
		return nil, err
	}

	return newAlliance, nil
}
