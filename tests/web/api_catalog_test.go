package web_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
)

type catalogRoleDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Alignment   string `json:"alignment"`
	Abilities   []struct {
		ID int32 `json:"id"`
	} `json:"abilities"`
	Perks []struct {
		ID int32 `json:"id"`
	} `json:"perks"`
}

func TestAPICatalogRolesListAndDetailUseStableDTOs(t *testing.T) {
	pool := mustPool(t)
	role, err := models.New(pool).CreateRole(context.Background(), models.CreateRoleParams{Name: "Oracle", Description: "Sees danger", Alignment: models.AlignmentGOOD})
	if err != nil {
		t.Fatal(err)
	}
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/api/v1/catalog/roles?q=" + url.QueryEscape("Ora"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want 200", resp.StatusCode)
	}
	assertAPIJSON(t, resp)
	var roles []catalogRoleDTO
	decodeAPIJSON(t, resp, &roles)
	if len(roles) != 1 || roles[0].ID != role.ID || roles[0].Name != "Oracle" {
		t.Fatalf("roles = %#v", roles)
	}

	resp = client.get("/api/v1/catalog/roles/" + strconv.Itoa(int(role.ID)))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("detail status = %d, want 200", resp.StatusCode)
	}
	var detail catalogRoleDTO
	decodeAPIJSON(t, resp, &detail)
	if detail.ID != role.ID || detail.Alignment != "GOOD" {
		t.Fatalf("detail = %#v", detail)
	}

	var raw []map[string]json.RawMessage
	decodeAPIJSON(t, client.get("/api/v1/catalog/roles"), &raw)
	if len(raw) != 1 {
		t.Fatalf("raw roles = %#v", raw)
	}
	if _, ok := raw[0]["created_at"]; ok {
		t.Error("role DTO leaked created_at")
	}
	_ = resp
}
