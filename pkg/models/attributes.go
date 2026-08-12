// Package models contains typed domain models for willhaben listing data.
package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Coordinates is a WGS84 latitude/longitude pair, as parsed from
// willhaben's "lat,lon" COORDINATES attribute.
type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// Attributes is the typed form of willhaben's attribute bag: the
// name/values pairs attached to every advert summary. willhaben models
// essentially all listing detail this way instead of using dedicated JSON
// fields; Attributes gives the fields observed in real search results a
// concrete Go type.
//
// Attributes unmarshals from and marshals to willhaben's wire format,
// {"attribute": [{"name": "...", "values": ["..."]}, ...]}, so it can be
// used as a drop-in replacement wherever that shape appears. Any attribute
// willhaben sends that isn't recognized below - or whose value can't be
// parsed into its expected type - is preserved in Extra instead of being
// silently dropped.
type Attributes struct {
	// Identifiers
	AdID           string
	AdUUID         string
	AdTypeID       string
	OrgID          string
	OrgUUID        string
	OrgName        string
	ProductID      string
	ProjectID      string
	PropertyTypeID string
	LocationID     string
	CategoryTreeID string
	AdvertiserRef  string
	UnitTitle      string
	UnitNumber     string

	// Listing content
	Heading          string
	Body             string
	PropertyType     string
	PropertyTypeFlat bool
	RoomsLayout      string // e.g. "3X3"
	NumberOfRooms    float64
	Floor            int
	NumberOfChildren int

	// Location
	Address         string
	Location        string
	District        string
	State           string
	Country         string
	PostCode        string
	LocationQuality float64
	Coordinates     Coordinates

	// Size, in square meters
	EstateSizeSqm     int
	LivingAreaSqm     int
	UsableAreaSqm     int
	FreeAreaType      string
	FreeAreaTypeNames []string
	FreeAreaTotalSqm  int

	// Price
	Price                      float64
	PriceDisplay               string
	PricePerSqm                float64
	PricePerSqmDisplay         string
	PricePerSqmDisplayWithUnit string
	RentPerMonth               float64

	// Media
	MainImage          string
	ImageURLs          []string
	ImageDescription   string
	AdSearchResultLogo string
	VirtualViewLink    string

	// Flags & metadata
	IsPrivate               bool
	IsBumped                bool
	UpsellingAdSearchResult bool
	Published               time.Time
	SeoURL                  string
	EstatePreference        []string

	// Extra holds attribute values that weren't recognized (or couldn't be
	// parsed into their expected type) by name, exactly as willhaben sent
	// them, keyed by willhaben's attribute name.
	Extra map[string][]string
}

// wireAttribute is a single named attribute as sent by willhaben.
type wireAttribute struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// UnmarshalJSON implements json.Unmarshaler, decoding willhaben's
// {"attribute": [{"name": "...", "values": ["..."]}, ...]} wire format.
func (a *Attributes) UnmarshalJSON(data []byte) error {
	var wire struct {
		Attribute []wireAttribute `json:"attribute"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("models: unmarshal attributes: %w", err)
	}

	*a = Attributes{}

	for _, attr := range wire.Attribute {
		if a.set(attr.Name, attr.Values) {
			continue
		}
		if a.Extra == nil {
			a.Extra = make(map[string][]string)
		}
		a.Extra[attr.Name] = attr.Values
	}

	return nil
}

// set applies a single willhaben attribute to the corresponding field. It
// returns false if the name is unrecognized or its value can't be parsed
// into the expected type, in which case the caller falls back to Extra.
func (a *Attributes) set(name string, values []string) bool {
	var first string
	if len(values) > 0 {
		first = values[0]
	}

	switch name {
	case "ADID":
		a.AdID = first
	case "AD_UUID":
		a.AdUUID = first
	case "ADTYPE_ID":
		a.AdTypeID = first
	case "ORGID":
		a.OrgID = first
	case "ORG_UUID":
		a.OrgUUID = first
	case "ORGNAME":
		a.OrgName = first
	case "PRODUCT_ID":
		a.ProductID = first
	case "PROJECT_ID":
		a.ProjectID = first
	case "PROPERTY_TYPE_ID":
		a.PropertyTypeID = first
	case "LOCATION_ID":
		a.LocationID = first
	case "categorytreeids":
		a.CategoryTreeID = first
	case "ADVERTISER_REF":
		a.AdvertiserRef = first
	case "UNIT_TITLE":
		a.UnitTitle = first
	case "UNIT_NUMBER":
		a.UnitNumber = first

	case "HEADING":
		a.Heading = first
	case "BODY_DYN":
		a.Body = first
	case "PROPERTY_TYPE":
		a.PropertyType = first
	case "PROPERTY_TYPE_FLAT":
		v, ok := parseBool(first)
		if !ok {
			return false
		}
		a.PropertyTypeFlat = v
	case "ROOMS":
		a.RoomsLayout = first
	case "NUMBER_OF_ROOMS":
		v, ok := parseFloat(first)
		if !ok {
			return false
		}
		a.NumberOfRooms = v
	case "FLOOR":
		v, ok := parseInt(first)
		if !ok {
			return false
		}
		a.Floor = v
	case "NUMBER_OF_CHILDREN":
		v, ok := parseInt(first)
		if !ok {
			return false
		}
		a.NumberOfChildren = v

	case "ADDRESS":
		a.Address = first
	case "LOCATION":
		a.Location = first
	case "DISTRICT":
		a.District = first
	case "STATE":
		a.State = first
	case "COUNTRY":
		a.Country = first
	case "POSTCODE":
		a.PostCode = first
	case "LOCATION_QUALITY":
		v, ok := parseFloat(first)
		if !ok {
			return false
		}
		a.LocationQuality = v
	case "COORDINATES":
		c, ok := parseCoordinates(first)
		if !ok {
			return false
		}
		a.Coordinates = c

	case "ESTATE_SIZE":
		v, ok := parseInt(first)
		if !ok {
			return false
		}
		a.EstateSizeSqm = v
	case "ESTATE_SIZE/LIVING_AREA":
		v, ok := parseInt(first)
		if !ok {
			return false
		}
		a.LivingAreaSqm = v
	case "ESTATE_SIZE/USEABLE_AREA":
		v, ok := parseInt(first)
		if !ok {
			return false
		}
		a.UsableAreaSqm = v
	case "FREE_AREA_TYPE":
		a.FreeAreaType = first
	case "FREE_AREA_TYPE_NAME":
		a.FreeAreaTypeNames = values
	case "FREE_AREA/FREE_AREA_AREA_TOTAL":
		v, ok := parseInt(first)
		if !ok {
			return false
		}
		a.FreeAreaTotalSqm = v

	case "PRICE":
		v, ok := parseFloat(first)
		if !ok {
			return false
		}
		a.Price = v
	case "PRICE_FOR_DISPLAY":
		a.PriceDisplay = first
	case "PRICE/SQUARE_METER":
		v, ok := parseFloat(first)
		if !ok {
			return false
		}
		a.PricePerSqm = v
	case "PRICE/SQUARE_METER_FOR_DISPLAY":
		a.PricePerSqmDisplay = first
	case "PRICE/SQUARE_METER_FOR_DISPLAY_WITH_UNIT":
		a.PricePerSqmDisplayWithUnit = first
	case "RENT/PER_MONTH_LETTINGS":
		v, ok := parseFloat(first)
		if !ok {
			return false
		}
		a.RentPerMonth = v

	case "MMO":
		a.MainImage = first
	case "ALL_IMAGE_URLS":
		a.ImageURLs = splitNonEmpty(first, ";")
	case "imagedescription":
		a.ImageDescription = first
	case "AD_SEARCHRESULT_LOGO":
		a.AdSearchResultLogo = first
	case "VIRTUAL_VIEW_LINK":
		a.VirtualViewLink = first

	case "ISPRIVATE":
		v, ok := parseBool(first)
		if !ok {
			return false
		}
		a.IsPrivate = v
	case "IS_BUMPED":
		v, ok := parseBool(first)
		if !ok {
			return false
		}
		a.IsBumped = v
	case "UPSELLING_AD_SEARCHRESULT":
		v, ok := parseBool(first)
		if !ok {
			return false
		}
		a.UpsellingAdSearchResult = v
	case "PUBLISHED_String":
		t, err := time.Parse(time.RFC3339, first)
		if err != nil {
			return false
		}
		a.Published = t
	case "PUBLISHED":
		// PUBLISHED_String (RFC3339) is preferred when present; only fall
		// back to this epoch-millisecond value if it hasn't been set yet.
		// Attribute order in the wire format isn't guaranteed, so once set
		// this is a no-op rather than a downgrade.
		if !a.Published.IsZero() {
			break
		}
		ms, err := strconv.ParseInt(first, 10, 64)
		if err != nil {
			return false
		}
		a.Published = time.UnixMilli(ms)
	case "SEO_URL":
		a.SeoURL = first
	case "ESTATE_PREFERENCE":
		a.EstatePreference = splitNonEmpty(first, ",")

	default:
		return false
	}

	return true
}

// MarshalJSON implements json.Marshaler, encoding Attributes back into
// willhaben's {"attribute": [{"name": "...", "values": ["..."]}, ...]} wire
// format.
func (a *Attributes) MarshalJSON() ([]byte, error) {
	wire := struct {
		Attribute []wireAttribute `json:"attribute"`
	}{Attribute: a.wire()}

	return json.Marshal(wire)
}

// wire rebuilds the willhaben wire-format attribute list ([]wireAttribute)
// from a's typed fields plus Extra. Factored out of MarshalJSON so Merge can
// reuse it without going through JSON.
func (a *Attributes) wire() []wireAttribute {
	var attrs []wireAttribute

	add := func(name, value string) {
		if value == "" {
			return
		}
		attrs = append(attrs, wireAttribute{Name: name, Values: []string{value}})
	}
	addValues := func(name string, values []string) {
		if len(values) == 0 {
			return
		}
		attrs = append(attrs, wireAttribute{Name: name, Values: values})
	}
	addInt := func(name string, value int) {
		if value != 0 {
			add(name, strconv.Itoa(value))
		}
	}
	addFloat := func(name string, value float64) {
		if value != 0 {
			add(name, strconv.FormatFloat(value, 'f', -1, 64))
		}
	}
	addBool := func(name string, value bool) {
		if value {
			add(name, strconv.FormatBool(value))
		}
	}

	add("ADID", a.AdID)
	add("AD_UUID", a.AdUUID)
	add("ADTYPE_ID", a.AdTypeID)
	add("ORGID", a.OrgID)
	add("ORG_UUID", a.OrgUUID)
	add("ORGNAME", a.OrgName)
	add("PRODUCT_ID", a.ProductID)
	add("PROJECT_ID", a.ProjectID)
	add("PROPERTY_TYPE_ID", a.PropertyTypeID)
	add("LOCATION_ID", a.LocationID)
	add("categorytreeids", a.CategoryTreeID)
	add("ADVERTISER_REF", a.AdvertiserRef)
	add("UNIT_TITLE", a.UnitTitle)
	add("UNIT_NUMBER", a.UnitNumber)

	add("HEADING", a.Heading)
	add("BODY_DYN", a.Body)
	add("PROPERTY_TYPE", a.PropertyType)
	addBool("PROPERTY_TYPE_FLAT", a.PropertyTypeFlat)
	add("ROOMS", a.RoomsLayout)
	addFloat("NUMBER_OF_ROOMS", a.NumberOfRooms)
	addInt("FLOOR", a.Floor)
	addInt("NUMBER_OF_CHILDREN", a.NumberOfChildren)

	add("ADDRESS", a.Address)
	add("LOCATION", a.Location)
	add("DISTRICT", a.District)
	add("STATE", a.State)
	add("COUNTRY", a.Country)
	add("POSTCODE", a.PostCode)
	addFloat("LOCATION_QUALITY", a.LocationQuality)
	if a.Coordinates != (Coordinates{}) {
		lat := strconv.FormatFloat(a.Coordinates.Latitude, 'f', -1, 64)
		lon := strconv.FormatFloat(a.Coordinates.Longitude, 'f', -1, 64)
		add("COORDINATES", lat+","+lon)
	}

	addInt("ESTATE_SIZE", a.EstateSizeSqm)
	addInt("ESTATE_SIZE/LIVING_AREA", a.LivingAreaSqm)
	addInt("ESTATE_SIZE/USEABLE_AREA", a.UsableAreaSqm)
	add("FREE_AREA_TYPE", a.FreeAreaType)
	addValues("FREE_AREA_TYPE_NAME", a.FreeAreaTypeNames)
	addInt("FREE_AREA/FREE_AREA_AREA_TOTAL", a.FreeAreaTotalSqm)

	addFloat("PRICE", a.Price)
	add("PRICE_FOR_DISPLAY", a.PriceDisplay)
	addFloat("PRICE/SQUARE_METER", a.PricePerSqm)
	add("PRICE/SQUARE_METER_FOR_DISPLAY", a.PricePerSqmDisplay)
	add("PRICE/SQUARE_METER_FOR_DISPLAY_WITH_UNIT", a.PricePerSqmDisplayWithUnit)
	addFloat("RENT/PER_MONTH_LETTINGS", a.RentPerMonth)

	add("MMO", a.MainImage)
	if len(a.ImageURLs) > 0 {
		add("ALL_IMAGE_URLS", strings.Join(a.ImageURLs, ";"))
	}
	add("imagedescription", a.ImageDescription)
	add("AD_SEARCHRESULT_LOGO", a.AdSearchResultLogo)
	add("VIRTUAL_VIEW_LINK", a.VirtualViewLink)

	addBool("ISPRIVATE", a.IsPrivate)
	addBool("IS_BUMPED", a.IsBumped)
	addBool("UPSELLING_AD_SEARCHRESULT", a.UpsellingAdSearchResult)
	if !a.Published.IsZero() {
		add("PUBLISHED_String", a.Published.UTC().Format(time.RFC3339))
		add("PUBLISHED", strconv.FormatInt(a.Published.UnixMilli(), 10))
	}
	add("SEO_URL", a.SeoURL)
	if len(a.EstatePreference) > 0 {
		add("ESTATE_PREFERENCE", strings.Join(a.EstatePreference, ", "))
	}

	extraNames := make([]string, 0, len(a.Extra))
	for name := range a.Extra {
		extraNames = append(extraNames, name)
	}
	sort.Strings(extraNames)
	for _, name := range extraNames {
		addValues(name, a.Extra[name])
	}

	return attrs
}

// Merge combines a with other into a new Attributes, attribute name by
// attribute name: other's value wins wherever both bags name the same
// attribute, but names only one of them has are kept. This is for combining
// a search result's attribute bag with a listing detail page's - the two
// overlap heavily but aren't identical, both in which attributes they carry
// (search-only: HEADING, SEO_URL, ALL_IMAGE_URLS, COORDINATES, flattened
// LOCATION/ADDRESS/DISTRICT/STATE/COUNTRY/POSTCODE; detail-only: full
// DESCRIPTION body, energy/construction/contact attributes, ...) and in the
// names used for the same fact (e.g. search's NUMBER_OF_ROOMS vs a detail
// page's NO_OF_ROOMS - both end up preserved, just under separate names).
func (a Attributes) Merge(other Attributes) Attributes {
	byName := make(map[string][]string, len(a.Extra)+len(other.Extra))
	var order []string

	take := func(wire []wireAttribute) {
		for _, attr := range wire {
			if _, ok := byName[attr.Name]; !ok {
				order = append(order, attr.Name)
			}
			byName[attr.Name] = attr.Values
		}
	}
	take(a.wire())
	take(other.wire())

	var merged Attributes
	for _, name := range order {
		values := byName[name]
		if merged.set(name, values) {
			continue
		}
		if merged.Extra == nil {
			merged.Extra = make(map[string][]string)
		}
		merged.Extra[name] = values
	}

	return merged
}

func parseInt(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil
}

func parseFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

func parseBool(s string) (bool, bool) {
	b, err := strconv.ParseBool(s)
	return b, err == nil
}

func parseCoordinates(s string) (Coordinates, bool) {
	lat, lon, found := strings.Cut(s, ",")
	if !found {
		return Coordinates{}, false
	}

	latitude, ok := parseFloat(strings.TrimSpace(lat))
	if !ok {
		return Coordinates{}, false
	}
	longitude, ok := parseFloat(strings.TrimSpace(lon))
	if !ok {
		return Coordinates{}, false
	}

	return Coordinates{Latitude: latitude, Longitude: longitude}, true
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	return out
}
