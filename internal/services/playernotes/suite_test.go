package playernotes_test

import (
	"os"
	"testing"

	"github.com/mccune1224/betrayal/tests/testutil"
)

// TestMain boots the DB-backed suite: loads env, enforces the production guard,
// serializes against other DB suites, and applies migrations once.
func TestMain(m *testing.M) {
	os.Exit(testutil.Bootstrap(m))
}
