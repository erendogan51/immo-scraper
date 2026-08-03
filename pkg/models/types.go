package models

// willhabenBaseURL is the prefix SeoURL values are relative to.
const willhabenBaseURL = "https://www.willhaben.at/iad/"

// AdvertStatus describes the lifecycle state of an advert, e.g. active.
type AdvertStatus struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	StatusID    int    `json:"statusId"`
}

// AdvertSummary is a single listing as returned by a search agent
// execution.
type AdvertSummary struct {
	ID           string       `json:"id"`
	VerticalID   int          `json:"verticalId"`
	AdTypeID     int          `json:"adTypeId"`
	ProductID    int          `json:"productId"`
	AdvertStatus AdvertStatus `json:"advertStatus"`
	Description  string       `json:"description"`
	Attributes   Attributes   `json:"attributes"`
}

// Price returns the raw PRICE attribute, e.g. 1599.
func (a AdvertSummary) Price() float64 { return a.Attributes.Price }

// PriceDisplay returns the human-formatted PRICE_FOR_DISPLAY attribute, e.g.
// "€ 1.599".
func (a AdvertSummary) PriceDisplay() string { return a.Attributes.PriceDisplay }

// Location returns the LOCATION attribute, e.g. "Wien, 20. Bezirk,
// Brigittenau".
func (a AdvertSummary) Location() string { return a.Attributes.Location }

// SeoURL returns the SEO_URL attribute, a path relative to
// "https://www.willhaben.at/iad/". Use URL for the absolute link.
func (a AdvertSummary) SeoURL() string { return a.Attributes.SeoURL }

// URL returns the listing's absolute URL.
func (a AdvertSummary) URL() string { return willhabenBaseURL + a.SeoURL() }
