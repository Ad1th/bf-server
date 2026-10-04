package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSerializeRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/test?query=val", strings.NewReader("hello request body"))
	req.Header.Set("X-Custom-Header", "custom-value")

	serialized, err := SerializeRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	str := string(serialized)
	if !strings.HasPrefix(str, "POST /test?query=val HTTP/1.1\r\n") {
		t.Errorf("expected request line, got %q", str)
	}
	if !strings.Contains(str, "X-Custom-Header: custom-value\r\n") {
		t.Errorf("expected header in serialized output, got %q", str)
	}
	if !strings.HasSuffix(str, "\r\nhello request body") {
		t.Errorf("expected body at the end, got %q", str)
	}
}

func TestParseBrainfuckOutput_Plain(t *testing.T) {
	output := []byte("Hello Plain Text")
	parsed, err := ParseBrainfuckOutput(output, http.StatusOK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", parsed.StatusCode)
	}
	if string(parsed.Body) != "Hello Plain Text" {
		t.Errorf("expected body %q, got %q", "Hello Plain Text", string(parsed.Body))
	}
}

func TestParseBrainfuckOutput_FullHTTP(t *testing.T) {
	output := []byte("HTTP/1.1 201 Created\r\nContent-Type: application/json\r\nX-App: bf\r\n\r\n{\"status\":\"ok\"}")
	parsed, err := ParseBrainfuckOutput(output, http.StatusOK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", parsed.StatusCode)
	}
	if parsed.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", parsed.Header.Get("Content-Type"))
	}
	if parsed.Header.Get("X-App") != "bf" {
		t.Errorf("expected X-App bf, got %s", parsed.Header.Get("X-App"))
	}
	if string(parsed.Body) != "{\"status\":\"ok\"}" {
		t.Errorf("expected body %q, got %q", "{\"status\":\"ok\"}", string(parsed.Body))
	}
}

func TestParseBrainfuckOutput_StatusHeader(t *testing.T) {
	output := []byte("Status: 418 I'm a teapot\r\nContent-Type: text/plain\r\n\r\nTeapot output")
	parsed, err := ParseBrainfuckOutput(output, http.StatusOK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.StatusCode != http.StatusTeapot {
		t.Errorf("expected status 418, got %d", parsed.StatusCode)
	}
	if parsed.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type text/plain, got %s", parsed.Header.Get("Content-Type"))
	}
	if string(parsed.Body) != "Teapot output" {
		t.Errorf("expected body %q, got %q", "Teapot output", string(parsed.Body))
	}
}

func TestParseBrainfuckOutput_Empty(t *testing.T) {
	parsed, err := ParseBrainfuckOutput([]byte{}, http.StatusOK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", parsed.StatusCode)
	}
	if !bytes.Equal(parsed.Body, []byte{}) {
		t.Errorf("expected empty body, got %v", parsed.Body)
	}
}
