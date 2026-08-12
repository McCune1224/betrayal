package web_test

import (
	"io"
	"net/http"
	"regexp"
	"testing"
)

func TestBrowserShellServesSvelteKitAssetsWithoutAuthentication(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	shell := client.get("/login")
	shellBody, err := io.ReadAll(shell.Body)
	if err != nil {
		t.Fatalf("read login shell: %v", err)
	}
	asset := regexp.MustCompile(`/_app/immutable/entry/start[^\"]+\.js`).FindString(string(shellBody))
	if asset == "" {
		t.Fatal("login shell did not reference a SvelteKit start asset")
	}

	response := client.get(asset)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("SvelteKit asset: status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if location := response.Header.Get("Location"); location != "" {
		t.Fatalf("SvelteKit asset: Location = %q, want no redirect", location)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "text/javascript" && contentType != "text/javascript; charset=utf-8" {
		t.Fatalf("SvelteKit asset: Content-Type = %q, want a JavaScript content type", contentType)
	}

	head := client.do(http.MethodHead, asset, nil, nil)
	if head.StatusCode != http.StatusOK {
		t.Fatalf("SvelteKit asset HEAD: status = %d, want %d", head.StatusCode, http.StatusOK)
	}
}
