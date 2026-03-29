package service

import (
	"strconv"
	"testing"

	"github.com/dvnpv/subscriptions-aggregator/pkg/month"
	"github.com/google/uuid"
)

func TestValidateAndBuildSubscription_SuccessWithoutEndDate(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()

	sub, err := validateAndBuildSubscription(
		id,
		"Netflix",
		999,
		userID,
		"01-2026",
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedStart, err := month.ParseMonthYear("01-2026")
	if err != nil {
		t.Fatalf("failed to parse expected start date: %v", err)
	}

	if sub.ID != id {
		t.Fatalf("expected id %v, got %v", id, sub.ID)
	}
	if sub.ServiceName != "Netflix" {
		t.Fatalf("expected service name Netflix, got %q", sub.ServiceName)
	}
	if sub.Price != 999 {
		t.Fatalf("expected price 999, got %d", sub.Price)
	}
	if sub.UserID.String() != userID {
		t.Fatalf("expected user id %s, got %s", userID, sub.UserID.String())
	}
	if !sub.StartDate.Equal(expectedStart) {
		t.Fatalf("expected start date %v, got %v", expectedStart, sub.StartDate)
	}
	if sub.EndDate != nil {
		t.Fatalf("expected end date to be nil, got %v", sub.EndDate)
	}
}

func TestValidateAndBuildSubscription_SuccessWithEndDate(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()
	endDateStr := "03-2026"

	sub, err := validateAndBuildSubscription(
		id,
		"YouTube Premium",
		499,
		userID,
		"01-2026",
		&endDateStr,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedStart, _ := month.ParseMonthYear("01-2026")
	expectedEnd, _ := month.ParseMonthYear("03-2026")

	if !sub.StartDate.Equal(expectedStart) {
		t.Fatalf("expected start date %v, got %v", expectedStart, sub.StartDate)
	}
	if sub.EndDate == nil {
		t.Fatal("expected end date to be set, got nil")
	}
	if !sub.EndDate.Equal(expectedEnd) {
		t.Fatalf("expected end date %v, got %v", expectedEnd, *sub.EndDate)
	}
}

func TestValidateAndBuildSubscription_TrimsServiceName(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()

	sub, err := validateAndBuildSubscription(
		id,
		"   Netflix   ",
		999,
		userID,
		"01-2026",
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sub.ServiceName != "Netflix" {
		t.Fatalf("expected trimmed service name Netflix, got %q", sub.ServiceName)
	}
}

func TestValidateAndBuildSubscription_EmptyServiceName(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()

	_, err := validateAndBuildSubscription(
		id,
		"   ",
		999,
		userID,
		"01-2026",
		nil,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "service_name is required" {
		t.Fatalf("expected error %q, got %q", "service_name is required", err.Error())
	}
}

func TestValidateAndBuildSubscription_InvalidPrice(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()

	testCases := []int{0, -1, -100}

	for _, price := range testCases {
		t.Run("price_"+strconv.Itoa(price), func(t *testing.T) {
			_, err := validateAndBuildSubscription(
				id,
				"Netflix",
				price,
				userID,
				"01-2026",
				nil,
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != "price must be greater than 0" {
				t.Fatalf("expected error %q, got %q", "price must be greater than 0", err.Error())
			}
		})
	}
}

func TestValidateAndBuildSubscription_InvalidUserID(t *testing.T) {
	id := uuid.New()

	_, err := validateAndBuildSubscription(
		id,
		"Netflix",
		999,
		"not-a-uuid",
		"01-2026",
		nil,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "invalid user_id" {
		t.Fatalf("expected error %q, got %q", "invalid user_id", err.Error())
	}
}

func TestValidateAndBuildSubscription_InvalidStartDate(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()

	_, err := validateAndBuildSubscription(
		id,
		"Netflix",
		999,
		userID,
		"2026-01",
		nil,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "invalid start_date" {
		t.Fatalf("expected error %q, got %q", "invalid start_date", err.Error())
	}
}

func TestValidateAndBuildSubscription_InvalidEndDate(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()
	endDateStr := "2026-03"

	_, err := validateAndBuildSubscription(
		id,
		"Netflix",
		999,
		userID,
		"01-2026",
		&endDateStr,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "invalid end_date" {
		t.Fatalf("expected error %q, got %q", "invalid end_date", err.Error())
	}
}

func TestValidateAndBuildSubscription_EndDateBeforeStartDate(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()
	endDateStr := "01-2025"

	_, err := validateAndBuildSubscription(
		id,
		"Netflix",
		999,
		userID,
		"02-2025",
		&endDateStr,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "end_date must be greater than or equal to start_date" {
		t.Fatalf(
			"expected error %q, got %q",
			"end_date must be greater than or equal to start_date",
			err.Error(),
		)
	}
}

func TestValidateAndBuildSubscription_EmptyEndDateString_TreatedAsNil(t *testing.T) {
	id := uuid.New()
	userID := uuid.New().String()
	endDateStr := "   "

	sub, err := validateAndBuildSubscription(
		id,
		"Netflix",
		999,
		userID,
		"01-2026",
		&endDateStr,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sub.EndDate != nil {
		t.Fatalf("expected nil end date, got %v", sub.EndDate)
	}
}
