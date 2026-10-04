package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// Config holds the server runtime configuration.
type Config struct {
	Addr       string        `json:"addr"`
	AppDir     string        `json:"app_dir"`
	MemorySize uint          `json:"memory_size"`
	MaxSteps   uint64        `json:"max_steps"`
	DevMode    bool          `json:"dev_mode"`
	ConfigFile string        `json:"-"`
	LogFormat  string        `json:"log_format"`
	LogLevel   string        `json:"log_level"`
	Timeout    time.Duration `json:"timeout"`
}

// DefaultConfig returns the default server configuration.
func DefaultConfig() *Config {
	return &Config{
		Addr:       ":8080",
		AppDir:     "./app",
		MemorySize: 30000,
		MaxSteps:   10000000,
		DevMode:    false,
		LogFormat:  "text",
		LogLevel:   "info",
		Timeout:    5 * time.Second,
	}
}

// LoadFromFile reads and decodes a JSON configuration file into Config.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	// Custom struct for unmarshaling duration
	type rawConfig struct {
		Addr       *string `json:"addr"`
		AppDir     *string `json:"app_dir"`
		MemorySize *uint   `json:"memory_size"`
		MaxSteps   *uint64 `json:"max_steps"`
		DevMode    *bool   `json:"dev_mode"`
		LogFormat  *string `json:"log_format"`
		LogLevel   *string `json:"log_level"`
		Timeout    *string `json:"timeout"`
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if raw.Addr != nil {
		cfg.Addr = *raw.Addr
	}
	if raw.AppDir != nil {
		cfg.AppDir = *raw.AppDir
	}
	if raw.MemorySize != nil {
		cfg.MemorySize = *raw.MemorySize
	}
	if raw.MaxSteps != nil {
		cfg.MaxSteps = *raw.MaxSteps
	}
	if raw.DevMode != nil {
		cfg.DevMode = *raw.DevMode
	}
	if raw.LogFormat != nil {
		cfg.LogFormat = *raw.LogFormat
	}
	if raw.LogLevel != nil {
		cfg.LogLevel = *raw.LogLevel
	}
	if raw.Timeout != nil {
		dur, err := time.ParseDuration(*raw.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format in config: %w", err)
		}
		cfg.Timeout = dur
	}

	return cfg, nil
}

// ParseFlags parses command-line arguments and applies them on top of defaults or config files.
func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("bf-server", flag.ContinueOnError)

	cfgPath := fs.String("config", "", "Path to optional JSON configuration file")
	addr := fs.String("addr", "", "HTTP server listen address (default \":8080\")")
	appDir := fs.String("app", "", "Directory containing Brainfuck route files (default \"./app\")")
	memSize := fs.Uint("memory", 0, "Brainfuck tape size in bytes (default 30000)")
	maxSteps := fs.Uint64("max-steps", 0, "Maximum execution steps per request (default 10000000, 0 for unlimited)")
	dev := fs.Bool("dev", false, "Enable developer hot-reload mode")
	logFmt := fs.String("log-format", "", "Log output format (text, json)")
	logLvl := fs.String("log-level", "", "Log level (debug, info, warn, error)")
	timeoutStr := fs.String("timeout", "", "Request timeout (e.g. 5s, 500ms)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	// If config file flag is provided or default config exists
	if *cfgPath != "" {
		loadedCfg, err := LoadFromFile(*cfgPath)
		if err != nil {
			return nil, err
		}
		cfg = loadedCfg
		cfg.ConfigFile = *cfgPath
	}

	// CLI flags override file config / defaults
	if *addr != "" {
		cfg.Addr = *addr
	}
	if *appDir != "" {
		cfg.AppDir = *appDir
	}
	if *memSize != 0 {
		cfg.MemorySize = *memSize
	}
	if *maxSteps != 0 {
		cfg.MaxSteps = *maxSteps
	}
	// Check if dev flag was explicitly passed
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "dev" {
			cfg.DevMode = *dev
		}
	})
	if *logFmt != "" {
		cfg.LogFormat = *logFmt
	}
	if *logLvl != "" {
		cfg.LogLevel = *logLvl
	}
	if *timeoutStr != "" {
		dur, err := time.ParseDuration(*timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
		cfg.Timeout = dur
	}

	return cfg, nil
}
