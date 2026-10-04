package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/madith/bf-server/pkg/config"
)

// Server is the HTTP server for Brainfuck applications.
type Server struct {
	Config     *config.Config
	Router     *Router
	Handler    *Handler
	HTTPServer *http.Server
	Logger     *slog.Logger
	listener   net.Listener
}

// New creates and configures a new Server instance.
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}

	router, err := NewRouter(cfg.AppDir, cfg.DevMode)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize router: %w", err)
	}

	handler := NewHandler(router, cfg.MemorySize, cfg.MaxSteps, cfg.Timeout, logger)

	// Chain middlewares
	var rootHandler http.Handler = handler
	rootHandler = LoggingMiddleware(logger)(rootHandler)
	rootHandler = RecoveryMiddleware(logger)(rootHandler)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		Config:     cfg,
		Router:     router,
		Handler:    handler,
		HTTPServer: httpServer,
		Logger:     logger,
	}, nil
}

// Start listens and serves HTTP requests.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		return fmt.Errorf("failed to bind on %s: %w", s.Config.Addr, err)
	}
	s.listener = ln

	s.Logger.Info("Brainfuck HTTP server started",
		"addr", ln.Addr().String(),
		"app_dir", s.Config.AppDir,
		"routes", s.Router.RouteCount(),
		"memory_cells", s.Config.MemorySize,
		"max_steps", s.Config.MaxSteps,
		"dev_mode", s.Config.DevMode,
	)

	err = s.HTTPServer.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.Info("Shutting down server...")
	return s.HTTPServer.Shutdown(ctx)
}

// Addr returns the network address the server is listening on.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.Config.Addr
}
