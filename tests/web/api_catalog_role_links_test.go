package web_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

type roleLinkAPIDTO struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	Abilities []struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	} `json:"abilities"`
	Perks []struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	} `json:"perks"`
}

func TestAPICatalogRoleLinks(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	createRole := func(name string) int32 {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles", []byte(`{"name":"`+name+`","description":"Sees what approaches","alignment":"GOOD"}`), true)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create role status = %d: %s", resp.StatusCode, client.body(resp))
		}
		var role roleLinkAPIDTO
		decodeAPIJSON(t, resp, &role)
		return role.ID
	}
	createEntity := func(collection, name string) int32 {
		body := `{"name":"` + name + `","description":"fixture"}`
		if collection == "abilities" {
			body = `{"name":"` + name + `","description":"fixture","rarity":"COMMON","default_charges":1,"any_ability":false}`
		}
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/"+collection, []byte(body), true)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %s status = %d: %s", collection, resp.StatusCode, client.body(resp))
		}
		var entity struct {
			ID int32 `json:"id"`
		}
		decodeAPIJSON(t, resp, &entity)
		return entity.ID
	}
	fetchRole := func(id int32) roleLinkAPIDTO {
		resp := client.get("/api/v1/catalog/roles/" + strconv.Itoa(int(id)))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get role status = %d", resp.StatusCode)
		}
		var role roleLinkAPIDTO
		decodeAPIJSON(t, resp, &role)
		return role
	}

	t.Run("links and unlinks abilities and perks on a role", func(t *testing.T) {
		roleID := createRole("Seer Lane")
		abilityID := createEntity("abilities", "Second Sight")
		perkID := createEntity("perks", "Silver Tongue")

		// Link ability by exact name; the response reflects the updated role.
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/abilities", []byte(`{"ability":"Second Sight"}`), true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("link ability status = %d: %s", resp.StatusCode, client.body(resp))
		}
		var linked roleLinkAPIDTO
		decodeAPIJSON(t, resp, &linked)
		if len(linked.Abilities) != 1 || linked.Abilities[0].ID != abilityID {
			t.Fatalf("role after ability link = %#v", linked)
		}

		// Idempotent re-link keeps a single join row.
		resp = apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/abilities", []byte(`{"ability":"Second Sight"}`), true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("re-link ability status = %d: %s", resp.StatusCode, client.body(resp))
		}
		if got := fetchRole(roleID); len(got.Abilities) != 1 {
			t.Fatalf("role after re-link abilities = %#v, want 1", got.Abilities)
		}

		// Link a perk.
		resp = apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/perks", []byte(`{"perk":"Silver Tongue"}`), true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("link perk status = %d: %s", resp.StatusCode, client.body(resp))
		}
		decodeAPIJSON(t, resp, &linked)
		if len(linked.Perks) != 1 || linked.Perks[0].ID != perkID {
			t.Fatalf("role after perk link = %#v", linked)
		}

		// Unlink both.
		resp = client.do(http.MethodDelete, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/abilities/"+strconv.Itoa(int(abilityID)), nil, nil)
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("unlink ability status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
		resp = client.do(http.MethodDelete, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/perks/"+strconv.Itoa(int(perkID)), nil, nil)
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("unlink perk status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
		role := fetchRole(roleID)
		if len(role.Abilities) != 0 || len(role.Perks) != 0 {
			t.Fatalf("role after unlink = abilities %#v perks %#v, want both empty", role.Abilities, role.Perks)
		}
	})

	t.Run("rejects unknown roles and entities", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles/999999/abilities", []byte(`{"ability":"Second Sight"}`), true)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("unknown role status = %d, want %d: %s", resp.StatusCode, http.StatusNotFound, client.body(resp))
		}

		roleID := createRole("Echo Knight")
		resp = apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/abilities", []byte(`{"ability":"No Such Ability"}`), true)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("unknown ability status = %d, want %d: %s", resp.StatusCode, http.StatusBadRequest, client.body(resp))
		}
		resp = apiRequest(t, client, http.MethodPost, "/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))+"/perks", []byte(`{"perk":"No Such Perk"}`), true)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("unknown perk status = %d, want %d: %s", resp.StatusCode, http.StatusBadRequest, client.body(resp))
		}

		// The role must be untouched by failed links.
		var raw map[string]json.RawMessage
		decodeAPIJSON(t, client.get("/api/v1/catalog/roles/"+strconv.Itoa(int(roleID))), &raw)
		if got := string(raw["name"]); got != `"Echo Knight"` {
			t.Fatalf("role name after failed links = %s", got)
		}
	})
}
