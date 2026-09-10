package willhaben

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erendogan51/immo-scraper/pkg/models"
)

// ListingResult is a single listing as returned by willhaben's listing
// detail page, embedded in that page's "__NEXT_DATA__" script tag at
// props.pageProps.advertDetails. Unlike SearchAgentResult - which wraps a
// page of summarized adverts plus search metadata - ListingResult describes
// exactly one advert, with detail not present in a search result: publish
// history, the full image gallery, structured address data, and the
// canonical listing URL. AdvertStatus and Attributes have the same wire
// shape as in a search result, so those types are reused here.
type ListingResult struct {
	ID                 string              `json:"id"`
	UUID               string              `json:"uuid"`
	VerticalID         int                 `json:"verticalId"`
	AdTypeID           int                 `json:"adTypeId"`
	ProductID          int                 `json:"productId"`
	Description        string              `json:"description"`
	StartDate          string              `json:"startDate"`
	EndDate            string              `json:"endDate"`
	PublishedDate      string              `json:"publishedDate"`
	FirstPublishedDate string              `json:"firstPublishedDate"`
	CreatedDate        string              `json:"createdDate"`
	ChangedDate        string              `json:"changedDate"`
	AdvertStatus       models.AdvertStatus `json:"advertStatus"`
	Attributes         models.Attributes   `json:"attributes"`
	AdvertImageList    struct {
		AdvertImage []ListingImage `json:"advertImage"`
	} `json:"advertImageList"`
	AdvertAddressDetails ListingAddress `json:"advertAddressDetails"`
	SeoMetaData          struct {
		CanonicalURL string `json:"canonicalUrl"`
	} `json:"seoMetaData"`
}

// ListingImage is a single image in a listing's gallery.
type ListingImage struct {
	ID                int    `json:"id"`
	Description       string `json:"description"`
	MainImageURL      string `json:"mainImageUrl"`
	ThumbnailImageURL string `json:"thumbnailImageUrl"`
	ReferenceImageURL string `json:"referenceImageUrl"`
}

// ListingAddress is the structured postal address of a listing.
type ListingAddress struct {
	AddressLines struct {
		Value []string `json:"value"`
	} `json:"addressLines"`
	PostCode     string `json:"postCode"`
	PostalName   string `json:"postalName"`
	Country      string `json:"country"`
	Province     string `json:"province"`
	District     string `json:"district"`
	Municipality string `json:"municipality"`
}

// nextDataOpenTag and nextDataCloseTag bracket the JSON payload Next.js
// embeds in every server-rendered page - including willhaben's listing
// detail pages - as <script id="__NEXT_DATA__" type="application/json">.
// Next.js escapes forward slashes inside that JSON as "\/", so an
// attribute value can't prematurely close the tag.
const (
	nextDataOpenTag  = `<script id="__NEXT_DATA__" type="application/json">`
	nextDataCloseTag = `</script>`
)

// GetListing fetches a listing's detail page and returns the single advert
// embedded in it. seoURL is a listing's SEO_URL attribute (as returned in a
// search result's Attributes), relative to "https://www.willhaben.at/iad/".
func (c *Client) GetListing(ctx context.Context, seoURL string) (*ListingResult, error) {
	path := "/iad/" + strings.TrimPrefix(seoURL, "/")

	req, err := c.newHTMLRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	body, err := c.doHTML(req)
	if err != nil {
		return nil, fmt.Errorf("willhaben: get listing %q: %w", seoURL, err)
	}

	listing, err := parseListingHTML(body)
	if err != nil {
		return nil, fmt.Errorf("willhaben: get listing %q: %w", seoURL, err)
	}

	return listing, nil
}

// parseListingHTML extracts the advert embedded in a listing detail page's
// "__NEXT_DATA__" script tag, at props.pageProps.advertDetails.
func parseListingHTML(html []byte) (*ListingResult, error) {
	page := string(html)

	start := strings.Index(page, nextDataOpenTag)
	if start == -1 {
		return nil, fmt.Errorf("willhaben: __NEXT_DATA__ script not found in listing page")
	}
	start += len(nextDataOpenTag)

	end := strings.Index(page[start:], nextDataCloseTag)
	if end == -1 {
		return nil, fmt.Errorf("willhaben: __NEXT_DATA__ script has no closing tag")
	}

	var wire struct {
		Props struct {
			PageProps struct {
				AdvertDetails ListingResult `json:"advertDetails"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	if err := json.Unmarshal([]byte(page[start:start+end]), &wire); err != nil {
		return nil, fmt.Errorf("willhaben: decode __NEXT_DATA__: %w", err)
	}

	return &wire.Props.PageProps.AdvertDetails, nil
}
