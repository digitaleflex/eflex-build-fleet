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

// Load keeps the lightweight configuration path used by the Build Agent.
func Load(addressKey, defaultAddress string) Config {
	level := logLevel()
	address := env(addressKey, defaultAddress)
	return Config{Address: address, LogLevel: level}
}

func LoadControlPlane() (Config, error) {
	cfg := Config{
		Address: env("CONTROL_PLANE_ADDR", ":8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		APIToken: strings.TrimSpace(os.Getenv("FLEET_API_TOKEN")),
		LogLevel: logLevel(),
	}
	if cfg.DatabaseURL == "" { return Config{}, fmt.Errorf("DATABASE_URL is required") }
	if len(cfg.APIToken) < 32 { return Config{}, fmt.Errorf("FLEET_API_TOKEN must contain at least 32 characters") }
	return cfg, nil
}

func logLevel() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug": return slog.LevelDebug
	case "warn", "warning": return slog.LevelWarn
	case "error": return slog.LevelError
	default: return slog.LevelInfo
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" { return value }
	return fallback
}
