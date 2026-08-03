package scrapper

import (
	"context"
	"log/slog"
	"time"

	"github.com/erendogan51/immo-scrapper/pkg/willhaben"
)

type Scrapper struct {
	WillhabenClient *willhaben.Client
}

func New(ctx context.Context, willhabenSearchURLs []string) (*Scrapper, error) {
	client := willhaben.NewClient(willhabenSearchURLs)

	return &Scrapper{
		WillhabenClient: client,
	}, nil
}

func (s *Scrapper) Start(ctx context.Context) error {
	willhabenScrapeTicker := time.NewTicker(5 * time.Minute)
	defer willhabenScrapeTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-willhabenScrapeTicker.C:
			err := s.ScrapeWillhabenTargets(ctx)
			if err != nil {
				slog.Error("error scraping targets", "error", err)
			}
		}
	}
}

func (s *Scrapper) ScrapeWillhabenTargets(ctx context.Context) error {
	s.WillhabenClient.SearchListings()
}
