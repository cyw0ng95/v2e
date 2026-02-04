package main

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSetupRouter_ServeIndexAndAssets(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupRouter_ServeIndexAndAssets", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		indexPath := filepath.Join(dir, "index.html")
		assetPath := filepath.Join(dir, "app.js")
		if err := ioutil.WriteFile(indexPath, []byte("INDEX_CONTENT"), 0o644); err != nil {
			t.Fatalf("failed to write index: %v", err)
		}
		if err := ioutil.WriteFile(assetPath, []byte("console.log(1);"), 0o644); err != nil {
			t.Fatalf("failed to write asset: %v", err)
		}

		router := setupRouter(nil, 1, dir)

		// Request root -> index
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for index, got %d", w.Code)
		}
		if body := w.Body.String(); body != "INDEX_CONTENT" {
			t.Fatalf("unexpected index content: %q", body)
		}

		// Request asset -> app.js
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/app.js", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for asset, got %d", w.Code)
		}
		if body := w.Body.String(); body != "console.log(1);" {
			t.Fatalf("unexpected asset content: %q", body)
		}

		// Non-existent path -> fallback to index.html (SPA)
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/some/nonexistent/path", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for SPA fallback, got %d", w.Code)
		}
		if body := w.Body.String(); body != "INDEX_CONTENT" {
			t.Fatalf("unexpected fallback content: %q", body)
		}
	})

}

func TestSetupRouter_APIPrefixReturns404(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupRouter_APIPrefixReturns404", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		// no files needed; router will skip static when not found but NoRoute must respond 404 for /restful prefix
		router := setupRouter(nil, 1, dir)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/restful/unknown", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for API prefix, got %d", w.Code)
		}
	})

}
