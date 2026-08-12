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
		RecipientChannelIDs: []string{"twin-1", "twin-2"},
		Message:             "The door is open.",
	}, sender, roller, pool)
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if !result.WarningSent {
		t.Fatal("Deliver did not report warning sent")
	}
	want := []sendCall{
		{ChannelID: "twin-1", Content: "Keep your guard up."},
		{ChannelID: "twin-2", Content: "Keep your guard up."},
		{ChannelID: "sender-channel", Content: "Whisper sent.\n\nSomething blurred between intention and arrival. The mirrors did not carry your words as spoken."},
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

func TestDeliverSendsOriginalMessageWhenDoubtDoesNotTrigger(t *testing.T) {
	sender := &recordingSender{}

	_, err := Deliver(DeliveryRequest{
		SenderChannelID:     "sender-channel",
		RecipientChannelIDs: []string{"twin-channel"},
		Message:             "The door is open.",
	}, sender, fixedRoller{hit: false}, []string{"Keep your guard up."})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	want := []sendCall{
		{ChannelID: "twin-channel", Content: "Your twin whispers:\n\n> The door is open.\n\nA message passed quietly through the mirrors."},
		{ChannelID: "sender-channel", Content: "Whisper sent.\n\nYour message found its way to your twin."},
	}
	if !reflect.DeepEqual(sender.calls, want) {
		t.Fatalf("send calls = %#v, want %#v", sender.calls, want)
	}
}

func TestRelationshipLabelSupportsLargerGroups(t *testing.T) {
	for size, want := range map[int]string{2: "twin", 3: "triplet", 4: "group of 4", 9: "group of 9"} {
		if got := relationshipLabel(size); got != want {
			t.Errorf("relationshipLabel(%d) = %q, want %q", size, got, want)
		}
	}
}

func TestSuspicionChanceIsFivePercent(t *testing.T) {
	if SuspicionChance != 0.05 {
		t.Fatalf("SuspicionChance = %v, want 0.05", SuspicionChance)
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
