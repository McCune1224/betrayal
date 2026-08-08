package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/services/gamereset"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

const resetConfirmation = "RESET BETRAYAL GAME"

type ResetHandler struct {
	service *gamereset.Service
	isProd  bool
}

func NewResetHandler(service *gamereset.Service, isProd bool) *ResetHandler {
	return &ResetHandler{service: service, isProd: isProd}
}

func (h *ResetHandler) Page(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	summary, err := h.service.Preview(ctx)
	data := pages.ResetPageData{Summary: summary, IsProd: h.isProd, Confirmation: resetConfirmation}
	if sources, sourceErr := h.service.ListSources(ctx); sourceErr == nil {
		for _, source := range sources {
			ready := source.Enabled && strings.TrimSpace(source.Url) != ""
			note := "Ready to fetch"
			if !source.Enabled {
				note = "Disabled — enable this source on the Sync page"
			} else if strings.TrimSpace(source.Url) == "" {
				note = "No URL configured — add it on the Sync page"
			}
			data.Sources = append(data.Sources, pages.ResetSourceData{Name: source.Name, Ready: ready, Note: note})
		}
	} else if err == nil {
		err = sourceErr
	}
	if err != nil {
		data.Error = err.Error()
	}
	return render(c, http.StatusOK, pages.ResetPage(data))
}

func (h *ResetHandler) Execute(c echo.Context) error {
	if !strings.EqualFold(strings.TrimSpace(c.FormValue("confirm")), resetConfirmation) {
		return h.error(c, "Type RESET BETRAYAL GAME exactly to confirm this reset")
	}
	if c.FormValue("understand") != "on" {
		return h.error(c, "Check the acknowledgement box before resetting the game")
	}

	result, err := h.service.Execute(c.Request().Context())
	if err != nil {
		return h.error(c, "Reset cancelled — no database changes were committed: "+err.Error())
	}
	data := pages.ResetResultData{Summary: result.Summary, Sources: result.Sources}
	c.Response().Header().Set("HX-Trigger", toastTrigger("Game reset complete — Day 0 restored and CSV data imported", "success"))
	return render(c, http.StatusOK, pages.ResetResult(data))
}

func (h *ResetHandler) error(c echo.Context, message string) error {
	c.Response().Header().Set("HX-Trigger", toastTrigger(message, "error"))
	return c.String(http.StatusBadRequest, message)
}
