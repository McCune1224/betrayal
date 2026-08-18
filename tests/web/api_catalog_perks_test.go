package web_test

import (
	"net/http"
	"strconv"
	"testing"
)

type catalogPerkAPIDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func TestAPICatalogPerksCRUD(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	perkJSON := `{"name":"Silvertongue","description":"Words bend in your favor."}`

	t.Run("rejects malformed create payload", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/perks", []byte(`{not json`), true)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("create status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("creates and reads a perk", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/perks", []byte(perkJSON), true)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create status = %d, want %d: %s", resp.StatusCode, http.StatusCreated, client.body(resp))
		}
		assertAPIJSON(t, resp)
		var created catalogPerkAPIDTO
		decodeAPIJSON(t, resp, &created)
		if created.Name != "Silvertongue" || created.Description == "" || created.ID == 0 {
			t.Fatalf("created perk = %#v", created)
		}

		detail := client.get("/api/v1/catalog/perks/" + strconv.Itoa(int(created.ID)))
		if detail.StatusCode != http.StatusOK {
			t.Fatalf("detail status = %d, want %d", detail.StatusCode, http.StatusOK)
		}
		var got catalogPerkAPIDTO
		decodeAPIJSON(t, detail, &got)
		if got.ID != created.ID || got.Name != "Silvertongue" {
			t.Fatalf("detail perk = %#v", got)
		}

		t.Run("updates the perk", func(t *testing.T) {
			resp := apiRequest(t, client, http.MethodPut, "/api/v1/catalog/perks/"+strconv.Itoa(int(created.ID)), []byte(`{"name":"Silvertongue","description":"Renamed description."}`), true)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("update status = %d, want %d: %s", resp.StatusCode, http.StatusOK, client.body(resp))
			}
			var updated catalogPerkAPIDTO
			decodeAPIJSON(t, resp, &updated)
			if updated.Description != "Renamed description." {
				t.Fatalf("updated perk = %#v", updated)
			}
		})

		t.Run("lists the perk", func(t *testing.T) {
			resp := client.get("/api/v1/catalog/perks")
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("list status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			var perks []catalogPerkAPIDTO
			decodeAPIJSON(t, resp, &perks)
			found := false
			for _, perk := range perks {
				if perk.ID == created.ID && perk.Name == "Silvertongue" {
					found = true
				}
			}
			if !found {
				t.Fatalf("perk %d missing from list: %#v", created.ID, perks)
			}
		})

		t.Run("deletes the perk and reports 404 afterwards", func(t *testing.T) {
			resp := client.do(http.MethodDelete, "/api/v1/catalog/perks/"+strconv.Itoa(int(created.ID)), nil, nil)
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
			}
			detail := client.get("/api/v1/catalog/perks/" + strconv.Itoa(int(created.ID)))
			if detail.StatusCode != http.StatusNotFound {
				t.Fatalf("detail after delete status = %d, want %d", detail.StatusCode, http.StatusNotFound)
			}
		})
	})
}
