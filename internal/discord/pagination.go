package discord

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	PaginationPageSize = 10
)

type PaginationData struct {
	Items       []any
	CurrentPage int
	PageSize    int
	Title       string
	Description string
	FormatFunc  func(item any) *discordgo.MessageEmbedField
	Color       int
}

type paginationState struct {
	data      *PaginationData
	expiresAt time.Time
}

var (
	paginationStates = make(map[string]*paginationState)
	paginationMutex  sync.RWMutex
)

const (
	paginationTimeout = 5 * time.Minute
)

// GetPageSize returns the pagination page size constant
func GetPageSize() int {
	return PaginationPageSize
}
func CreatePaginatedEmbed(data *PaginationData) *discordgo.MessageEmbed {
	totalPages := (len(data.Items) + data.PageSize - 1) / data.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Ensure current page is within bounds
	if data.CurrentPage < 0 {
		data.CurrentPage = 0
	}
	if data.CurrentPage >= totalPages {
		data.CurrentPage = totalPages - 1
	}

	// Get items for this page
	startIdx := data.CurrentPage * data.PageSize
	endIdx := startIdx + data.PageSize
	if endIdx > len(data.Items) {
		endIdx = len(data.Items)
	}

	var fields []*discordgo.MessageEmbedField
	for i := startIdx; i < endIdx; i++ {
		fields = append(fields, data.FormatFunc(data.Items[i]))
	}

	embed := &discordgo.MessageEmbed{
		Title:       data.Title,
		Description: data.Description,
		Fields:      fields,
		Color:       data.Color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d", data.CurrentPage+1, totalPages),
		},
	}

	return embed
}

// GetPaginationComponents returns the action row with pagination buttons
func GetPaginationComponents(paginationID string, data *PaginationData) []discordgo.MessageComponent {
	totalPages := (len(data.Items) + data.PageSize - 1) / data.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	canPrevious := data.CurrentPage > 0
	canNext := data.CurrentPage < totalPages-1

	prevBtn := discordgo.Button{
		Label:    "← Previous",
		Style:    discordgo.PrimaryButton,
		CustomID: paginationID + ":prev",
		Disabled: !canPrevious,
	}

	nextBtn := discordgo.Button{
		Label:    "Next →",
		Style:    discordgo.PrimaryButton,
		CustomID: paginationID + ":next",
		Disabled: !canNext,
	}

	doneBtn := discordgo.Button{
		Label:    "Done",
		Style:    discordgo.SecondaryButton,
		CustomID: paginationID + ":done",
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{prevBtn, nextBtn, doneBtn},
		},
	}
}

// StorePaginationState stores pagination state with expiration
func StorePaginationState(paginationID string, data *PaginationData) {
	paginationMutex.Lock()
	defer paginationMutex.Unlock()

	paginationStates[paginationID] = &paginationState{
		data:      data,
		expiresAt: time.Now().Add(paginationTimeout),
	}

	// Clean up expired states
	go cleanupExpiredStates()
}

// GetPaginationState retrieves pagination state
func GetPaginationState(paginationID string) *PaginationData {
	paginationMutex.RLock()
	defer paginationMutex.RUnlock()

	state, exists := paginationStates[paginationID]
	if !exists {
		return nil
	}

	if time.Now().After(state.expiresAt) {
		return nil
	}

	return state.data
}

// UpdatePaginationState updates the pagination state
func UpdatePaginationState(paginationID string, data *PaginationData) {
	paginationMutex.Lock()
	defer paginationMutex.Unlock()

	if state, exists := paginationStates[paginationID]; exists {
		state.data = data
		state.expiresAt = time.Now().Add(paginationTimeout)
	}
}

// DeletePaginationState removes pagination state
func DeletePaginationState(paginationID string) {
	paginationMutex.Lock()
	defer paginationMutex.Unlock()

	delete(paginationStates, paginationID)
}

// cleanupExpiredStates removes expired pagination states
func cleanupExpiredStates() {
	paginationMutex.Lock()
	defer paginationMutex.Unlock()

	now := time.Now()
	for id, state := range paginationStates {
		if now.After(state.expiresAt) {
			delete(paginationStates, id)
		}
	}
}
