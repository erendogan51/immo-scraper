package logger

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/erendogan51/immo-scraper/pkg/config"
)

func SetLogger(cfg config.LoggerConfig) error {
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		return errors.New("unknown name")
	}

	var logger *slog.Logger

	if cfg.LogStyle == config.LogStyleText {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	}

	slog.SetDefault(logger)

	return nil
}
