package willhaben

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/erendogan51/immo-scraper/pkg/db/sql/db"
	"github.com/erendogan51/immo-scraper/pkg/models"
	"github.com/google/uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// toUpsertParams builds the row to persist for ad, as found via a search
// under searchPath. detail is the same advert's listing detail page, if it
// was fetched successfully (via GetListing); pass nil if it wasn't - the
// row is then populated from the search result alone, leaving the
// detail-only columns (AdUuid, StartDate, ..., Images) unset.
func toUpsertParams(ad models.AdvertSummary, detail *ListingResult, searchPath string) (db.UpsertListingParams, error) {
	attrs := ad.Attributes
	if detail != nil {
		attrs = attrs.Merge(detail.Attributes)
	}

	attrsJSON, err := json.Marshal(&attrs)
	if err != nil {
		return db.UpsertListingParams{}, fmt.Errorf("marshal attributes: %w", err)
	}

	params := db.UpsertListingParams{
		AdID:           ad.ID,
		VerticalID:     int32(ad.VerticalID),
		AdTypeID:       int32(ad.AdTypeID),
		ProductID:      int32(ad.ProductID),
		AdvertStatusID: ad.AdvertStatus.ID,
		AdvertStatus:   ad.AdvertStatus.Description,
		Description:    ad.Description,
		Heading:        text(attrs.Heading),
		PropertyType:   text(attrs.PropertyType),
		RoomsLayout:    text(attrs.RoomsLayout),
		NumberOfRooms:  decimalPtr(attrs.NumberOfRooms),
		Floor:          int4(attrs.Floor),
		Address:        text(attrs.Address),
		Location:       text(attrs.Location),
		District:       text(attrs.District),
		State:          text(attrs.State),
		Country:        text(attrs.Country),
		Postcode:       text(attrs.PostCode),
		Latitude:       float8(attrs.Coordinates.Latitude),
		Longitude:      float8(attrs.Coordinates.Longitude),
		LivingAreaSqm:  int4(attrs.LivingAreaSqm),
		UsableAreaSqm:  int4(attrs.UsableAreaSqm),
		EstateSizeSqm:  int4(attrs.EstateSizeSqm),
		Price:          decimalPtr(attrs.Price),
		PriceDisplay:   text(attrs.PriceDisplay),
		PricePerSqm:    decimalPtr(attrs.PricePerSqm),
		RentPerMonth:   decimalPtr(attrs.RentPerMonth),
		OrgID:          text(attrs.OrgID),
		OrgName:        text(attrs.OrgName),
		IsPrivate:      attrs.IsPrivate,
		MainImage:      text(attrs.MainImage),
		SeoUrl:         attrs.SeoURL,
		PublishedAt:    timestamptz(attrs.Published),
		Attributes:     attrsJSON,
		SearchPath:     text(searchPath),
		Images:         []byte("[]"),
	}

	if detail != nil {
		params.AdUuid = nullUUID(detail.UUID)
		params.StartDate = parseWireDate(detail.StartDate)
		params.EndDate = parseWireDate(detail.EndDate)
		params.FirstPublishedAt = parseWireDate(detail.FirstPublishedDate)
		params.CreatedAt = parseWireDate(detail.CreatedDate)
		params.ChangedAt = parseWireDate(detail.ChangedDate)
		params.Municipality = text(detail.AdvertAddressDetails.Municipality)
		params.CanonicalUrl = text(detail.SeoMetaData.CanonicalURL)

		if images := detail.AdvertImageList.AdvertImage; len(images) > 0 {
			if imagesJSON, err := json.Marshal(images); err == nil {
				params.Images = imagesJSON
			}
		}
	}

	return params, nil
}

func text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func int4(n int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(n), Valid: n != 0}
}

func float8(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: f != 0}
}

// decimalPtr returns nil for f == 0, matching Attributes' convention of
// leaving numeric fields at their zero value when willhaben didn't send the
// corresponding attribute.
func decimalPtr(f float64) *pgxdecimal.Decimal {
	if f == 0 {
		return nil
	}

	d := pgxdecimal.Decimal(decimal.NewFromFloat(f))

	return &d
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{Time: t, Valid: true}
}

// parseWireDate parses a listing detail page's date fields (startDate,
// endDate, publishedDate, ...), which use the same non-RFC3339 layout as a
// search result's PUBLISHED_String attribute.
func parseWireDate(s string) pgtype.Timestamptz {
	if s == "" {
		return pgtype.Timestamptz{}
	}

	t, err := time.Parse(searchDateLayout, s)
	if err != nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{Time: t, Valid: true}
}

func nullUUID(s string) uuid.NullUUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: id, Valid: true}
}
