package month

import (
	"errors"
	"fmt"
	"time"
)

const Layout = "01-2006"

func ParseMonthYear(value string) (time.Time, error) {
	t, err := time.Parse(Layout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month format, expected MM-YYYY: %w", err)
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonthYear(t time.Time) string {
	return t.Format(Layout)
}

func ValidateRange(from, to time.Time) error {
	if from.After(to) {
		return errors.New("from must be less than or equal to to")
	}
	return nil
}

func MonthsIntersectionCount(subStart time.Time, subEnd *time.Time, periodStart time.Time, periodEnd time.Time) int {
	effectiveEnd := periodEnd
	if subEnd != nil && subEnd.Before(periodEnd) {
		effectiveEnd = *subEnd
	}

	start := maxMonth(subStart, periodStart)
	end := minMonth(effectiveEnd, periodEnd)

	if start.After(end) {
		return 0
	}

	return monthsInclusive(start, end)
}

func monthsInclusive(start, end time.Time) int {
	return (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
}

func minMonth(a, b time.Time) time.Time {
	if a.Before(b) {
		return normalize(a)
	}
	return normalize(b)
}

func maxMonth(a, b time.Time) time.Time {
	if a.After(b) {
		return normalize(a)
	}
	return normalize(b)
}

func normalize(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
