package cycle

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain boots the suite: loads env, enforces the production guard,
// serializes against other DB suites, and applies migrations once.
func TestMain(m *testing.M) {
	os.Exit(testutil.Bootstrap(m))
}

// stubTransport answers every Discord REST call with an empty channel list
// (no "alliances" category) unless overridden per-test.
type stubTransport struct{}

func (stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("[]")),
	}, nil
}

func newStubSession(t *testing.T) *discordgo.Session {
	t.Helper()
	s, err := discordgo.New("Bot test-token")
	require.NoError(t, err)
	s.Client = &http.Client{Transport: stubTransport{}}
	return s
}

// TestGetCycleChannelIDs_MissingAlliancesCategory is the regression test for
// the 2026-08-15 production panic: a guild with no "alliances" category must
// not nil-pointer dereference at cycle.go:221, and the command must still
// gather the channels it can find (vote + action) rather than fail wholesale.
func TestGetCycleChannelIDs_MissingAlliancesCategory(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	ctx := context.Background()
	require.NoError(t, q.UpsertActionChannel(ctx, "111"))
	require.NoError(t, q.UpsertVoteChannel(ctx, "222"))

	c := &Cycle{dbPool: pool}
	channels, err := c.getCycleChannelIDs(newStubSession(t), &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{GuildID: "123456789"}})

	require.NoError(t, err)
	// Vote + action only; alliance channels are skipped, not fatal.
	assert.Equal(t, []string{"222", "111"}, channels)
}

// TestGetCycleChannelIDs_UnconfiguredActionChannel: with no action_channel row,
// the error must name the missing config instead of raw pgx "no rows in result set".
func TestGetCycleChannelIDs_UnconfiguredActionChannel(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)

	c := &Cycle{dbPool: pool}
	_, err := c.getCycleChannelIDs(newStubSession(t), &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{GuildID: "123456789"}})

	require.Error(t, err)
	assert.ErrorContains(t, err, "action channel")
}
