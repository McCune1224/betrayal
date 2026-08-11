// Package ui embeds the temporary static UI shell until the frontend build replaces it.
package ui

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

//go:embed dist dist/* dist/assets/*
var files embed.FS

var hashedAsset = regexp.MustCompile(`(?:^|/)[^/]+\.[0-9a-f]{8,}\.[^/]+$`)

// Handler serves embedded UI assets and falls back to the SPA index.
func Handler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "" {
		if data, err := fs.ReadFile(files, "dist/"+name); err == nil {
			if hashedAsset.MatchString(name) {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
			return
		}
	}

	index, err := fs.ReadFile(files, "dist/index.html")
	if err != nil {
		http.Error(w, "embedded UI index is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(index))
}
