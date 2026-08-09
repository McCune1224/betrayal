package handlers

import (
	"errors"
	"strings"
	"testing"
)

func TestSyncDatabaseErrorHintIdentifiesMissingMigration(t *testing.T) {
	err := errors.New(`ERROR: relation "sync_source" does not exist (SQLSTATE 42P01)`)

	hint := syncDatabaseErrorHint(err)

	if !strings.Contains(hint, "000030") {
		t.Fatalf("hint %q should identify migration 000030", hint)
	}
	if !strings.Contains(hint, "/admin/migrations") {
		t.Fatalf("hint %q should link to the migrations page", hint)
	}
}

func TestSyncDatabaseErrorHintLeavesOtherErrorsUnchanged(t *testing.T) {
	err := errors.New("connection refused")

	if got := syncDatabaseErrorHint(err); got != "" {
		t.Fatalf("unexpected hint for unrelated error: %q", got)
	}
}
