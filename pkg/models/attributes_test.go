package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func unmarshalAttrs(t *testing.T, wire string) Attributes {
	t.Helper()

	var a Attributes
	if err := json.Unmarshal([]byte(wire), &a); err != nil {
		t.Fatalf("unmarshal attributes: %v", err)
	}

	return a
}

func TestAttributesMerge(t *testing.T) {
	t.Parallel()

	// A search result's attribute bag: has SEO_URL/HEADING (which a listing
	// detail page never carries), and an older PRICE.
	summary := unmarshalAttrs(t, `{"attribute": [
		{"name": "SEO_URL", "values": ["immobilien/d/eigentumswohnung/wien/773562369/"]},
		{"name": "HEADING", "values": ["Altbauwohnung"]},
		{"name": "PRICE", "values": ["850000"]},
		{"name": "LOCATION", "values": ["Wien, 09. Bezirk, Alsergrund"]}
	]}`)

	// A listing detail page's attribute bag: no SEO_URL/HEADING/LOCATION,
	// an updated PRICE, and a DESCRIPTION attribute the parser doesn't
	// recognize by name (falls into Extra).
	detail := unmarshalAttrs(t, `{"attribute": [
		{"name": "PRICE", "values": ["890000"]},
		{"name": "DESCRIPTION", "values": ["Long body text"]}
	]}`)

	merged := summary.Merge(detail)

	if merged.SeoURL != "immobilien/d/eigentumswohnung/wien/773562369/" {
		t.Errorf("SeoURL = %q, want the summary-only value preserved", merged.SeoURL)
	}
	if merged.Heading != "Altbauwohnung" {
		t.Errorf("Heading = %q, want the summary-only value preserved", merged.Heading)
	}
	if merged.Location != "Wien, 09. Bezirk, Alsergrund" {
		t.Errorf("Location = %q, want the summary-only value preserved", merged.Location)
	}
	if merged.Price != 890000 {
		t.Errorf("Price = %v, want detail's value (890000) to win over summary's (850000)", merged.Price)
	}
	if got, want := merged.Extra["DESCRIPTION"], []string{"Long body text"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Extra[DESCRIPTION] = %v, want %v", got, want)
	}
}

func TestAttributesMergeEmptyOther(t *testing.T) {
	t.Parallel()

	summary := unmarshalAttrs(t, `{"attribute": [{"name": "PRICE", "values": ["1599"]}]}`)

	merged := summary.Merge(Attributes{})

	if merged.Price != 1599 {
		t.Errorf("Price = %v, want 1599 unchanged when merging in an empty Attributes", merged.Price)
	}
}
