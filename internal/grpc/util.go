package grpc

import "time"

// ptr returns a pointer to an int64 value, or nil if zero.
func ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// ptrStr returns a pointer to a string value, or nil if empty.
func ptrStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// parseDate parses a YYYY-MM-DD date string.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
