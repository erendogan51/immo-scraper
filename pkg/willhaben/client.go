// Package willhaben provides a client for willhaben.at's public search API,
// used to run listing searches and return matching adverts as structured
// data instead of scraped HTML. It's the same webapi/ad-search endpoint
// willhaben's own search results pages call (confirmed via a HAR capture of
// an anonymous, logged-out browsing session), so no account or login is
// required.
package willhaben

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://www.willhaben.at"
	defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64)"
	defaultTimeout   = 30 * time.Second

	// whClientHeader mirrors the X-WH-Client value willhaben's own frontend
	// sends; the search API rejects requests missing it.
	whClientHeader = "api@willhaben.at;responsive_web;server;1.0.0;desktop"
)

// Client talks to willhaben's public search API (www.willhaben.at/webapi/ad-search),
// the same one willhaben's own search results pages use to render listings
// for anonymous visitors. No authentication is needed.
//
// Create one with NewClient.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	searchURLs []string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the default http.Client used for requests.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithBaseURL overrides the default base URL (https://www.willhaben.at).
// Mainly useful for pointing at a mock server in tests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		if userAgent != "" {
			c.userAgent = userAgent
		}
	}
}

// NewClient creates a Client for willhaben's public search API.
func NewClient(searchURLs []string, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    defaultBaseURL,
		userAgent:  defaultUserAgent,
		searchURLs: searchURLs,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// parseSearchURL splits a willhaben search results URL into the SEO path
// SearchListings expects (the path segment after "/iad/") and its query
// parameters (the search's active filters).
func parseSearchURL(rawURL string) (seoPath string, params url.Values, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}

	seoPath = strings.TrimPrefix(u.Path, "/iad/")
	seoPath = strings.Trim(seoPath, "/")
	if seoPath == "" {
		return "", nil, fmt.Errorf("no search path found in URL %q", rawURL)
	}

	return seoPath, u.Query(), nil
}

// resolvePage decides which result page to request: the -page flag if set
// (nonzero), otherwise the "page" query parameter carried by -search-url
// (e.g. from pasting a URL like ".../wien?page=2" copied while paging
// through results in the browser), otherwise page 1.
func resolvePage(params url.Values) int {
	if raw := params.Get("page"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			return n
		}
	}

	return 1
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values) (*http.Request, error) {
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("willhaben: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-WH-Client", whClientHeader)

	return req, nil
}

// do executes req and decodes a JSON response body into out. out may be nil
// to discard the body.
func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("willhaben: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("willhaben: read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("willhaben: decode response: %w", err)
	}

	return nil
}
