package alliance

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/pkg/data"
)

func (ah *AllianceHandler) CreateAllinaceRequest(allianceName string, requestorID string) error {
	req := &data.AllianceRequest{
		Name:        allianceName,
		RequesterID: requestorID,
	}

	// Check for existing request
	existingRequest, err := ah.m.Alliances.GetRequestByRequesterID(requestorID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}
	if existingRequest != nil {
		return ErrCreateRequestAlreadyExists
	}

	// Check to make sure name doesn't already exists
	existingAlliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}
	if existingAlliance != nil {
		return ErrAllianceAlreadyExists
	}

	// Check to make sure not already in an alliance
	existingMember, err := ah.m.Alliances.GetByMemberID(requestorID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}

	if existingMember != nil {
		return ErrMemberAlreadyExists
	}

	// Create the request
	err = ah.m.Alliances.InsertRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (ah *AllianceHandler) ApproveCreateRequest(requestName string, s *discordgo.Session, e *discordgo.InteractionCreate) (*data.Alliance, error) {
	pendingRequest, err := ah.m.Alliances.GetRequestByName(requestName)
	if err != nil {
		return nil, err
	}

	// make the channel to put alliance in (should always be hidden channel initially with only the requester and admin(s))
	channelName := strings.ReplaceAll(pendingRequest.Name, " ", "-")
	channel, err := discord.CreateChannelWithinCategory(s, e, "alliances", channelName, true)
	if err != nil {
		return nil, err
	}

	newAlliance := &data.Alliance{
		Name:      pendingRequest.Name,
		ChannelID: channel.ID,
		MemberIDs: pq.StringArray{pendingRequest.RequesterID},
	}

	err = ah.m.Alliances.Insert(newAlliance)
	if err != nil {
		s.ChannelDelete(channel.ID)
		return nil, err
	}

	// Delete the request
	err = ah.m.Alliances.DeleteRequest(pendingRequest)
	if err != nil {
		return nil, err
	}

	discord.AddMemberToChannel(s, channel.ID, pendingRequest.RequesterID)

	return newAlliance, nil
}

func (ah *AllianceHandler) DeclineRequest(allianceName string) error {
	request, err := ah.m.Alliances.GetRequestByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRequestNotFound
		}
		return err
	}
	err = ah.m.Alliances.DeleteRequest(request)
	if err != nil {
		return err
	}
	return nil
}
