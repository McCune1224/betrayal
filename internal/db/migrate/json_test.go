package dbmigrate_test

import (
	"encoding/json"
	"testing"

	dbmigrate "github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/stretchr/testify/require"
)

func TestMigrationJSONUsesStableLowercaseAPIFields(t *testing.T) {
	encoded, err := json.Marshal(dbmigrate.Migration{Version: 7, Name: "sync_source", Applied: true, Dirty: false})
	require.NoError(t, err)
	require.JSONEq(t, `{"version":7,"name":"sync_source","applied":true,"dirty":false}`, string(encoded))
}
