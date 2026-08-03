package util

import (
	"testing"
	"time"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestToPGBool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    bool
		expected pgtype.Bool
	}{
		{
			name:  "true",
			input: true,
			expected: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		},
		{
			name:  "false",
			input: false,
			expected: pgtype.Bool{
				Bool:  false,
				Valid: true,
			},
		},
		{
			name: "zero value",
			expected: pgtype.Bool{
				Bool:  false,
				Valid: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGBool(tc.input))
		})
	}

	pointerTestCases := []struct {
		name     string
		input    *bool
		expected pgtype.Bool
	}{
		{
			name:  "pointer value",
			input: Pointer(true),
			expected: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		},
		{
			name:  "nil value",
			input: nil,
			expected: pgtype.Bool{
				Bool:  false,
				Valid: false,
			},
		},
	}

	for _, tc := range pointerTestCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGBool(tc.input))
		})
	}

}

func TestToPGInt4(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    int32
		expected pgtype.Int4
	}{
		{
			name:  "zero value",
			input: 0,
			expected: pgtype.Int4{
				Int32: 0,
				Valid: true,
			},
		},
		{
			name:  "some value",
			input: 123123,
			expected: pgtype.Int4{
				Int32: 123123,
				Valid: true,
			},
		},
		{
			name: "empty value",
			expected: pgtype.Int4{
				Int32: 0,
				Valid: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGInt4(tc.input))
		})
	}

	pointerTestCases := []struct {
		name     string
		input    *int32
		expected pgtype.Int4
	}{
		{
			name:  "pointer value",
			input: Pointer(int32(123123)),
			expected: pgtype.Int4{
				Int32: 123123,
				Valid: true,
			},
		},
		{
			name:  "nil value",
			input: nil,
			expected: pgtype.Int4{
				Int32: 0,
				Valid: false,
			},
		},
	}

	for _, tc := range pointerTestCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGInt4(tc.input))
		})
	}

}

func TestToPGTimestamp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    time.Time
		expected pgtype.Timestamptz
	}{
		{
			name:  "zero value",
			input: time.Time{},
			expected: pgtype.Timestamptz{
				Time:  time.Time{},
				Valid: false,
			},
		},
		{
			name:  "some value",
			input: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: pgtype.Timestamptz{
				Time:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGTimestamp(tc.input))
		})
	}

	pointerTestCases := []struct {
		name     string
		input    *time.Time
		expected pgtype.Timestamptz
	}{
		{
			name:  "nil value",
			input: &time.Time{},
			expected: pgtype.Timestamptz{
				Time:  time.Time{},
				Valid: false,
			},
		},
		{
			name:  "some value",
			input: Pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected: pgtype.Timestamptz{
				Time:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Valid: true,
			},
		},
	}

	for _, tc := range pointerTestCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGTimestamp(tc.input))
		})
	}
}

func TestToPGText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected pgtype.Text
	}{
		{
			name:  "zero value",
			input: "",
			expected: pgtype.Text{
				String: "",
				Valid:  false,
			},
		},
		{
			name:  "some value",
			input: t.Name(),
			expected: pgtype.Text{
				String: t.Name(),
				Valid:  true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGText(tc.input))
		})
	}

	pointerTestCases := []struct {
		name     string
		input    *string
		expected pgtype.Text
	}{
		{
			name:  "nil value",
			input: nil,
			expected: pgtype.Text{
				String: "",
				Valid:  false,
			},
		},
		{
			name:  "some value",
			input: Pointer(t.Name()),
			expected: pgtype.Text{
				String: t.Name(),
				Valid:  true,
			},
		},
	}

	for _, tc := range pointerTestCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToPGText(tc.input))
		})
	}
}

func TestToPGDecimal(t *testing.T) {
	t.Parallel()

	testCasesF32 := []struct {
		name     string
		input    float32
		expected pgxdecimal.Decimal
	}{
		{
			name:     "zero value",
			input:    0,
			expected: pgxdecimal.Decimal{},
		},
		{
			name:     "some value",
			input:    1.23,
			expected: pgxdecimal.Decimal(decimal.NewFromFloat32(1.23)),
		},
	}

	for _, tc := range testCasesF32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, decimal.Decimal(tc.expected).Equal(decimal.Decimal(*ToPGDecimal(tc.input))))
		})
	}

	pointerTestCasesF32 := []struct {
		name     string
		input    *float32
		expected pgxdecimal.Decimal
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: pgxdecimal.Decimal(decimal.Zero),
		},
		{
			name:     "some value",
			input:    Pointer(float32(1.23)),
			expected: pgxdecimal.Decimal(decimal.NewFromFloat32(1.23)),
		},
	}

	for _, tc := range pointerTestCasesF32 {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, decimal.Decimal(tc.expected).Equal(decimal.Decimal(*ToPGDecimal(tc.input))))
		})
	}

	testCasesF64 := []struct {
		name     string
		input    float64
		expected pgxdecimal.Decimal
	}{
		{
			name:     "zero value",
			input:    0,
			expected: pgxdecimal.Decimal{},
		},
		{
			name:     "some value",
			input:    1.23,
			expected: pgxdecimal.Decimal(decimal.NewFromFloat32(1.23)),
		},
	}

	for _, tc := range testCasesF64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, decimal.Decimal(tc.expected).Equal(decimal.Decimal(*ToPGDecimal(tc.input))))
		})
	}

	pointerTestCases64 := []struct {
		name     string
		input    *float64
		expected pgxdecimal.Decimal
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: pgxdecimal.Decimal(decimal.Zero),
		},
		{
			name:     "some value",
			input:    Pointer(1.23),
			expected: pgxdecimal.Decimal(decimal.NewFromFloat32(1.23)),
		},
	}

	for _, tc := range pointerTestCases64 {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, decimal.Decimal(tc.expected).Equal(decimal.Decimal(*ToPGDecimal(tc.input))))
		})
	}
}
