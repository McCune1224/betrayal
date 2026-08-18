package whisper

import (
	"reflect"
	"testing"
)

// cyclingRoller always triggers the doubt swap and walks the pool index in
// order, so a test can prove every pool entry is reachable and the original
// message is never delivered on a hit.
type cyclingRoller struct{ index int }

func (r *cyclingRoller) Hit(float64) bool { return true }
func (r *cyclingRoller) Intn(n int) int {
	i := r.index
	r.index = (r.index + 1) % n
	return i
}

// TestDeliverPicksRandomDoubtMessagesFromPool verifies that on a suspicion hit
// every recipient in the group receives the SAME randomly chosen pool message
// and the player's original text is never delivered. Over a full cycle of the
// pool every entry must be observed, proving Intn (not index 0) is consulted.
func TestDeliverPicksRandomDoubtMessagesFromPool(t *testing.T) {
	pool := []string{"doubt A", "doubt B", "doubt C"}
	seen := map[string]int{}
	roller := &cyclingRoller{}
	for i := 0; i < 90; i++ {
		sender := &recordingSender{}
		result, err := Deliver(DeliveryRequest{
			SenderChannelID:     "sender-channel",
			RecipientChannelIDs: []string{"twin-1", "twin-2"},
			Message:             "The door is open.",
		}, sender, roller, pool)
		if err != nil {
			t.Fatalf("Deliver returned error: %v", err)
		}
		if !result.WarningSent {
			t.Fatal("Deliver did not report the doubt swap on a hit")
		}
		if len(sender.calls) != 2 {
			t.Fatalf("send calls = %#v, want both recipients", sender.calls)
		}
		first := sender.calls[0].Content
		for _, call := range sender.calls[1:] {
			if call.Content != first {
				t.Fatalf("recipients received different content: %q vs %q", first, call.Content)
			}
		}
		inPool := false
		for _, p := range pool {
			if first == p {
				inPool = true
				break
			}
		}
		if !inPool {
			t.Fatalf("delivered %q, want one of the pool; the original message must never be sent on a hit", first)
		}
		seen[first]++
	}
	for _, p := range pool {
		if seen[p] == 0 {
			t.Fatalf("pool message %q was never selected; Intn is not being consulted", p)
		}
	}
}

// TestDeliverNeverSwapsWhenWarningPoolIsEmpty verifies the gate
// `len(warningPool) > 0`: with no doubt messages configured the player's
// original message goes out untouched even when the roller would hit.
func TestDeliverNeverSwapsWhenWarningPoolIsEmpty(t *testing.T) {
	sender := &recordingSender{}
	result, err := Deliver(DeliveryRequest{
		SenderChannelID:     "sender-channel",
		RecipientChannelIDs: []string{"twin-channel"},
		Message:             "The door is open.",
	}, sender, fixedRoller{hit: true}, nil)
	if err != nil {
		t.Fatalf("Deliver returned error: %v", err)
	}
	if result.WarningSent {
		t.Fatal("doubt swap reported with an empty warning pool")
	}
	want := []sendCall{{ChannelID: "twin-channel", Content: "Your twin whispers:\n\n> The door is open.\n\nA message passed quietly through the mirrors."}}
	if !reflect.DeepEqual(sender.calls, want) {
		t.Fatalf("send calls = %#v, want original message delivered", sender.calls)
	}
	if result.SenderStatus != "Your message found its way to your twin." {
		t.Fatalf("sender status = %q, want the plain receipt when no swap occurs", result.SenderStatus)
	}
}

// TestDeliverNeverSwapsWithoutRecipientsOrRoller pins the remaining gate terms:
// no recipient channels means nothing to deliver (and no swap), and a nil
// roller must never panic or swap even with a configured pool.
func TestDeliverNeverSwapsWithoutRecipientsOrRoller(t *testing.T) {
	pool := []string{"Keep your guard up."}

	t.Run("no recipients", func(t *testing.T) {
		sender := &recordingSender{}
		result, err := Deliver(DeliveryRequest{
			SenderChannelID: "sender-channel",
			Message:         "The door is open.",
		}, sender, fixedRoller{hit: true}, pool)
		if err != nil {
			t.Fatalf("Deliver returned error: %v", err)
		}
		if result.WarningSent {
			t.Fatal("doubt swap reported with no recipients")
		}
		if len(sender.calls) != 0 {
			t.Fatalf("send calls = %#v, want no deliveries", sender.calls)
		}
	})

	t.Run("nil roller", func(t *testing.T) {
		sender := &recordingSender{}
		result, err := Deliver(DeliveryRequest{
			SenderChannelID:     "sender-channel",
			RecipientChannelIDs: []string{"twin-channel"},
			Message:             "The door is open.",
		}, sender, nil, pool)
		if err != nil {
			t.Fatalf("Deliver returned error: %v", err)
		}
		if result.WarningSent {
			t.Fatal("doubt swap reported with a nil roller")
		}
		want := []sendCall{{ChannelID: "twin-channel", Content: "Your twin whispers:\n\n> The door is open.\n\nA message passed quietly through the mirrors."}}
		if !reflect.DeepEqual(sender.calls, want) {
			t.Fatalf("send calls = %#v, want original message delivered", sender.calls)
		}
	})
}
