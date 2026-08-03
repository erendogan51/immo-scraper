package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erendogan51/immo-scrapper/pkg/config"
	"github.com/erendogan51/immo-scrapper/pkg/logger"
	"github.com/erendogan51/immo-scrapper/pkg/monitoring"
)

const requestTimeout = 60 * time.Second

func main() {
	ctx := context.Background()

	cfg, err := config.GetConfig(ctx)
	if err != nil {
		panic(err)
	}

	err = logger.SetLogger(cfg.Logger)
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal)
	go func() {
		err := monitoring.Expose(cfg.MetricsPort)
		if err != nil {
			slog.Error(
				"failed to expose metrics port",
				"error",
				err,
			)
		}
	}()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) //nolint:govet, staticcheck

	sig := <-sigs
	slog.Info("Received signal", "signal", sig)
	slog.Info("finished process")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
