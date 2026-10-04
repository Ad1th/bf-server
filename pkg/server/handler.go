package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/madith/bf-server/pkg/interpreter"
)

// Handler handles incoming HTTP requests by executing matching Brainfuck programs.
type Handler struct {
	Router     *Router
	MemorySize uint
	MaxSteps   uint64
	Timeout    time.Duration
	Logger     *slog.Logger
}

// NewHandler creates a new HTTP handler for Brainfuck applications.
func NewHandler(router *Router, memSize uint, maxSteps uint64, timeout time.Duration, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		Router:     router,
		MemorySize: memSize,
		MaxSteps:   maxSteps,
		Timeout:    timeout,
		Logger:     logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.Timeout)
		defer cancel()
	}

	route, err := h.Router.Match(r.URL.Path)
	if err != nil {
		h.Logger.Error("router error", "path", r.URL.Path, "error", err)
		http.Error(w, fmt.Sprintf("Router Error: %v", err), http.StatusInternalServerError)
		return
	}

	if route == nil {
		http.NotFound(w, r)
		return
	}

	// Prepare request payload for Brainfuck stdin
	reqBytes, err := SerializeRequest(r)
	if err != nil {
		h.Logger.Error("failed to serialize request", "error", err)
		http.Error(w, "Failed to serialize request", http.StatusInternalServerError)
		return
	}

	var outBuf bytes.Buffer
	vm := interpreter.NewVM(
		interpreter.WithMemorySize(h.MemorySize),
		interpreter.WithMaxSteps(h.MaxSteps),
		interpreter.WithInput(bytes.NewReader(reqBytes)),
		interpreter.WithOutput(&outBuf),
		interpreter.WithContext(ctx),
	)

	stats, runErr := vm.Run(route.Program)
	if runErr != nil {
		h.handleExecutionError(w, r, route, runErr, stats)
		return
	}

	// Determine default status code based on route
	defaultStatus := http.StatusOK
	if route.IsDefault && route.URLPath == "404" {
		defaultStatus = http.StatusNotFound
	}

	parsed, err := ParseBrainfuckOutput(outBuf.Bytes(), defaultStatus)
	if err != nil {
		h.Logger.Error("failed to parse Brainfuck response", "error", err)
		http.Error(w, fmt.Sprintf("Response Parsing Error: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply custom headers from Brainfuck
	for k, vals := range parsed.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	// Add telemetry headers
	w.Header().Set("X-Brainfuck-Steps", strconv.FormatUint(stats.Steps, 10))
	w.Header().Set("X-Brainfuck-Memory", strconv.Itoa(stats.MaxMemoryUsed))
	w.Header().Set("X-Brainfuck-Duration", stats.Duration.String())
	w.Header().Set("Server", "bf-server/1.0")

	// Infer Content-Type if not set by Brainfuck program
	if w.Header().Get("Content-Type") == "" {
		if len(parsed.Body) > 0 {
			bodyStr := strings.TrimSpace(string(parsed.Body))
			if strings.HasPrefix(strings.ToLower(bodyStr), "<!doctype html") || strings.HasPrefix(strings.ToLower(bodyStr), "<html") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
			} else {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
	}

	w.WriteHeader(parsed.StatusCode)
	if len(parsed.Body) > 0 {
		_, _ = w.Write(parsed.Body)
	}
}

func (h *Handler) handleExecutionError(w http.ResponseWriter, r *http.Request, route *Route, err error, stats *interpreter.ExecutionStats) {
	h.Logger.Warn("Brainfuck execution error",
		"path", r.URL.Path,
		"file", route.FilePath,
		"error", err,
		"steps", stats.Steps,
	)

	w.Header().Set("X-Brainfuck-Steps", strconv.FormatUint(stats.Steps, 10))
	w.Header().Set("Server", "bf-server/1.0")

	if errors.Is(err, interpreter.ErrStepLimitExceeded) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusLoopDetected)
		fmt.Fprintf(w, "Execution Error: execution step limit exceeded (%d steps executed, max: %d)\n", stats.Steps, h.MaxSteps)
		return
	}

	if errors.Is(err, interpreter.ErrPointerOutOfBounds) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Execution Error: memory pointer out of bounds (tape size: %d bytes)\n", h.MemorySize)
		return
	}

	if errors.Is(err, interpreter.ErrExecutionCanceled) || errors.Is(err, context.DeadlineExceeded) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintf(w, "Execution Error: request timed out or was canceled\n")
		return
	}

	// Generic internal server error
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Runtime Error: %v\n", err)
}
