package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Addr != ":8080" {
		t.Errorf("expected default addr :8080, got %s", cfg.Addr)
	}
	if cfg.MemorySize != 30000 {
		t.Errorf("expected default memory 30000, got %d", cfg.MemorySize)
	}
	if cfg.MaxSteps != 10000000 {
		t.Errorf("expected default max steps 10000000, got %d", cfg.MaxSteps)
	}
	if cfg.DevMode != false {
		t.Errorf("expected default dev mode false")
	}
}

func TestParseFlags(t *testing.T) {
	args := []string{
		"--addr", ":9090",
		"--app", "./my-app",
		"--memory", "65536",
		"--max-steps", "500000",
		"--dev",
		"--log-format", "json",
		"--log-level", "debug",
		"--timeout", "10s",
	}

	cfg, err := ParseFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Addr != ":9090" {
		t.Errorf("expected addr :9090, got %s", cfg.Addr)
	}
	if cfg.AppDir != "./my-app" {
		t.Errorf("expected appDir ./my-app, got %s", cfg.AppDir)
	}
	if cfg.MemorySize != 65536 {
		t.Errorf("expected memory 65536, got %d", cfg.MemorySize)
	}
	if cfg.MaxSteps != 500000 {
		t.Errorf("expected max-steps 500000, got %d", cfg.MaxSteps)
	}
	if !cfg.DevMode {
		t.Errorf("expected dev mode true")
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected log format json, got %s", cfg.LogFormat)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Timeout)
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test-config.json")

	content := `{
		"addr": "127.0.0.1:3000",
		"app_dir": "/var/bf-app",
		"memory_size": 40000,
		"max_steps": 2000000,
		"dev_mode": true,
		"log_format": "json",
		"log_level": "warn",
		"timeout": "2s"
	}`

	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Addr != "127.0.0.1:3000" {
		t.Errorf("expected addr 127.0.0.1:3000, got %s", cfg.Addr)
	}
	if cfg.AppDir != "/var/bf-app" {
		t.Errorf("expected appDir /var/bf-app, got %s", cfg.AppDir)
	}
	if cfg.MemorySize != 40000 {
		t.Errorf("expected memory 40000, got %d", cfg.MemorySize)
	}
	if cfg.MaxSteps != 2000000 {
		t.Errorf("expected max-steps 2000000, got %d", cfg.MaxSteps)
	}
	if !cfg.DevMode {
		t.Errorf("expected dev mode true")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("expected log level warn, got %s", cfg.LogLevel)
	}
	if cfg.Timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got %v", cfg.Timeout)
	}
}
