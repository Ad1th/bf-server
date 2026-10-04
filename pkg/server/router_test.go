package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRouter_RouteMapping(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files:
	// index.bf
	// hello.bf
	// api/users.bf
	// api/index.bf
	// 404.bf
	files := map[string]string{
		"index.bf":     "+++++++++++++++++++++++++++++++++.", // '!'
		"hello.bf":     "++++++++++++++++++++++++++++++++++.",
		"api/users.bf": "+++++++++++++++++++++++++++++++++++.",
		"api/index.bf": "++++++++++++++++++++++++++++++++++++.",
		"404.bf":       "+++++++++++++++++++++++++++++++++++++.",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	router, err := NewRouter(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	tests := []struct {
		urlPath      string
		expectedPath string
		is404        bool
	}{
		{"/", "/", false},
		{"/index", "/index", false},
		{"/hello", "/hello", false},
		{"/hello/", "/hello", false},
		{"/api/users", "/api/users", false},
		{"/api", "/api", false},
		{"/api/index", "/api/index", false},
		{"/not-found-page", "404", true},
	}

	for _, tt := range tests {
		t.Run(tt.urlPath, func(t *testing.T) {
			route, err := router.Match(tt.urlPath)
			if err != nil {
				t.Fatalf("unexpected error matching route: %v", err)
			}
			if route == nil {
				t.Fatalf("expected route match, got nil")
			}
			if tt.is404 {
				if !route.IsDefault || route.URLPath != "404" {
					t.Errorf("expected 404 fallback route, got %s", route.URLPath)
				}
			} else {
				if route.URLPath != tt.expectedPath {
					t.Errorf("expected route %s, got %s", tt.expectedPath, route.URLPath)
				}
			}
		})
	}
}

func TestRouter_DevModeHotReload(t *testing.T) {
	tmpDir := t.TempDir()

	indexFile := filepath.Join(tmpDir, "index.bf")
	if err := os.WriteFile(indexFile, []byte("+++++."), 0644); err != nil {
		t.Fatalf("failed to write index.bf: %v", err)
	}

	router, err := NewRouter(tmpDir, true) // devMode = true
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	// First match
	r1, err := router.Match("/")
	if err != nil || r1 == nil {
		t.Fatalf("expected initial match, got %v", err)
	}
	if len(r1.Program.Instructions) == 0 {
		t.Fatalf("expected instructions")
	}

	// Add new file dynamically
	newFile := filepath.Join(tmpDir, "dynamic.bf")
	if err := os.WriteFile(newFile, []byte("++++++++++."), 0644); err != nil {
		t.Fatalf("failed to write dynamic.bf: %v", err)
	}

	// In dev mode, Match("/") should automatically discover the new route
	r2, err := router.Match("/dynamic")
	if err != nil {
		t.Fatalf("unexpected error in dev mode reload: %v", err)
	}
	if r2 == nil || r2.URLPath != "/dynamic" {
		t.Fatalf("expected /dynamic to be loaded, got %v", r2)
	}
}
