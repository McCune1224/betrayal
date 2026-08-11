package api

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestPlayersHandlerDetailRejectsInvalidIDAsJSON(t *testing.T) {
	e := echo.New()
	h := NewPlayersHandler(nil)
	req := httptest.NewRequest("GET", "/api/v1/players/not-an-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/players/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-an-id")

	if err := h.Detail(c); err != nil {
		t.Fatalf("Detail returned error: %v", err)
	}
	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := rec.Header().Get("content-type"); got == "" || got[:len("application/json")] != "application/json" {
		t.Fatalf("content type = %q, want JSON", got)
	}
}
