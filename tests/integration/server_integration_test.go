package integration

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/madith/bf-server/pkg/config"
	"github.com/madith/bf-server/pkg/server"
)

func TestIntegration_BasicApp(t *testing.T) {
	appDir := filepath.Join("..", "..", "examples", "basic")

	cfg := &config.Config{
		Addr:       "127.0.0.1:0",
		AppDir:     appDir,
		MemorySize: 30000,
		MaxSteps:   10000000,
		DevMode:    false,
		Timeout:    5 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(srv.HTTPServer.Handler)
	defer ts.Close()

	client := ts.Client()

	// Test GET /
	t.Run("GET / (index)", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "Welcome to bf-server!") {
			t.Errorf("expected welcome message, got %q", string(body))
		}

		if res.Header.Get("X-Brainfuck-Steps") == "" {
			t.Errorf("expected X-Brainfuck-Steps header")
		}
	})

	// Test GET /hello
	t.Run("GET /hello", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/hello")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "Hello, World!") {
			t.Errorf("expected Hello World message, got %q", string(body))
		}
	})

	// Test GET /about
	t.Run("GET /about", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/about")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "bf-server") {
			t.Errorf("expected about message, got %q", string(body))
		}
	})

	// Test 404 Fallback
	t.Run("GET /unmapped-route (404.bf fallback)", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/unmapped-route")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "404 Not Found") {
			t.Errorf("expected 404 message, got %q", string(body))
		}
	})
}

func TestIntegration_EchoApp(t *testing.T) {
	appDir := filepath.Join("..", "..", "examples", "echo")

	cfg := &config.Config{
		Addr:       "127.0.0.1:0",
		AppDir:     appDir,
		MemorySize: 30000,
		MaxSteps:   10000000,
		DevMode:    false,
		Timeout:    5 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(srv.HTTPServer.Handler)
	defer ts.Close()

	payload := "CUSTOM_PAYLOAD_BODY_FOR_ECHO"
	req, err := http.NewRequest("POST", ts.URL+"/", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("X-Test-Echo", "Verified")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "POST / HTTP/1.1") {
		t.Errorf("expected echo to contain request line, got %q", bodyStr)
	}
	if !strings.Contains(bodyStr, "X-Test-Echo: Verified") {
		t.Errorf("expected echo to contain custom header, got %q", bodyStr)
	}
	if !strings.Contains(bodyStr, payload) {
		t.Errorf("expected echo to contain payload body, got %q", bodyStr)
	}
}

func TestIntegration_CustomStatusTeapot(t *testing.T) {
	appDir := filepath.Join("..", "..", "examples", "custom_status")

	cfg := &config.Config{
		Addr:       "127.0.0.1:0",
		AppDir:     appDir,
		MemorySize: 30000,
		MaxSteps:   10000000,
		DevMode:    false,
		Timeout:    5 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(srv.HTTPServer.Handler)
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL + "/teapot")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusTeapot {
		t.Errorf("expected 418 I'm a teapot, got %d", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "teapot") {
		t.Errorf("expected teapot in response body, got %q", string(body))
	}
}

func TestIntegration_HTMLApp(t *testing.T) {
	appDir := filepath.Join("..", "..", "examples", "html")

	cfg := &config.Config{
		Addr:       "127.0.0.1:0",
		AppDir:     appDir,
		MemorySize: 30000,
		MaxSteps:   10000000,
		DevMode:    false,
		Timeout:    5 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(srv.HTTPServer.Handler)
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}

	ct := res.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected Content-Type text/html, got %q", ct)
	}

	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "Brainfuck Web Server") {
		t.Errorf("expected HTML content in body, got %q", string(body))
	}
}

func TestIntegration_StepLimitExceeded(t *testing.T) {
	tmpDir := t.TempDir()
	loopFile := filepath.Join(tmpDir, "infinite.bf")
	// Infinite loop
	if err := os.WriteFile(loopFile, []byte("+[]"), 0644); err != nil {
		t.Fatalf("failed to write infinite.bf: %v", err)
	}

	cfg := &config.Config{
		Addr:       "127.0.0.1:0",
		AppDir:     tmpDir,
		MemorySize: 30000,
		MaxSteps:   500, // Small limit for test
		DevMode:    false,
		Timeout:    5 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(srv.HTTPServer.Handler)
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL + "/infinite")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusLoopDetected {
		t.Errorf("expected 508 Loop Detected, got %d", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	if !bytes.Contains(body, []byte("step limit exceeded")) {
		t.Errorf("expected error message in body, got %s", string(body))
	}
}

func TestIntegration_BrainfuckNativeRouter(t *testing.T) {
	appDir := filepath.Join("..", "..", "examples", "router")

	cfg := &config.Config{
		Addr:       "127.0.0.1:0",
		AppDir:     appDir,
		MemorySize: 30000,
		MaxSteps:   10000000,
		DevMode:    false,
		Timeout:    5 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ts := httptest.NewServer(srv.HTTPServer.Handler)
	defer ts.Close()

	client := ts.Client()

	// 1. GET / routed natively in Brainfuck
	t.Run("GET /", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", res.StatusCode)
		}
		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "Routed natively in Brainfuck") {
			t.Errorf("expected native routing body, got %q", string(body))
		}
	})

	// 2. GET /hello routed natively in Brainfuck
	t.Run("GET /hello", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/hello")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", res.StatusCode)
		}
		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "Greetings from Brainfuck router") {
			t.Errorf("expected hello router body, got %q", string(body))
		}
	})

	// 3. GET /unmatched routed to 404 in Brainfuck
	t.Run("GET /unmatched", func(t *testing.T) {
		res, err := client.Get(ts.URL + "/unmatched")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", res.StatusCode)
		}
		body, _ := io.ReadAll(res.Body)
		if !strings.Contains(string(body), "Brainfuck router could not find this endpoint") {
			t.Errorf("expected 404 router body, got %q", string(body))
		}
	})
}
