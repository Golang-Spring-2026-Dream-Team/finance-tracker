package repository

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func numericToString4(n pgtype.Numeric) string {
	if !n.Valid || n.NaN || n.Int == nil {
		return "0.0000"
	}

	r := new(big.Rat).SetInt(n.Int)
	switch {
	case n.Exp > 0:
		r = r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	case n.Exp < 0:
		r = r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	return r.FloatString(4)
}

func stringToNumeric(input string) (pgtype.Numeric, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return pgtype.Numeric{}, fmt.Errorf("amount is required")
	}
	var n pgtype.Numeric
	if err := n.Scan(trimmed); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid numeric value")
	}
	return n, nil
}

func textFromPtr(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func int8FromPtr(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func dateFromString(iso string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date, expected YYYY-MM-DD")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func timestamptzToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func dateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func optionalNumericFromPtr(v *string) (pgtype.Numeric, error) {
	if v == nil {
		return pgtype.Numeric{}, nil
	}
	return stringToNumeric(*v)
}

func optionalDateFromPtr(v *string) (pgtype.Date, error) {
	if v == nil {
		return pgtype.Date{}, nil
	}
	return dateFromString(*v)
}

// decimalGreaterOrEqual reports whether decimal string a >= b.
// If either value cannot be parsed it returns true (does not block the operation).
func decimalGreaterOrEqual(a, b string) bool {
	ra, ok1 := new(big.Rat).SetString(a)
	rb, ok2 := new(big.Rat).SetString(b)
	if !ok1 || !ok2 {
		return true
	}
	return ra.Cmp(rb) >= 0
}
