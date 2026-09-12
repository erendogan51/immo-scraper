package willhaben

import (
	"encoding/json"
	"testing"

	"github.com/erendogan51/immo-scraper/pkg/models"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/shopspring/decimal"
)

func decimalString(t *testing.T, d *pgxdecimal.Decimal) string {
	t.Helper()

	if d == nil {
		return ""
	}

	return decimal.Decimal(*d).String()
}

func testAdvert(t *testing.T) models.AdvertSummary {
	t.Helper()

	var ad models.AdvertSummary
	body := `{
		"id": "773562369",
		"verticalId": 2,
		"adTypeId": 1,
		"productId": 1,
		"description": "Renovierte Altbauwohnung",
		"advertStatus": {"id": "active", "description": "aktiv", "statusId": 50},
		"attributes": {"attribute": [
			{"name": "PRICE", "values": ["850000"]},
			{"name": "PRICE_FOR_DISPLAY", "values": ["€ 850.000"]},
			{"name": "LOCATION", "values": ["Wien, 09. Bezirk, Alsergrund"]},
			{"name": "SEO_URL", "values": ["immobilien/d/eigentumswohnung/wien/773562369/"]},
			{"name": "HEADING", "values": ["Altbauwohnung"]}
		]}
	}`
	if err := json.Unmarshal([]byte(body), &ad); err != nil {
		t.Fatalf("unmarshal advert: %v", err)
	}

	return ad
}

func TestToUpsertParamsSearchOnly(t *testing.T) {
	t.Parallel()

	ad := testAdvert(t)

	params, err := toUpsertParams(ad, nil, "immobilien/eigentumswohnung/wien")
	if err != nil {
		t.Fatalf("toUpsertParams returned error: %v", err)
	}

	if params.AdID != "773562369" {
		t.Errorf("AdID = %q, want 773562369", params.AdID)
	}
	if got := decimalString(t, params.Price); got != "850000" {
		t.Errorf("Price = %v, want 850000", got)
	}
	if !params.Location.Valid || params.Location.String != "Wien, 09. Bezirk, Alsergrund" {
		t.Errorf("Location = %+v", params.Location)
	}
	if params.AdUuid.Valid {
		t.Error("AdUuid.Valid = true, want false without a detail page")
	}
	if string(params.Images) != "[]" {
		t.Errorf("Images = %s, want []", params.Images)
	}
	if !params.SearchPath.Valid || params.SearchPath.String != "immobilien/eigentumswohnung/wien" {
		t.Errorf("SearchPath = %+v", params.SearchPath)
	}
}

func TestToUpsertParamsWithDetail(t *testing.T) {
	t.Parallel()

	ad := testAdvert(t)

	detail := &ListingResult{
		ID:                 "773562369",
		UUID:               "48d3f827-209b-42c0-a5a2-b34e9d305250",
		StartDate:          "2026-05-18T00:00:00+0200",
		FirstPublishedDate: "2026-05-18T14:04:00+0200",
	}
	detail.SeoMetaData.CanonicalURL = "https://www.willhaben.at/iad/immobilien/d/eigentumswohnung/wien/773562369/"
	detail.AdvertAddressDetails.Municipality = "Wien"
	if err := json.Unmarshal([]byte(`{"attribute": [
		{"name": "PRICE", "values": ["890000"]},
		{"name": "DESCRIPTION", "values": ["Long body text"]}
	]}`), &detail.Attributes); err != nil {
		t.Fatalf("unmarshal detail attributes: %v", err)
	}
	detail.AdvertImageList.AdvertImage = []ListingImage{
		{ID: 1, MainImageURL: "https://cache.willhaben.at/1.jpg"},
	}

	params, err := toUpsertParams(ad, detail, "immobilien/eigentumswohnung/wien")
	if err != nil {
		t.Fatalf("toUpsertParams returned error: %v", err)
	}

	if !params.AdUuid.Valid || params.AdUuid.UUID.String() != "48d3f827-209b-42c0-a5a2-b34e9d305250" {
		t.Errorf("AdUuid = %+v", params.AdUuid)
	}
	if !params.StartDate.Valid {
		t.Error("StartDate.Valid = false, want true")
	}
	if !params.Municipality.Valid || params.Municipality.String != "Wien" {
		t.Errorf("Municipality = %+v", params.Municipality)
	}
	if !params.CanonicalUrl.Valid {
		t.Error("CanonicalUrl.Valid = false, want true")
	}
	if string(params.Images) == "[]" {
		t.Error("Images = [], want the detail page's image gallery")
	}

	// The merged attribute bag should keep the summary-only SEO_URL/HEADING
	// alongside the detail's updated PRICE and its unrecognized DESCRIPTION.
	if got := decimalString(t, params.Price); got != "890000" {
		t.Errorf("Price = %v, want detail's 890000 to win over summary's 850000", got)
	}
	if !params.Heading.Valid || params.Heading.String != "Altbauwohnung" {
		t.Errorf("Heading = %+v, want the summary-only value preserved", params.Heading)
	}

	var storedAttrs models.Attributes
	if err := json.Unmarshal(params.Attributes, &storedAttrs); err != nil {
		t.Fatalf("unmarshal stored attributes: %v", err)
	}
	if got, want := storedAttrs.Extra["DESCRIPTION"], []string{"Long body text"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("stored Extra[DESCRIPTION] = %v, want %v", got, want)
	}
}
