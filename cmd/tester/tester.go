package main

import (
	"context"
	"log/slog"

	"github.com/erendogan51/immo-scraper/pkg/config"
	"github.com/erendogan51/immo-scraper/pkg/db/sql"
	"github.com/erendogan51/immo-scraper/pkg/logger"
	"github.com/erendogan51/immo-scraper/pkg/monitoring"
	"github.com/erendogan51/immo-scraper/pkg/scraper"
)

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

	_, err = sql.GetDBPool(ctx, cfg.DB)
	if err != nil {
		panic(err)
	}

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

	scraper, err := scraper.New(ctx, []string{
		//"https://www.willhaben.at/iad/immobilien/eigentumswohnung/wien",
		"https://www.willhaben.at/iad/immobilien/immobilien/angebote?sfId=02595528-be0a-464f-a1e4-d9a704e9b1ab",
	})
	if err != nil {
		panic(err)
	}

	err = scraper.ScrapeWillhabenTargets(ctx)
	if err != nil {
		panic(err)
	}

}
