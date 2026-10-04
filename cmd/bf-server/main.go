package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/madith/bf-server/pkg/config"
	"github.com/madith/bf-server/pkg/logger"
	"github.com/madith/bf-server/pkg/server"
)

const banner = `
  _     __                                   
 | |__ / _|___ ___ _ ___ _____ _ _ 
 | '_ \  _(_-</ -_) '_\ V / -_) '_|
 |_.__/_| /__/\___|_|  \_/\___|_|   v1.0.0
 HTTP Web Server powered by Brainfuck
`

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	log := logger.Setup(cfg.LogFormat, cfg.LogLevel, os.Stdout)

	if cfg.LogFormat != "json" {
		fmt.Print(banner)
		fmt.Printf(" [>] Listening on  : http://localhost%s\n", cfg.Addr)
		fmt.Printf(" [>] App Directory : %s\n", cfg.AppDir)
		fmt.Printf(" [>] Memory Size   : %d cells\n", cfg.MemorySize)
		fmt.Printf(" [>] Max Steps     : %d\n", cfg.MaxSteps)
		fmt.Printf(" [>] Dev Mode      : %t\n\n", cfg.DevMode)
	}

	srv, err := server.New(cfg, log)
	if err != nil {
		log.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	// Server runner channel
	serverErrChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			serverErrChan <- err
		}
	}()

	// Signal notification for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		log.Error("Server encountered fatal error", "error", err)
		os.Exit(1)
	case sig := <-quit:
		log.Info("Received shutdown signal", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Error during server shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server stopped successfully")
}
