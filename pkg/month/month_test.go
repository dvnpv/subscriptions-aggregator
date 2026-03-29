package month

import (
	"testing"
	"time"
)

func mustParse(t *testing.T, value string) time.Time {
	t.Helper()
	res, err := ParseMonthYear(value)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", value, err)
	}
	return res
}

func TestParseMonthYear(t *testing.T) {
	got, err := ParseMonthYear("07-2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Year() != 2025 || got.Month() != time.July || got.Day() != 1 {
		t.Fatalf("unexpected parsed date: %v", got)
	}
}

func TestMonthsIntersectionCount(t *testing.T) {
	subStart := mustParse(t, "07-2025")
	subEnd := mustParse(t, "10-2025")
	periodStart := mustParse(t, "08-2025")
	periodEnd := mustParse(t, "09-2025")

	got := MonthsIntersectionCount(subStart, &subEnd, periodStart, periodEnd)
	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestMonthsIntersectionCountOpenEnded(t *testing.T) {
	subStart := mustParse(t, "07-2025")
	periodStart := mustParse(t, "08-2025")
	periodEnd := mustParse(t, "10-2025")

	got := MonthsIntersectionCount(subStart, nil, periodStart, periodEnd)
	if got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}
