package web_test

import (
	"net/http"
	"strconv"
	"testing"
)

type catalogCategoryAPIDTO struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func TestAPICatalogCategoriesCRUDAndAssignment(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	t.Run("creates, updates, lists, and deletes a category", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/categories", []byte(`{"name":"Poisons"}`), true)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create status = %d, want %d: %s", resp.StatusCode, http.StatusCreated, client.body(resp))
		}
		var created catalogCategoryAPIDTO
		decodeAPIJSON(t, resp, &created)
		if created.Name != "Poisons" || created.ID == 0 {
			t.Fatalf("created category = %#v", created)
		}

		detail := client.get("/api/v1/catalog/categories/" + strconv.Itoa(int(created.ID)))
		if detail.StatusCode != http.StatusOK {
			t.Fatalf("detail status = %d, want %d", detail.StatusCode, http.StatusOK)
		}
		var got catalogCategoryAPIDTO
		decodeAPIJSON(t, detail, &got)
		if got.ID != created.ID || got.Name != "Poisons" {
			t.Fatalf("detail category = %#v", got)
		}

		resp = apiRequest(t, client, http.MethodPut, "/api/v1/catalog/categories/"+strconv.Itoa(int(created.ID)), []byte(`{"name":"Toxins"}`), true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("update status = %d, want %d: %s", resp.StatusCode, http.StatusOK, client.body(resp))
		}
		var updated catalogCategoryAPIDTO
		decodeAPIJSON(t, resp, &updated)
		if updated.Name != "Toxins" {
			t.Fatalf("updated category = %#v", updated)
		}

		resp = client.get("/api/v1/catalog/categories")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		var all []catalogCategoryAPIDTO
		decodeAPIJSON(t, resp, &all)
		found := false
		for _, c := range all {
			if c.ID == created.ID {
				found = true
			}
		}
		if !found {
			t.Fatalf("category %d missing from list: %#v", created.ID, all)
		}

		resp = client.do(http.MethodDelete, "/api/v1/catalog/categories/"+strconv.Itoa(int(created.ID)), nil, nil)
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
		detail = client.get("/api/v1/catalog/categories/" + strconv.Itoa(int(created.ID)))
		if detail.StatusCode != http.StatusNotFound {
			t.Fatalf("detail after delete status = %d, want %d", detail.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("assigns and removes categories on items and abilities", func(t *testing.T) {
		catResp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/categories", []byte(`{"name":"Charm"}`), true)
		var cat catalogCategoryAPIDTO
		decodeAPIJSON(t, catResp, &cat)

		itemResp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/items", []byte(`{"name":"Locket","description":"Small locked charm","rarity":"COMMON","cost":5}`), true)
		var item struct {
			ID int32 `json:"id"`
		}
		decodeAPIJSON(t, itemResp, &item)

		abilityResp := apiRequest(t, client, http.MethodPost, "/api/v1/catalog/abilities", []byte(`{"name":"Charm Talk","description":"Words carry weight","rarity":"COMMON","default_charges":1,"any_ability":false}`), true)
		var ability struct {
			ID int32 `json:"id"`
		}
		decodeAPIJSON(t, abilityResp, &ability)

		assign := func(base string) {
			resp := apiRequest(t, client, http.MethodPost, base+"/categories", []byte(`{"category":"Charm"}`), true)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("assign to %s status = %d, want %d: %s", base, resp.StatusCode, http.StatusOK, client.body(resp))
			}
		}
		assign("/api/v1/catalog/items/" + strconv.Itoa(int(item.ID)))
		assign("/api/v1/catalog/abilities/" + strconv.Itoa(int(ability.ID)))

		itemDetail := client.get("/api/v1/catalog/items/" + strconv.Itoa(int(item.ID)))
		var itemDTO struct {
			Categories []string `json:"categories"`
		}
		decodeAPIJSON(t, itemDetail, &itemDTO)
		if len(itemDTO.Categories) != 1 || itemDTO.Categories[0] != "Charm" {
			t.Fatalf("item categories = %#v, want [Charm]", itemDTO.Categories)
		}

		abilityDetail := client.get("/api/v1/catalog/abilities/" + strconv.Itoa(int(ability.ID)))
		var abilityDTO struct {
			Categories []string `json:"categories"`
		}
		decodeAPIJSON(t, abilityDetail, &abilityDTO)
		if len(abilityDTO.Categories) != 1 || abilityDTO.Categories[0] != "Charm" {
			t.Fatalf("ability categories = %#v, want [Charm]", abilityDTO.Categories)
		}

		unassign := func(base string) {
			resp := client.do(http.MethodDelete, base+"/categories/"+strconv.Itoa(int(cat.ID)), nil, nil)
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("unassign from %s status = %d, want %d", base, resp.StatusCode, http.StatusNoContent)
			}
		}
		unassign("/api/v1/catalog/items/" + strconv.Itoa(int(item.ID)))
		unassign("/api/v1/catalog/abilities/" + strconv.Itoa(int(ability.ID)))

		decodeAPIJSON(t, client.get("/api/v1/catalog/items/"+strconv.Itoa(int(item.ID))), &itemDTO)
		if len(itemDTO.Categories) != 0 {
			t.Fatalf("item categories after unassign = %#v, want []", itemDTO.Categories)
		}
		decodeAPIJSON(t, client.get("/api/v1/catalog/abilities/"+strconv.Itoa(int(ability.ID))), &abilityDTO)
		if len(abilityDTO.Categories) != 0 {
			t.Fatalf("ability categories after unassign = %#v, want []", abilityDTO.Categories)
		}
	})
}
