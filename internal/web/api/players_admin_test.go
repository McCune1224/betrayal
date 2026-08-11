package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePlayerIDPreservesDiscordSnowflake(t *testing.T) {
	var raw json.RawMessage = []byte(`"123456789012345678"`)
	id, err := parsePlayerID(raw)

	require.NoError(t, err)
	require.Equal(t, int64(123456789012345678), id)
}

func TestParsePlayerIDAcceptsLegacyJSONNumber(t *testing.T) {
	var raw json.RawMessage = []byte(`123456789012345678`)
	id, err := parsePlayerID(raw)

	require.NoError(t, err)
	require.Equal(t, int64(123456789012345678), id)
}

func TestParsePlayerIDRejectsInvalidValue(t *testing.T) {
	for _, raw := range []json.RawMessage{nil, []byte(`""`), []byte(`0`), []byte(`"not-a-snowflake"`)} {
		_, err := parsePlayerID(raw)
		require.Error(t, err)
	}
}
