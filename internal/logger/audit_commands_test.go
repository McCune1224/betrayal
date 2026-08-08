package logger

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestExtractCommandArgumentsSkipsMalformedOptions(t *testing.T) {
	got := ExtractCommandArguments(nil, []*discordgo.ApplicationCommandInteractionDataOption{
		nil,
		{Name: "broken-user", Type: discordgo.ApplicationCommandOptionUser},
		{Name: "broken-string", Type: discordgo.ApplicationCommandOptionString, Value: 42},
	})

	if _, ok := got["broken-user"]; ok {
		t.Fatalf("malformed user option should not be logged: %#v", got)
	}
	if _, ok := got["broken-string"]; ok {
		t.Fatalf("malformed string option should not be logged: %#v", got)
	}
}

func TestIsAdminMemberMatchesRoleNamesToMemberRoleIDs(t *testing.T) {
	member := &discordgo.Member{Roles: []string{"role-host"}}
	roles := []*discordgo.Role{{ID: "role-host", Name: "Host"}}

	if !isAdminMember(member, roles) {
		t.Fatal("member with Host role should be recognized as an admin")
	}
}

func TestExtractCommandArgumentsPreservesResolvedOptionIDsSafely(t *testing.T) {
	got := ExtractCommandArguments(nil, []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "user", Type: discordgo.ApplicationCommandOptionUser, Value: "user-1"},
		{Name: "channel", Type: discordgo.ApplicationCommandOptionChannel, Value: "channel-1"},
		{Name: "role", Type: discordgo.ApplicationCommandOptionRole, Value: "role-1"},
		{Name: "mentionable", Type: discordgo.ApplicationCommandOptionMentionable, Value: "mentionable-1"},
	})

	want := map[string]interface{}{
		"user":        map[string]interface{}{"id": "user-1"},
		"channel":     map[string]interface{}{"id": "channel-1"},
		"role":        map[string]interface{}{"id": "role-1"},
		"mentionable": map[string]interface{}{"id": "mentionable-1", "type": "mentionable"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolved option audit output = %#v, want %#v", got, want)
	}
}
