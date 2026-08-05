package web_test

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
)

// testPlayerID is a snowflake-range ID that will never collide with real
// players in the shared local dev DB.
const testPlayerID = int64(9000000000000000001)

func TestPlayerEditPage(t *testing.T) {
	pool := mustPool(t)
	seedPlayer(t, pool, testPlayerID)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/players/" + strconv.FormatInt(testPlayerID, 10) + "/edit")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /players/:id/edit: expected 200, got %d", resp.StatusCode)
	}
	body := client.body(resp)
	if !strings.Contains(body, "STATS") {
		t.Fatal("edit page should render the stats section")
	}
	if !strings.Contains(body, `value="100"`) {
		t.Fatal("edit page should show the seeded coin count (100)")
	}
	if !strings.Contains(body, "EDIT") {
		t.Fatal("edit page should render the header edit badge")
	}

	// Unknown player -> 404
	resp = client.get("/players/8999999999999999999/edit")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET unknown player edit: expected 404, got %d", resp.StatusCode)
	}
}

func TestPlayerEditStats(t *testing.T) {
	pool := mustPool(t)
	seedPlayer(t, pool, testPlayerID)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.do(http.MethodPost, "/players/"+strconv.FormatInt(testPlayerID, 10)+"/edit", url.Values{
		"coins": {"250"},
		"luck":  {"7"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /players/:id/edit: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), `value="250"`) {
		t.Fatal("stats partial should reflect the new coin value")
	}

	player, err := models.New(pool).GetPlayer(context.Background(), testPlayerID)
	if err != nil {
		t.Fatalf("get player: %v", err)
	}
	if player.Coins != 250 || player.Luck != 7 {
		t.Fatalf("expected coins=250 luck=7, got coins=%d luck=%d", player.Coins, player.Luck)
	}

	// Invalid coin value -> 400, values unchanged
	resp = client.do(http.MethodPost, "/players/"+strconv.FormatInt(testPlayerID, 10)+"/edit", url.Values{
		"coins": {"abc"},
	}, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid coins: expected 400, got %d", resp.StatusCode)
	}
}

// TestPlayerEditItems covers the item add/remove flows via the inventory
// service (fuzzy name lookup, quantity handling).
func TestPlayerEditItems(t *testing.T) {
	pool := mustPool(t)
	seedPlayer(t, pool, testPlayerID)
	item := seedItem(t, pool, "Test Revolver")
	client := newTestClient(t, testServer(t, pool))
	client.login()

	pid := strconv.FormatInt(testPlayerID, 10)

	// Add 2
	resp := client.do(http.MethodPost, "/players/"+pid+"/items/add", url.Values{
		"item_name": {"Test Revolver"},
		"quantity":  {"2"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add item: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}
	if !strings.Contains(client.body(resp), "Test Revolver") {
		t.Fatal("items partial should list the added item")
	}

	q := models.New(pool)
	rows, err := q.ListPlayerItemInventory(context.Background(), testPlayerID)
	if err != nil || len(rows) != 1 || rows[0].Quantity != 2 {
		t.Fatalf("expected 1 item with qty 2, got %+v err=%v", rows, err)
	}

	// Remove 1 -> qty 1
	resp = client.do(http.MethodPost, "/players/"+pid+"/items/remove", url.Values{
		"item_name": {"Test Revolver"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("remove item: expected 200, got %d", resp.StatusCode)
	}
	rows, _ = q.ListPlayerItemInventory(context.Background(), testPlayerID)
	if len(rows) != 1 || rows[0].Quantity != 1 {
		t.Fatalf("after remove expected qty 1, got %+v", rows)
	}

	_ = item // seeded + cleaned up by helper
}

// TestPlayerEditAbilities covers add/remove for abilities (default charges
// when quantity is omitted).
func TestPlayerEditAbilities(t *testing.T) {
	pool := mustPool(t)
	seedPlayer(t, pool, testPlayerID)
	seedAbility(t, pool, "Test Clairvoyance")
	client := newTestClient(t, testServer(t, pool))
	client.login()

	pid := strconv.FormatInt(testPlayerID, 10)

	// Add with blank quantity -> default charges (2 for the seeded ability)
	resp := client.do(http.MethodPost, "/players/"+pid+"/abilities/add", url.Values{
		"ability_name": {"Test Clairvoyance"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add ability: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}
	if !strings.Contains(client.body(resp), "Test Clairvoyance") {
		t.Fatal("abilities partial should list the added ability")
	}

	q := models.New(pool)
	rows, err := q.ListPlayerAbilityJoin(context.Background(), testPlayerID)
	if err != nil || len(rows) != 1 || rows[0].Quantity != 2 {
		t.Fatalf("expected 1 ability with default 2 charges, got %+v err=%v", rows, err)
	}

	// Duplicate add is rejected by the service
	resp = client.do(http.MethodPost, "/players/"+pid+"/abilities/add", url.Values{
		"ability_name": {"Test Clairvoyance"},
	}, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("duplicate ability: expected 400, got %d", resp.StatusCode)
	}

	// Remove
	resp = client.do(http.MethodPost, "/players/"+pid+"/abilities/remove", url.Values{
		"ability_name": {"Test Clairvoyance"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("remove ability: expected 200, got %d", resp.StatusCode)
	}
	rows, _ = q.ListPlayerAbilityJoin(context.Background(), testPlayerID)
	if len(rows) != 0 {
		t.Fatalf("after remove expected 0 abilities, got %+v", rows)
	}
}

// TestPlayerEditStatuses covers add/remove for statuses.
func TestPlayerEditStatuses(t *testing.T) {
	pool := mustPool(t)
	seedPlayer(t, pool, testPlayerID)
	seedStatus(t, pool, "Test Hex")
	client := newTestClient(t, testServer(t, pool))
	client.login()

	pid := strconv.FormatInt(testPlayerID, 10)

	resp := client.do(http.MethodPost, "/players/"+pid+"/statuses/add", url.Values{
		"status_name": {"Test Hex"},
		"quantity":    {"3"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add status: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}

	q := models.New(pool)
	rows, err := q.ListPlayerStatusInventory(context.Background(), testPlayerID)
	if err != nil || len(rows) != 1 || rows[0].Quantity != 3 {
		t.Fatalf("expected 1 status qty 3, got %+v err=%v", rows, err)
	}

	resp = client.do(http.MethodPost, "/players/"+pid+"/statuses/remove", url.Values{
		"status_name": {"Test Hex"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("remove status: expected 200, got %d", resp.StatusCode)
	}
	// Remove removes one at a time: 3 -> 2
	rows, _ = q.ListPlayerStatusInventory(context.Background(), testPlayerID)
	if len(rows) != 1 || rows[0].Quantity != 2 {
		t.Fatalf("after one remove expected qty 2, got %+v", rows)
	}

	// Remove the rest -> 0
	resp = client.do(http.MethodPost, "/players/"+pid+"/statuses/remove", url.Values{
		"status_name": {"Test Hex"},
	}, nil)
	resp = client.do(http.MethodPost, "/players/"+pid+"/statuses/remove", url.Values{
		"status_name": {"Test Hex"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("final status remove: expected 200, got %d", resp.StatusCode)
	}
	rows, _ = q.ListPlayerStatusInventory(context.Background(), testPlayerID)
	if len(rows) != 0 {
		t.Fatalf("after removing all expected 0 statuses, got %+v", rows)
	}
}

// TestPlayerEditPerks covers add/remove for perks.
func TestPlayerEditPerks(t *testing.T) {
	pool := mustPool(t)
	seedPlayer(t, pool, testPlayerID)
	seedPerk(t, pool, "Test Sixth Sense")
	client := newTestClient(t, testServer(t, pool))
	client.login()

	pid := strconv.FormatInt(testPlayerID, 10)

	resp := client.do(http.MethodPost, "/players/"+pid+"/perks/add", url.Values{
		"perk_name": {"Test Sixth Sense"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add perk: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}
	if !strings.Contains(client.body(resp), "Test Sixth Sense") {
		t.Fatal("perks partial should list the added perk")
	}

	q := models.New(pool)
	perks, err := q.ListPlayerPerk(context.Background(), testPlayerID)
	if err != nil || len(perks) != 1 {
		t.Fatalf("expected 1 perk, got %+v err=%v", perks, err)
	}

	// Unknown perk -> 400
	resp = client.do(http.MethodPost, "/players/"+pid+"/perks/add", url.Values{
		"perk_name": {"No Such Perk"},
	}, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown perk: expected 400, got %d", resp.StatusCode)
	}

	resp = client.do(http.MethodPost, "/players/"+pid+"/perks/remove", url.Values{
		"perk_name": {"Test Sixth Sense"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("remove perk: expected 200, got %d", resp.StatusCode)
	}
	perks, _ = q.ListPlayerPerk(context.Background(), testPlayerID)
	if len(perks) != 0 {
		t.Fatalf("after remove expected 0 perks, got %+v", perks)
	}
}
