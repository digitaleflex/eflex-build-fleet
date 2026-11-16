package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/digitaleflex/eflex-build-fleet/internal/config"
	"github.com/digitaleflex/eflex-build-fleet/internal/httpserver"
)

func main() {
	cfg := config.Load("BUILD_AGENT_ADDR", ":8081")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := httpserver.New("build-agent", cfg.Address, logger).Run(ctx); err != nil {
		logger.Error("build agent stopped", "error", err)
		os.Exit(1)
	}
}
