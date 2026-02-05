package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// dateFromTime creates a pgtype.Date from a time.Time.
func dateFromTime(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

// uuidToPtr converts a pgtype.UUID to a *string.
func uuidToPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := formatUUID(u.Bytes)
	return &s
}

// formatUUID formats UUID bytes as a string.
func formatUUID(b [16]byte) string {
	return string([]byte{
		hexDigit(b[0] >> 4), hexDigit(b[0] & 0x0f),
		hexDigit(b[1] >> 4), hexDigit(b[1] & 0x0f),
		hexDigit(b[2] >> 4), hexDigit(b[2] & 0x0f),
		hexDigit(b[3] >> 4), hexDigit(b[3] & 0x0f),
		'-',
		hexDigit(b[4] >> 4), hexDigit(b[4] & 0x0f),
		hexDigit(b[5] >> 4), hexDigit(b[5] & 0x0f),
		'-',
		hexDigit(b[6] >> 4), hexDigit(b[6] & 0x0f),
		hexDigit(b[7] >> 4), hexDigit(b[7] & 0x0f),
		'-',
		hexDigit(b[8] >> 4), hexDigit(b[8] & 0x0f),
		hexDigit(b[9] >> 4), hexDigit(b[9] & 0x0f),
		'-',
		hexDigit(b[10] >> 4), hexDigit(b[10] & 0x0f),
		hexDigit(b[11] >> 4), hexDigit(b[11] & 0x0f),
		hexDigit(b[12] >> 4), hexDigit(b[12] & 0x0f),
		hexDigit(b[13] >> 4), hexDigit(b[13] & 0x0f),
		hexDigit(b[14] >> 4), hexDigit(b[14] & 0x0f),
		hexDigit(b[15] >> 4), hexDigit(b[15] & 0x0f),
	})
}

func hexDigit(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + b - 10
}

// int4FromInt creates a pgtype.Int4 from an int.
func int4FromInt(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// int4ToInt converts a pgtype.Int4 to an int.
func int4ToInt(i pgtype.Int4) int {
	if !i.Valid {
		return 0
	}
	return int(i.Int32)
}

// numericToFloat64Ptr converts a pgtype.Numeric to a *float64.
func numericToFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, _ := n.Float64Value()
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// numericFromPtr creates a pgtype.Numeric from a *float64.
func numericFromPtr(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	_ = n.Scan(*f)
	return n
}

// timestamptzFromTime creates a pgtype.Timestamptz from a time.Time.
func timestamptzFromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// uuidFromString creates a pgtype.UUID from a string (non-nullable).
func uuidFromString(s string) pgtype.UUID {
	var uuid pgtype.UUID
	_ = uuid.Scan(s)
	return uuid
}
