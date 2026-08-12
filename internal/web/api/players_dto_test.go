package api

import (
	"encoding/json"
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
	"github.com/stretchr/testify/require"
)

func TestPlayerDTOKeepsDiscordSnowflakeAsString(t *testing.T) {
	dto := playerDTOFor(models.Player{ID: 9007199254740993}, "Oracle")
	encoded, err := json.Marshal(dto)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(encoded, &body))
	require.Equal(t, "9007199254740993", body["id"])
}
