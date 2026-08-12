package willhaben

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/erendogan51/immo-scrapper/pkg/db/sql"
	"github.com/erendogan51/immo-scrapper/pkg/db/sql/db"
	"github.com/google/uuid"

	"github.com/erendogan51/immo-scrapper/pkg/models"
)

// searchDateLayout matches willhaben's date format, e.g.
// "2026-07-20T15:30:36+0200" - not quite RFC3339 (no colon in the offset),
// so it can't be unmarshaled into time.Time directly.
const searchDateLayout = "2006-01-02T15:04:05-0700"

// searchPathPrefix is prepended to a search's SEO path to build the public
// search API path, e.g. "immobilien/mietwohnungen/wien" becomes
// "/webapi/ad-search/search/atz/seo/immobilien/mietwohnungen/wien". This
// matches the endpoint willhaben's own search results page calls for
// anonymous, logged-out visitors.
const searchPathPrefix = "/webapi/ad-search/search/atz/seo/"

// defaultRows is the page size requested when none is otherwise specified,
// matching willhaben's own default.
const defaultRows = 30

// pageDelay is how long scrapeTarget waits between requesting successive
// search result pages.
const pageDelay = 500 * time.Millisecond

// pageDelayJitterMs is added to pageDelay, randomized, to avoid a
// mechanically regular request cadence.
const pageDelayJitterMs = 200

// detailDelay is how long scrapeAdvert waits before fetching a listing's
// detail page. It's longer than pageDelay since a detail page is a full
// HTML fetch (heavier than the JSON search API) and is requested once per
// advert rather than once per page of up to defaultRows adverts.
const detailDelay = time.Second

// detailDelayJitterMs is added to detailDelay, randomized.
const detailDelayJitterMs = 1000

// politeDelay sleeps for base plus a random jitter in [0, jitterMs)
// milliseconds, to space out requests to willhaben.
func politeDelay(base time.Duration, jitterMs int) {
	time.Sleep(base + time.Duration(rand.Intn(jitterMs))*time.Millisecond)
}

// SearchAgentResult is the response of executing a saved search agent.
type SearchAgentResult struct {
	ID                      int    `json:"id"`
	Description             string `json:"description"`
	Heading                 string `json:"heading"`
	VerticalID              int    `json:"verticalId"`
	SearchID                int    `json:"searchId"`
	RowsRequested           int    `json:"rowsRequested"`
	RowsFound               int    `json:"rowsFound"`
	RowsReturned            int    `json:"rowsReturned"`
	PageRequested           int    `json:"pageRequested"`
	SearchDate              string `json:"searchDate"`
	LastUserAlertViewedDate string `json:"lastUserAlertViewedDate"`
	NewAdsSeparatorPosition int    `json:"newAdsSeparatorPosition"`
	AdvertSummaryList       struct {
		AdvertSummary []models.AdvertSummary `json:"advertSummary"`
	} `json:"advertSummaryList"`
}

// Adverts returns the listings contained in this result page.
func (r *SearchAgentResult) Adverts() []models.AdvertSummary {
	return r.AdvertSummaryList.AdvertSummary
}

// HasMore reports whether additional result pages exist beyond this one.
func (r *SearchAgentResult) HasMore() bool {
	return r.PageRequested*r.RowsRequested < r.RowsFound
}

// ParseSearchDate parses SearchDate into a time.Time.
func (r *SearchAgentResult) ParseSearchDate() (time.Time, error) {
	return time.Parse(searchDateLayout, r.SearchDate)
}

// SearchListings runs a willhaben search and returns its matching listings,
// using the same public search API willhaben's own search results pages
// call to render listings for anonymous visitors — no login is required.
//
// seoPath is the path segment of a willhaben search results URL, e.g.
// "immobilien/mietwohnungen/wien" for
// https://www.willhaben.at/iad/immobilien/mietwohnungen/wien/. It also
// accepts the same path with additional filters applied (as saved by a
// search agent or narrowed down in the browser), in which case those
// filters should be passed via params exactly as they appear in the
// search URL's query string (e.g. "PRICE_TO", "NO_OF_ROOMS"); pass nil for
// an unfiltered search. page is 1-indexed; values below 1 are treated as 1.
func (c *Client) SearchListings(ctx context.Context, seoPath string, params url.Values, page int) (*SearchAgentResult, error) {
	if page < 1 {
		page = 1
	}

	query := url.Values{}
	for k, v := range params {
		query[k] = v
	}
	// sfId identifies the browsing session to willhaben's search API; it's
	// minted client-side (the server never issues it), so any value works.
	query.Set("sfId", uuid.NewString())
	query.Set("page", strconv.Itoa(page))
	if query.Get("rows") == "" {
		query.Set("rows", strconv.Itoa(defaultRows))
	}

	path := searchPathPrefix + strings.TrimPrefix(seoPath, "/")

	req, err := c.newRequest(ctx, http.MethodGet, path, query)
	if err != nil {
		return nil, err
	}

	var result SearchAgentResult
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("willhaben: search %q: %w", seoPath, err)
	}

	return &result, nil
}

func (c *Client) ScrapeListings(ctx context.Context) error {
	var multiErr error
	for _, searchURL := range c.searchURLs {
		err := c.scrapeTarget(ctx, searchURL)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to scrape %s", searchURL), "error", err)
			if multiErr == nil {
				multiErr = err
			} else {
				multiErr = fmt.Errorf("%w: %w", multiErr, err)
			}
		}
	}

	return multiErr
}

func (c *Client) scrapeTarget(ctx context.Context, searchURL string) error {
	seoPath, params, err := parseSearchURL(searchURL)
	if err != nil {
		return fmt.Errorf("parse search-url: %w", err)
	}

	queries, err := sql.GetDBQueries(ctx)
	if err != nil {
		return err
	}

	for page := resolvePage(params); ; page++ {
		result, err := c.SearchListings(ctx, seoPath, params, page)
		if err != nil {
			return fmt.Errorf("search listings: %w", err)
		}

		adverts := result.Adverts()
		slog.Info(fmt.Sprintf("scraping %d listings, page %d", len(adverts), page))

		for _, ad := range adverts {
			// willhaben's search API has no filter to request private-seller
			// listings only (confirmed against captured search traffic), so
			// non-private adverts are dropped here instead - before the
			// detail-page fetch and its delay, since we're not going to
			// persist them anyway.
			if !ad.Attributes.IsPrivate {
				continue
			}

			politeDelay(detailDelay, detailDelayJitterMs)

			if err := c.scrapeAdvert(ctx, queries, ad, seoPath); err != nil {
				slog.Error(fmt.Sprintf("failed to scrape advert %s", ad.ID), "error", err)
			}
		}

		if !result.HasMore() {
			return nil
		}

		politeDelay(pageDelay, pageDelayJitterMs)
	}
}

// scrapeAdvert fetches ad's listing detail page and upserts the combined
// search-result and detail data. A detail-fetch failure isn't fatal to the
// advert as a whole - it's still stored using only the data already
// available from the search result, so a scrape isn't held back by one
// listing's detail page (temporarily) failing to load.
func (c *Client) scrapeAdvert(ctx context.Context, queries *db.Queries, ad models.AdvertSummary, searchPath string) error {
	detail, err := c.GetListing(ctx, ad.SeoURL())
	if err != nil {
		slog.Error(fmt.Sprintf("failed to fetch listing detail for advert %s", ad.ID), "error", err)
		detail = nil
	}

	params, err := toUpsertParams(ad, detail, searchPath)
	if err != nil {
		return fmt.Errorf("map advert %s: %w", ad.ID, err)
	}

	if err := queries.UpsertListing(ctx, params); err != nil {
		return fmt.Errorf("upsert advert %s: %w", ad.ID, err)
	}

	return nil
}
