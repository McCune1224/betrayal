package alliance

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
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

	// Check for existing request
	existingRequest, err := ah.m.Alliances.GetRequestByRequesterID(requestorID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}
	if existingRequest != nil {
		return ErrAlreadyExists
	}

	// Check to make sure name doesn't already exists
	existingAlliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}
	if existingAlliance != nil {
		return ErrAlreadyExists
	}

	// Check to make sure owner doesn't already have an alliance
	existingOwnned, err := ah.m.Alliances.GetByOwnerID(requestorID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}

	if existingOwnned != nil {
		return ErrAlreadyOwner
	}

	// Check to make sure owner isn't already in an alliance
	existingMember, err := ah.m.Alliances.GetByMemberID(requestorID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return err
	}

	if existingMember != nil {
		return ErrAlreadyExists
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
	log.Println("PENDING REQUEST: ", pendingRequest)

	// make the channel to put alliance in
	channelName := strings.ReplaceAll(pendingRequest.Name, " ", "-")
	channel, err := discord.AddChannelWithinCategory(s, e, "alliances", channelName)
	if err != nil {
		return nil, err
	}
	log.Println("CHANNEL: ", channel)

	newAlliance := &data.Alliance{
		Name:      pendingRequest.Name,
		OwnerID:   pendingRequest.RequesterID,
		ChannelID: channel.ID,
		MemberIDs: pq.StringArray{},
	}

	err = ah.m.Alliances.Insert(newAlliance)
	if err != nil {
		s.ChannelDelete(channel.ID)
		return nil, err
	}
	log.Println("ALLIANCE INSERTED: ", newAlliance)

	// Delete the request
	err = ah.m.Alliances.DeleteRequest(pendingRequest)
	if err != nil {
		return nil, err
	}

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
