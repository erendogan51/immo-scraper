package willhaben

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const listingPageHTML = `<!DOCTYPE html><html><head></head><body>
<script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"advertDetails":{
	"id": "773562369",
	"uuid": "48d3f827-209b-42c0-a5a2-b34e9d305250",
	"verticalId": 2,
	"adTypeId": 1,
	"productId": 1,
	"description": "Renovierte Altbauwohnung",
	"startDate": "2026-05-18T00:00:00+0200",
	"endDate": "2027-05-27T00:00:00+0200",
	"publishedDate": "2026-08-03T14:04:00+0200",
	"firstPublishedDate": "2026-05-18T14:04:00+0200",
	"createdDate": "2026-05-18T14:03:00+0200",
	"changedDate": "2026-08-03T14:05:03+0200",
	"advertStatus": {"id": "active", "description": "aktiv", "statusId": 50},
	"attributes": {"attribute": [
		{"name": "PRICE", "values": ["890000"]},
		{"name": "DESCRIPTION", "values": ["Long body text with a closing tag look-alike: <\/not-a-tag>"]}
	]},
	"advertImageList": {"advertImage": [
		{"id": 1, "description": "Küche", "mainImageUrl": "https://cache.willhaben.at/1.jpg", "thumbnailImageUrl": "https://cache.willhaben.at/1_thumb.jpg", "referenceImageUrl": "https://cache.willhaben.at/1_ref.jpg"}
	]},
	"advertAddressDetails": {
		"addressLines": {"value": ["Wien, 09. Bezirk, Alsergrund"]},
		"postCode": "1090",
		"postalName": "Wien, 09. Bezirk, Alsergrund",
		"country": "Österreich",
		"province": "Wien",
		"district": "Wien",
		"municipality": "Wien"
	},
	"seoMetaData": {"canonicalUrl": "https://www.willhaben.at/iad/immobilien/d/eigentumswohnung/wien/773562369/"}
}}}}</script>
</body></html>`

func TestGetListing(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAccept = r.Header.Get("Accept")

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(listingPageHTML))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	listing, err := client.GetListing(t.Context(), "immobilien/d/eigentumswohnung/wien/773562369/")
	if err != nil {
		t.Fatalf("GetListing returned error: %v", err)
	}

	if gotPath != "/iad/immobilien/d/eigentumswohnung/wien/773562369/" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotAccept != "text/html" {
		t.Errorf("Accept header = %q, want text/html", gotAccept)
	}

	if listing.ID != "773562369" {
		t.Errorf("ID = %q, want 773562369", listing.ID)
	}
	if listing.UUID != "48d3f827-209b-42c0-a5a2-b34e9d305250" {
		t.Errorf("UUID = %q", listing.UUID)
	}
	if listing.StartDate != "2026-05-18T00:00:00+0200" {
		t.Errorf("StartDate = %q", listing.StartDate)
	}
	if listing.AdvertAddressDetails.Municipality != "Wien" {
		t.Errorf("Municipality = %q", listing.AdvertAddressDetails.Municipality)
	}
	if listing.SeoMetaData.CanonicalURL != "https://www.willhaben.at/iad/immobilien/d/eigentumswohnung/wien/773562369/" {
		t.Errorf("CanonicalURL = %q", listing.SeoMetaData.CanonicalURL)
	}
	if len(listing.AdvertImageList.AdvertImage) != 1 {
		t.Fatalf("len(AdvertImage) = %d, want 1", len(listing.AdvertImageList.AdvertImage))
	}
	if got, want := listing.AdvertImageList.AdvertImage[0].MainImageURL, "https://cache.willhaben.at/1.jpg"; got != want {
		t.Errorf("MainImageURL = %q, want %q", got, want)
	}
	if got, want := listing.Attributes.Price, 890000.0; got != want {
		t.Errorf("Attributes.Price = %v, want %v", got, want)
	}
	if got, want := listing.Attributes.Extra["DESCRIPTION"][0], "Long body text with a closing tag look-alike: </not-a-tag>"; got != want {
		t.Errorf("Attributes.Extra[DESCRIPTION] = %q, want %q", got, want)
	}
}

func TestGetListingNoNextData(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>not a listing page</body></html>"))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, err := client.GetListing(t.Context(), "immobilien/d/eigentumswohnung/wien/773562369/")
	if err == nil {
		t.Fatal("GetListing returned no error, want one")
	}
}
