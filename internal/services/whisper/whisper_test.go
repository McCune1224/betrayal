package whisper

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolveRecipientsRequiresCompleteTwinGroup(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "channel-10", GroupID: "triplet", Alive: true},
		{PlayerID: 11, ChannelID: "channel-11", GroupID: "triplet", Alive: true},
		{PlayerID: 12, ChannelID: "channel-12", GroupID: "triplet", Alive: true},
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
	if result.SenderStatus != "Something blurred between intention and arrival. The mirrors did not carry your words as spoken." {
		t.Fatalf("sender status = %q, want vague doubt status", result.SenderStatus)
	}
	want := []sendCall{
		{ChannelID: "twin-1", Content: "Keep your guard up."},
		{ChannelID: "twin-2", Content: "Keep your guard up."},
	}
	if !reflect.DeepEqual(sender.calls, want) {
		t.Fatalf("send calls = %#v, want %#v", sender.calls, want)
	}
}

func TestResolveSenderRecipientsSendsToEveryLinkedMemberExceptSender(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "sender-channel", GroupID: "pair", Alive: true},
		{PlayerID: 11, ChannelID: "twin-channel", GroupID: "pair", Alive: true},
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
	if len(sender.calls) != 1 {
		t.Fatalf("sender calls = %#v, want only recipient delivery", sender.calls)
	}
	want := []sendCall{{ChannelID: "twin-channel", Content: "Your twin whispers:\n\n> The door is open.\n\nA message passed quietly through the mirrors."}}
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

func TestResolveSenderDeliveryExcludesDeadRecipientsAndReportsStatus(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "sender-channel", GroupID: "triplet", Alive: true},
		{PlayerID: 11, ChannelID: "living-twin", GroupID: "triplet", Alive: true},
		{PlayerID: 12, ChannelID: "dead-twin", GroupID: "triplet", Alive: false},
	}

	got, err := ResolveSenderDelivery(10, players)
	if err != nil {
		t.Fatalf("ResolveSenderDelivery returned error: %v", err)
	}
	if want := []string{"living-twin"}; !reflect.DeepEqual(got.ChannelIDs, want) {
		t.Fatalf("ResolveSenderDelivery channels = %#v, want %#v", got.ChannelIDs, want)
	}
	if got.GroupSize != 3 || got.AliveRecipients != 1 || got.DeadRecipients != 1 {
		t.Fatalf("ResolveSenderDelivery status = %#v, want triplet with one alive and one dead recipient", got)
	}
}

func TestResolveSenderDeliveryRejectsDeadSender(t *testing.T) {
	players := []PlayerConfessional{
		{PlayerID: 10, ChannelID: "sender-channel", GroupID: "pair", Alive: false},
		{PlayerID: 11, ChannelID: "twin-channel", GroupID: "pair", Alive: true},
	}
	if _, err := ResolveSenderDelivery(10, players); !errors.Is(err, ErrDeadSender) {
		t.Fatalf("dead sender error = %v, want %v", err, ErrDeadSender)
	}
}

func TestDeliverUsesShatteredWindowReceiptWhenAllRecipientsAreDead(t *testing.T) {
	sender := &recordingSender{}

	result, err := Deliver(DeliveryRequest{
		SenderChannelID: "sender-channel",
		Message:         "The door is open.",
		GroupSize:       2,
		DeadRecipients:  1,
	}, sender, fixedRoller{hit: true}, []string{"Keep your guard up."})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	var want []sendCall
	if result.SenderStatus != "The twin window has shattered. There is no living reflection left to receive your words." {
		t.Fatalf("sender status = %q, want shattered-window status", result.SenderStatus)
	}
	if !reflect.DeepEqual(sender.calls, want) {
		t.Fatalf("send calls = %#v, want %#v", sender.calls, want)
	}
}

func TestDeliverUsesCrackedMirrorReceiptWhenSomeTripletMembersAreDead(t *testing.T) {
	sender := &recordingSender{}

	result, err := Deliver(DeliveryRequest{
		SenderChannelID:     "sender-channel",
		RecipientChannelIDs: []string{"living-sibling"},
		Message:             "The door is open.",
		GroupSize:           3,
		AliveRecipients:     1,
		DeadRecipients:      1,
	}, sender, fixedRoller{hit: true}, []string{"Keep your guard up."})
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	want := []sendCall{{ChannelID: "living-sibling", Content: "Keep your guard up."}}
	if result.SenderStatus != "A crack runs through the triplet’s mirror. Your words reached the reflections that remain." {
		t.Fatalf("sender status = %q, want cracked-mirror status", result.SenderStatus)
	}
	if !reflect.DeepEqual(sender.calls, want) {
		t.Fatalf("send calls = %#v, want %#v", sender.calls, want)
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
