package logger

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

type auditSinkFake struct{ audits []CommandAudit }

func (f *auditSinkFake) LogCommand(audit CommandAudit) { f.audits = append(f.audits, audit) }

func TestCommandAuditLifecycleRecordsOneSuccessfulFinalAuditWithDuration(t *testing.T) {
	sink := new(auditSinkFake)
	clock := &fakeAuditClock{now: time.Unix(100, 0)}
	lifecycle := newCommandAuditLifecycle(sink, clock.Now)
	start := lifecycle.Start()
	clock.now = start.Add(25 * time.Millisecond)

	lifecycle.Finish(CommandAudit{CommandName: "roll"}, start, nil)

	if len(sink.audits) != 1 {
		t.Fatalf("audit records = %d, want 1", len(sink.audits))
	}
	got := sink.audits[0]
	if got.Status != "success" || got.ExecutionTimeMs != 25 {
		t.Fatalf("final success audit = %+v, want success with 25ms", got)
	}
}

func TestCommandAuditLifecycleRecordsOneFailedFinalAuditWithDuration(t *testing.T) {
	sink := new(auditSinkFake)
	clock := &fakeAuditClock{now: time.Unix(100, 0)}
	lifecycle := newCommandAuditLifecycle(sink, clock.Now)
	start := lifecycle.Start()
	clock.now = start.Add(7 * time.Millisecond)

	lifecycle.Finish(CommandAudit{CommandName: "roll"}, start, errors.New("command failed"))

	if len(sink.audits) != 1 {
		t.Fatalf("audit records = %d, want 1", len(sink.audits))
	}
	got := sink.audits[0]
	if got.Status != "error" || got.ExecutionTimeMs != 7 || got.ErrorMessage == nil || *got.ErrorMessage != "command failed" {
		t.Fatalf("final error audit = %+v, want error with 7ms and message", got)
	}
}

func TestCommandAuditLifecycleClampsSubMillisecondDurationToOne(t *testing.T) {
	sink := new(auditSinkFake)
	clock := &fakeAuditClock{now: time.Unix(100, 0)}
	lifecycle := newCommandAuditLifecycle(sink, clock.Now)
	start := lifecycle.Start()
	clock.now = start

	lifecycle.Finish(CommandAudit{CommandName: "roll"}, start, nil)

	if sink.audits[0].ExecutionTimeMs != 1 {
		t.Fatalf("duration = %d, want minimum 1ms", sink.audits[0].ExecutionTimeMs)
	}
}

type fakeAuditClock struct{ now time.Time }

func (c *fakeAuditClock) Now() time.Time { return c.now }

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

func TestResolveAdminStatusPreservesGuildRoleLookupFailure(t *testing.T) {
	member := &discordgo.Member{Roles: []string{"role-host"}}
	lookupErr := errors.New("discord unavailable")

	got, err := resolveAdminStatus(member, func() ([]*discordgo.Role, error) {
		return nil, lookupErr
	})

	if got != nil {
		t.Fatalf("admin status = %v, want unknown", *got)
	}
	if !errors.Is(err, lookupErr) {
		t.Fatalf("role lookup error = %v, want %v", err, lookupErr)
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
