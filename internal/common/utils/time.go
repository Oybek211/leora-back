package utils

import "time"

// NowUTC returns current time in RFC3339.
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ParseRFC3339 safely parses RFC3339 timestamps.
func ParseRFC3339(value string) time.Time {
	t, _ := time.Parse(time.RFC3339, value)
	return t
}
