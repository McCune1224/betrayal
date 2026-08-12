package whisper

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolveRecipientsRequiresCompleteTwinGroup(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "channel-10", GroupID: "triplet"},
		{PlayerID: 11, ChannelID: "channel-11", GroupID: "triplet"},
		{PlayerID: 12, ChannelID: "channel-12", GroupID: "triplet"},
	}

	got, err := ResolveRecipients(10, 11, players)
	if err != nil {
		t.Fatalf("ResolveRecipients returned error: %v", err)
	}
	want := []string{"channel-10", "channel-11", "channel-12"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveRecipients = %#v, want %#v", got, want)
	}
}

func TestResolveRecipientsRejectsSelfTargetAndIncompleteGroup(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "channel-10", GroupID: "pair"},
		{PlayerID: 11, ChannelID: "channel-11", GroupID: "pair"},
	}
	if _, err := ResolveRecipients(10, 10, players); !errors.Is(err, ErrSelfTarget) {
		t.Fatalf("self target error = %v, want %v", err, ErrSelfTarget)
	}
	if _, err := ResolveRecipients(10, 11, players[:1]); !errors.Is(err, ErrNoRecipients) {
		t.Fatalf("missing target error = %v, want %v", err, ErrNoRecipients)
	}
	if _, err := ResolveRecipients(10, 10, []PlayerConfessional{{PlayerID: 10, ChannelID: "channel-10", GroupID: "pair"}}); !errors.Is(err, ErrSelfTarget) {
		t.Fatalf("self target should take precedence, got %v", err)
	}
}

func TestDeliverSendsAllRecipientsThenSenderReceiptAndOptionalWarning(t *testing.T) {
	sender := &recordingSender{}
	roller := fixedRoller{hit: true}
	pool := []string{"Keep your guard up."}

	result, err := Deliver(DeliveryRequest{
		SenderChannelID:     "sender-channel",
		RecipientChannelIDs: []string{"twin-1", "twin-2", "twin-3"},
		Message:             "The door is open.",
		Timestamp:           "<t:1700000000:F>",
	}, sender, roller, pool)
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if !result.WarningSent {
		t.Fatal("Deliver did not report warning sent")
	}
	want := []sendCall{
		{ChannelID: "twin-1", Content: "<t:1700000000:F>\n\nThe door is open."},
		{ChannelID: "twin-2", Content: "<t:1700000000:F>\n\nThe door is open."},
		{ChannelID: "twin-3", Content: "<t:1700000000:F>\n\nThe door is open."},
		{ChannelID: "twin-1", Content: "Keep your guard up."},
		{ChannelID: "twin-2", Content: "Keep your guard up."},
		{ChannelID: "twin-3", Content: "Keep your guard up."},
		{ChannelID: "sender-channel", Content: "Whisper sent at <t:1700000000:F>\n\nThe door is open."},
	}
	if !reflect.DeepEqual(sender.calls, want) {
		t.Fatalf("send calls = %#v, want %#v", sender.calls, want)
	}
}

func TestResolveSenderRecipientsSendsToEveryLinkedMemberExceptSender(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "sender-channel", GroupID: "pair"},
		{PlayerID: 11, ChannelID: "twin-channel", GroupID: "pair"},
	}

	got, err := ResolveSenderRecipients(10, players)
	if err != nil {
		t.Fatalf("ResolveSenderRecipients returned error: %v", err)
	}
	if want := []string{"twin-channel"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolveSenderRecipients = %#v, want %#v", got, want)
	}
}

type recordingSender struct{ calls []sendCall }
type sendCall struct{ ChannelID, Content string }

func (s *recordingSender) Send(channelID, content string) error {
	s.calls = append(s.calls, sendCall{ChannelID: channelID, Content: content})
	return nil
}

type fixedRoller struct{ hit bool }

func (r fixedRoller) Hit(float64) bool { return r.hit }
func (r fixedRoller) Intn(int) int     { return 0 }
