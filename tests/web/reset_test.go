package web_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResetPageShowsGuidedNewGameFlow(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/admin/reset")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := client.body(resp)
	require.Contains(t, body, "START FRESH")
	require.Contains(t, body, "CLEAR GAME &amp; RELOAD CSV DATA")
	require.Contains(t, body, "RESET BETRAYAL GAME")
	require.Contains(t, body, "Discord channel setup")
}

func TestResetRequiresExplicitConfirmation(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.do(http.MethodPost, "/admin/reset", url.Values{
		"confirm":    {"not the phrase"},
		"understand": {"on"},
	}, nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
