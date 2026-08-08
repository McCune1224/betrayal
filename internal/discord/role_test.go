package discord

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGetAdminRoleUsersReportsRoleFetchFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v9/guilds/guild-1/roles" {
			t.Fatalf("request path = %q", r.URL.Path)
		}
		http.Error(w, "discord unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	oldEndpoint := discordgo.EndpointGuilds
	discordgo.EndpointGuilds = server.URL + "/api/v9/guilds/"
	defer func() { discordgo.EndpointGuilds = oldEndpoint }()

	session, err := discordgo.New("Bot test-token")
	if err != nil {
		t.Fatal(err)
	}
	event := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "guild-1",
		Member:  &discordgo.Member{User: &discordgo.User{ID: "user-1"}, Roles: []string{"role-1"}},
	}}

	users, err := GetAdminRoleUsers(session, event, "Host")
	if err == nil {
		t.Fatal("GetAdminRoleUsers error = nil, want role-fetch error")
	}
	if users != nil {
		t.Fatalf("users = %#v, want nil on role-fetch failure", users)
	}
}

var errRoleLookup = errors.New("role lookup failed")

func TestResolveAdminRoleReportsLookupFailure(t *testing.T) {
	member := &discordgo.Member{Roles: []string{"role-1"}}
	wantErr := errRoleLookup
	_, err := resolveAdminRole(member, func() ([]*discordgo.Role, error) {
		return nil, wantErr
	})
	if err != wantErr {
		t.Fatalf("resolveAdminRole error = %v, want %v", err, wantErr)
	}
}
