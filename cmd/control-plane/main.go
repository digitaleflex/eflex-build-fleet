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
	cfg := config.Load("CONTROL_PLANE_ADDR", ":8080")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := httpserver.New("control-plane", cfg.Address, logger).Run(ctx); err != nil {
		logger.Error("control plane stopped", "error", err)
		os.Exit(1)
	}
}
