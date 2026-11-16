package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Address string
	DatabaseURL string
	APIToken string
	LogLevel slog.Level
}

func Load() (Config, error) {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug": level = slog.LevelDebug
	case "warn", "warning": level = slog.LevelWarn
	case "error": level = slog.LevelError
	}
	cfg := Config{
		Address: env("CONTROL_PLANE_ADDR", ":8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		APIToken: strings.TrimSpace(os.Getenv("FLEET_API_TOKEN")),
		LogLevel: level,
	}
	if cfg.DatabaseURL == "" { return Config{}, fmt.Errorf("DATABASE_URL is required") }
	if len(cfg.APIToken) < 32 { return Config{}, fmt.Errorf("FLEET_API_TOKEN must contain at least 32 characters") }
	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" { return value }
	return fallback
}
