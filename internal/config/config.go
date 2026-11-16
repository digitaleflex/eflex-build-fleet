package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Address string
	LogLevel slog.Level
}

func Load(addressKey, defaultAddress string) Config {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug": level = slog.LevelDebug
	case "warn", "warning": level = slog.LevelWarn
	case "error": level = slog.LevelError
	}
	address := strings.TrimSpace(os.Getenv(addressKey))
	if address == "" { address = defaultAddress }
	return Config{Address: address, LogLevel: level}
}
