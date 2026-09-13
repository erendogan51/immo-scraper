package willhaben

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const execSuccessBody = `{
	"id": 131,
	"description": "Mietwohnungen",
	"heading": "Mietwohnungen in Wien",
	"rowsRequested": 30,
	"rowsFound": 62,
	"rowsReturned": 30,
	"pageRequested": 1,
	"searchDate": "2026-07-20T15:30:36+0200",
	"advertSummaryList": {
		"advertSummary": [
			{
				"id": "972241755",
				"description": "Gartenwohnung",
				"advertStatus": {"id": "active", "description": "aktiv", "statusId": 50},
				"attributes": {
					"attribute": [
						{"name": "PRICE", "values": ["1599"]},
						{"name": "PRICE_FOR_DISPLAY", "values": ["€ 1.599"]},
						{"name": "LOCATION", "values": ["Wien, 20. Bezirk, Brigittenau"]},
						{"name": "SEO_URL", "values": ["immobilien/d/mietwohnungen/wien/972241755/"]}
					]
				}
			}
		]
	}
}`

func TestSearchListings(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotQuery url.Values
	var gotWHClient string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		gotWHClient = r.Header.Get("X-WH-Client")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(execSuccessBody))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	result, err := client.SearchListings(t.Context(), "immobilien/mietwohnungen/wien", url.Values{"NO_OF_ROOMS": {"2"}}, 1)
	if err != nil {
		t.Fatalf("SearchListings returned error: %v", err)
	}

	if gotPath != "/webapi/ad-search/search/atz/seo/immobilien/mietwohnungen/wien" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotQuery.Get("page") != "1" {
		t.Errorf("page = %q, want 1", gotQuery.Get("page"))
	}

	gottenRows, err := strconv.Atoi(gotQuery.Get("rows"))
	require.NoError(t, err)

	if gottenRows != defaultRows {
		t.Errorf("rows = %q, want 30", gotQuery.Get("rows"))
	}
	if gotQuery.Get("NO_OF_ROOMS") != "2" {
		t.Errorf("NO_OF_ROOMS filter not forwarded: %q", gotQuery.Get("NO_OF_ROOMS"))
	}
	if gotQuery.Get("sfId") == "" {
		t.Error("sfId not set")
	}
	if gotWHClient != whClientHeader {
		t.Errorf("X-WH-Client header not sent: %q", gotWHClient)
	}

	if result.RowsFound != 62 {
		t.Errorf("RowsFound = %d, want 62", result.RowsFound)
	}
	if !result.HasMore() {
		t.Error("HasMore() = false, want true (30 of 62 returned)")
	}

	adverts := result.Adverts()
	if len(adverts) != 1 {
		t.Fatalf("len(Adverts()) = %d, want 1", len(adverts))
	}

	ad := adverts[0]
	if got, want := ad.Price(), 1599.0; got != want {
		t.Errorf("Price() = %v, want %v", got, want)
	}
	if got, want := ad.PriceDisplay(), "€ 1.599"; got != want {
		t.Errorf("PriceDisplay() = %q, want %q", got, want)
	}
	if got, want := ad.Location(), "Wien, 20. Bezirk, Brigittenau"; got != want {
		t.Errorf("Location() = %q, want %q", got, want)
	}
	if got, want := ad.URL(), "https://www.willhaben.at/iad/immobilien/d/mietwohnungen/wien/972241755/"; got != want {
		t.Errorf("URL() = %q, want %q", got, want)
	}
}

func TestSearchListingsUnauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, err := client.SearchListings(t.Context(), "immobilien/mietwohnungen/wien", nil, 1)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("SearchListings error = %v, want wrapped ErrUnauthorized", err)
	}
}

func TestSearchListingsServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, err := client.SearchListings(t.Context(), "immobilien/mietwohnungen/wien", nil, 1)
	if err == nil {
		t.Fatal("SearchListings returned no error, want APIError")
	}
}
