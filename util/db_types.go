package util

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func createPostgresType[T, R any](input any, rspFn func(a T) R) R {
	switch typeValue := input.(type) {
	case *T:
		if typeValue == nil {
			var r R
			return r
		}
		return rspFn(*typeValue)
	case T:
		return rspFn(typeValue)
	default:
		var r T
		return rspFn(r)
	}
}

func ToPGTimestamp[T time.Time | *time.Time](t T) pgtype.Timestamptz {
	return createPostgresType[time.Time, pgtype.Timestamptz](t, func(a time.Time) pgtype.Timestamptz {
		if a.IsZero() || a.Unix() == 0 {
			return pgtype.Timestamptz{}
		}

		return pgtype.Timestamptz{
			Time:  a,
			Valid: true,
		}
	})
}

func ToPGText[T string | *string](str T) pgtype.Text {
	return createPostgresType[string, pgtype.Text](str, func(a string) pgtype.Text {
		if a == "" {
			return pgtype.Text{}
		}

		return pgtype.Text{
			String: a,
			Valid:  true,
		}
	})
}

func ToPGUUIDNull[T uuid.UUID | *uuid.UUID](u T) uuid.NullUUID {
	return createPostgresType[uuid.UUID, uuid.NullUUID](u, func(a uuid.UUID) uuid.NullUUID {
		return uuid.NullUUID{
			UUID:  a,
			Valid: true,
		}
	})
}

func ToPGInt4[T ~int32 | *int32](i T) pgtype.Int4 {
	return createPostgresType[int32, pgtype.Int4](i, func(a int32) pgtype.Int4 {
		return pgtype.Int4{
			Int32: a,
			Valid: true,
		}
	})
}

func ToPGDecimal[T float32 | float64 | *float32 | *float64](f T) *pgxdecimal.Decimal {
	switch typeValue := any(f).(type) {
	case *float32:
		if typeValue == nil {
			p := pgxdecimal.Decimal(decimal.Zero)
			return &p
		}

		fString, err := decimal.NewFromString(fmt.Sprintf("%f", *typeValue))
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to convert float32 to decimal via string format, falling back to float64 conversion")
			p := pgxdecimal.Decimal(decimal.NewFromFloat(float64(*typeValue)))
			return &p
		}

		p := pgxdecimal.Decimal(fString)
		return &p
	case float32:
		fString, err := decimal.NewFromString(fmt.Sprintf("%f", typeValue))
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to convert float32 to decimal via string format, falling back to float64 conversion")
			p := pgxdecimal.Decimal(decimal.NewFromFloat(float64(typeValue)))
			return &p
		}

		p := pgxdecimal.Decimal(fString)
		return &p
	case *float64:
		if typeValue == nil {
			p := pgxdecimal.Decimal(decimal.Zero)
			return &p
		}

		p := pgxdecimal.Decimal(decimal.NewFromFloat(*typeValue))
		return &p
	case float64:
		p := pgxdecimal.Decimal(decimal.NewFromFloat(typeValue))
		return &p
	default:
		p := pgxdecimal.Decimal(decimal.Zero)
		return &p
	}
}

func ToPGBool[T bool | *bool](b T) pgtype.Bool {
	return createPostgresType[bool, pgtype.Bool](b, func(a bool) pgtype.Bool {
		return pgtype.Bool{
			Bool:  a,
			Valid: true,
		}
	})
}

func FromPGDecimal(d *pgxdecimal.Decimal) *float64 {
	if d == nil {
		return nil
	}

	f, err := d.Float64Value()
	if err != nil {
		return nil
	}

	res := f.Float64
	return &res
}

func FromPGInt4(a pgtype.Int4) *int32 {
	if !a.Valid {
		return nil
	}

	return &a.Int32
}

func FromPGBool(a pgtype.Bool) *bool {
	if !a.Valid {
		return nil
	}

	return &a.Bool
}

func FromPGText(a pgtype.Text) *string {
	if !a.Valid {
		return nil
	}

	return &a.String
}

func FromPGTimestamp(a pgtype.Timestamptz) *time.Time {
	if !a.Valid {
		return nil
	}

	utc := a.Time.UTC()
	return &utc
}
