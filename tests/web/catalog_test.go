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

// catalogName is a unique-ish prefix so catalog tests don't collide with any
// real seeded items/abilities/statuses in the shared local DB.
const catalogPrefix = "Wt6 Test "

// TestItemsCRUD covers the full item lifecycle: list -> search -> create ->
// detail -> update -> delete.
func TestItemsCRUD(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	// List page
	resp := client.get("/items")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /items: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "Item Editor") {
		t.Fatal("items page should render")
	}

	// Search partial
	resp = client.get("/items/search?q=" + url.QueryEscape(catalogPrefix))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /items/search: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "NO RESULTS") && !strings.Contains(client.body(resp), "EDIT") {
		t.Fatal("search partial should render results or empty state")
	}

	// Create
	name := catalogPrefix + "Revolver"
	resp = client.do(http.MethodPost, "/items", url.Values{
		"name":        {name},
		"description": {"A trusty six-shooter"},
		"rarity":      {"RARE"},
		"cost":        {"75"},
	}, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /items: expected 303, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}
	loc := resp.Header.Get("Location")
	if !strings.HasPrefix(loc, "/items/") {
		t.Fatalf("create should redirect to the new item, got %q", loc)
	}

	idStr := strings.TrimPrefix(loc, "/items/")
	itemID, _ := strconv.Atoi(idStr)
	t.Cleanup(func() {
		_ = models.New(pool).DeleteItem(context.Background(), int32(itemID))
	})

	// Detail page
	resp = client.get("/items/" + idStr)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /items/:id: expected 200, got %d", resp.StatusCode)
	}
	body := client.body(resp)
	if !strings.Contains(body, name) || !strings.Contains(body, "RARE") {
		t.Fatal("item detail should show the created values")
	}

	// Update
	resp = client.do(http.MethodPost, "/items/"+idStr, url.Values{
		"name":        {name},
		"description": {"A trusty six-shooter, now with engraving"},
		"rarity":      {"EPIC"},
		"cost":        {"150"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /items/:id: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}

	item, err := models.New(pool).GetItem(context.Background(), int32(itemID))
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if item.Rarity != models.RarityEPIC || item.Cost != 150 {
		t.Fatalf("expected EPIC/150 after update, got %+v", item)
	}

	// Invalid rarity is rejected
	resp = client.do(http.MethodPost, "/items/"+idStr, url.Values{
		"name":        {name},
		"description": {"x"},
		"rarity":      {"BANANA"},
		"cost":        {"1"},
	}, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid rarity: expected 400, got %d", resp.StatusCode)
	}

	// Delete
	resp = client.do(http.MethodPost, "/items/"+idStr+"/delete", nil, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /items/:id/delete: expected 303, got %d", resp.StatusCode)
	}
	resp = client.get("/items/" + idStr)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET deleted item: expected 404, got %d", resp.StatusCode)
	}
}

// TestAbilitiesCRUD covers the ability lifecycle.
func TestAbilitiesCRUD(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/abilities")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /abilities: expected 200, got %d", resp.StatusCode)
	}

	name := catalogPrefix + "Clairvoyance"
	resp = client.do(http.MethodPost, "/abilities", url.Values{
		"name":            {name},
		"description":     {"See through deception"},
		"rarity":          {"LEGENDARY"},
		"default_charges": {"3"},
		"any_ability":     {"on"},
	}, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /abilities: expected 303, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}
	idStr := strings.TrimPrefix(resp.Header.Get("Location"), "/abilities/")
	abilityID, _ := strconv.Atoi(idStr)
	t.Cleanup(func() {
		_ = models.New(pool).DeleteAbilityInfo(context.Background(), int32(abilityID))
	})

	// Search finds it
	resp = client.get("/abilities/search?q=" + url.QueryEscape(name))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /abilities/search: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), name) {
		t.Fatal("ability search should find the created ability")
	}

	// Detail + update
	resp = client.get("/abilities/" + idStr)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /abilities/:id: expected 200, got %d", resp.StatusCode)
	}
	resp = client.do(http.MethodPost, "/abilities/"+idStr, url.Values{
		"name":            {name},
		"description":     {"See through deception, twice"},
		"rarity":          {"MYTHICAL"},
		"default_charges": {"5"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /abilities/:id: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}

	ability, err := models.New(pool).GetAbilityInfo(context.Background(), int32(abilityID))
	if err != nil {
		t.Fatalf("get ability: %v", err)
	}
	if ability.Rarity != models.RarityMYTHICAL || ability.DefaultCharges != 5 {
		t.Fatalf("expected MYTHICAL/5 after update, got %+v", ability)
	}

	// Delete
	resp = client.do(http.MethodPost, "/abilities/"+idStr+"/delete", nil, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /abilities/:id/delete: expected 303, got %d", resp.StatusCode)
	}
	resp = client.get("/abilities/" + idStr)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET deleted ability: expected 404, got %d", resp.StatusCode)
	}
}

// TestStatusesCRUD covers the status lifecycle.
func TestStatusesCRUD(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/statuses")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /statuses: expected 200, got %d", resp.StatusCode)
	}

	name := catalogPrefix + "Hex"
	resp = client.do(http.MethodPost, "/statuses", url.Values{
		"name":          {name},
		"description":   {"A creeping curse"},
		"hour_duration": {"36"},
	}, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /statuses: expected 303, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}
	idStr := strings.TrimPrefix(resp.Header.Get("Location"), "/statuses/")
	statusID, _ := strconv.Atoi(idStr)
	t.Cleanup(func() {
		_ = models.New(pool).DeleteStatus(context.Background(), int32(statusID))
	})

	// Detail shows the duration set at create time
	resp = client.get("/statuses/" + idStr)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /statuses/:id: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "36") {
		t.Fatal("status detail should show the hour duration")
	}

	// Update
	resp = client.do(http.MethodPost, "/statuses/"+idStr, url.Values{
		"name":          {name},
		"description":   {"A creeping curse, upgraded"},
		"hour_duration": {"48"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /statuses/:id: expected 200, got %d (body: %s)", resp.StatusCode, client.body(resp))
	}

	// Search finds it
	resp = client.get("/statuses/search?q=" + url.QueryEscape(name))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /statuses/search: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), name) {
		t.Fatal("status search should find the created status")
	}

	// Delete
	resp = client.do(http.MethodPost, "/statuses/"+idStr+"/delete", nil, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /statuses/:id/delete: expected 303, got %d", resp.StatusCode)
	}
	resp = client.get("/statuses/" + idStr)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET deleted status: expected 404, got %d", resp.StatusCode)
	}
}
