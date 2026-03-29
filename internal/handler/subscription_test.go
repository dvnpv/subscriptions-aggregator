package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dvnpv/subscriptions-aggregator/internal/dto"
	"github.com/dvnpv/subscriptions-aggregator/internal/model"
	"github.com/dvnpv/subscriptions-aggregator/internal/repository"
	"github.com/dvnpv/subscriptions-aggregator/pkg/month"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockSubscriptionService struct {
	createFn         func(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error)
	getByIDFn        func(ctx context.Context, id string) (*model.Subscription, error)
	updateFn         func(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error)
	deleteFn         func(ctx context.Context, id string) error
	listFn           func(ctx context.Context, userIDStr, serviceName string, limit, offset int) ([]model.Subscription, error)
	calculateTotalFn func(ctx context.Context, fromStr, toStr, userIDStr, serviceName string) (*dto.TotalResponse, error)
}

func (m *mockSubscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
	return m.createFn(ctx, req)
}

func (m *mockSubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockSubscriptionService) Update(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
	return m.updateFn(ctx, id, req)
}

func (m *mockSubscriptionService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func (m *mockSubscriptionService) List(ctx context.Context, userIDStr, serviceName string, limit, offset int) ([]model.Subscription, error) {
	return m.listFn(ctx, userIDStr, serviceName, limit, offset)
}

func (m *mockSubscriptionService) CalculateTotal(ctx context.Context, fromStr, toStr, userIDStr, serviceName string) (*dto.TotalResponse, error) {
	return m.calculateTotalFn(ctx, fromStr, toStr, userIDStr, serviceName)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func testSubscription(t *testing.T) *model.Subscription {
	t.Helper()

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	startDate, err := month.ParseMonthYear("01-2026")
	if err != nil {
		t.Fatalf("failed to parse start date: %v", err)
	}
	endDate, err := month.ParseMonthYear("03-2026")
	if err != nil {
		t.Fatalf("failed to parse end date: %v", err)
	}

	return &model.Subscription{
		ID:          id,
		ServiceName: "Netflix",
		Price:       999,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     &endDate,
		CreatedAt:   time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 11, 13, 0, 0, 0, time.UTC),
	}
}

func TestSubscriptionHandler_Create_Success(t *testing.T) {
	sub := testSubscription(t)

	svc := &mockSubscriptionService{
		createFn: func(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
			if req.ServiceName != "Netflix" {
				t.Fatalf("expected service name Netflix, got %q", req.ServiceName)
			}
			if req.Price != 999 {
				t.Fatalf("expected price 999, got %d", req.Price)
			}
			return sub, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	body := `{
        "service_name":"Netflix",
        "price":999,
        "user_id":"22222222-2222-2222-2222-222222222222",
        "start_date":"01-2026",
        "end_date":"03-2026"
    }`

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	var resp dto.SubscriptionResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != sub.ID.String() {
		t.Fatalf("expected id %s, got %s", sub.ID.String(), resp.ID)
	}
	if resp.ServiceName != "Netflix" {
		t.Fatalf("expected service name Netflix, got %q", resp.ServiceName)
	}
}

func TestSubscriptionHandler_Create_InvalidJSON(t *testing.T) {
	svc := &mockSubscriptionService{
		createFn: func(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
			t.Fatal("service.Create should not be called on invalid JSON")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_Success(t *testing.T) {
	sub := testSubscription(t)

	svc := &mockSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			if id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected id %q", id)
			}
			return sub, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_NotFound(t *testing.T) {
	svc := &mockSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			return nil, repository.ErrNotFound
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/missing", nil)
	req = withURLParam(req, "id", "missing")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_InvalidID(t *testing.T) {
	svc := &mockSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			return nil, errors.New("invalid subscription id")
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/bad-id", nil)
	req = withURLParam(req, "id", "bad-id")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Delete_Success(t *testing.T) {
	svc := &mockSubscriptionService{
		deleteFn: func(ctx context.Context, id string) error {
			if id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected id %q", id)
			}
			return nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestSubscriptionHandler_List_Success(t *testing.T) {
	sub := *testSubscription(t)

	svc := &mockSubscriptionService{
		listFn: func(ctx context.Context, userIDStr, serviceName string, limit, offset int) ([]model.Subscription, error) {
			if userIDStr != "22222222-2222-2222-2222-222222222222" {
				t.Fatalf("unexpected userID %q", userIDStr)
			}
			if serviceName != "Netflix" {
				t.Fatalf("unexpected serviceName %q", serviceName)
			}
			if limit != 5 {
				t.Fatalf("expected limit 5, got %d", limit)
			}
			if offset != 10 {
				t.Fatalf("expected offset 10, got %d", offset)
			}
			return []model.Subscription{sub}, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/?user_id=22222222-2222-2222-2222-222222222222&service_name=Netflix&limit=5&offset=10",
		nil,
	)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp dto.ListSubscriptionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count != 1 {
		t.Fatalf("expected count 1, got %d", resp.Count)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Items[0].ServiceName != "Netflix" {
		t.Fatalf("expected Netflix, got %q", resp.Items[0].ServiceName)
	}
}

func TestSubscriptionHandler_List_InvalidLimit(t *testing.T) {
	svc := &mockSubscriptionService{
		listFn: func(ctx context.Context, userIDStr, serviceName string, limit, offset int) ([]model.Subscription, error) {
			t.Fatal("service.List should not be called on invalid limit")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_GetTotal_Success(t *testing.T) {
	svc := &mockSubscriptionService{
		calculateTotalFn: func(ctx context.Context, fromStr, toStr, userIDStr, serviceName string) (*dto.TotalResponse, error) {
			if fromStr != "01-2026" || toStr != "03-2026" {
				t.Fatalf("unexpected range: from=%q to=%q", fromStr, toStr)
			}
			return &dto.TotalResponse{
				Total:       2997,
				From:        "01-2026",
				To:          "03-2026",
				UserID:      userIDStr,
				ServiceName: serviceName,
			}, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=01-2026&to=03-2026", nil)
	rec := httptest.NewRecorder()

	h.GetTotal(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp dto.TotalResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 2997 {
		t.Fatalf("expected total 2997, got %d", resp.Total)
	}
}

func TestSubscriptionHandler_GetTotal_MissingRequiredParams(t *testing.T) {
	svc := &mockSubscriptionService{
		calculateTotalFn: func(ctx context.Context, fromStr, toStr, userIDStr, serviceName string) (*dto.TotalResponse, error) {
			t.Fatal("service.CalculateTotal should not be called when from/to are missing")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=01-2026", nil)
	rec := httptest.NewRecorder()

	h.GetTotal(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestToSubscriptionResponse(t *testing.T) {
	sub := testSubscription(t)

	resp := toSubscriptionResponse(sub)

	if resp.ID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected id %q", resp.ID)
	}
	if resp.ServiceName != "Netflix" {
		t.Fatalf("unexpected service name %q", resp.ServiceName)
	}
	if resp.Price != 999 {
		t.Fatalf("unexpected price %d", resp.Price)
	}
	if resp.UserID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("unexpected user id %q", resp.UserID)
	}
	if resp.StartDate != "01-2026" {
		t.Fatalf("unexpected start date %q", resp.StartDate)
	}
	if resp.EndDate == nil || *resp.EndDate != "03-2026" {
		t.Fatalf("unexpected end date %v", resp.EndDate)
	}
	if resp.CreatedAt != "2026-01-10T12:00:00Z" {
		t.Fatalf("unexpected created_at %q", resp.CreatedAt)
	}
	if resp.UpdatedAt != "2026-01-11T13:00:00Z" {
		t.Fatalf("unexpected updated_at %q", resp.UpdatedAt)
	}
}

func TestSubscriptionHandler_Update_Success(t *testing.T) {
	sub := testSubscription(t)

	svc := &mockSubscriptionService{
		updateFn: func(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
			if id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected id %q", id)
			}
			if req.ServiceName != "Netflix Premium" {
				t.Fatalf("expected service name Netflix Premium, got %q", req.ServiceName)
			}
			if req.Price != 1299 {
				t.Fatalf("expected price 1299, got %d", req.Price)
			}
			return sub, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	body := `{
        "service_name":"Netflix Premium",
        "price":1299,
        "user_id":"22222222-2222-2222-2222-222222222222",
        "start_date":"01-2026",
        "end_date":"04-2026"
    }`

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp dto.SubscriptionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != sub.ID.String() {
		t.Fatalf("expected id %s, got %s", sub.ID.String(), resp.ID)
	}
}

func TestSubscriptionHandler_Update_InvalidJSON(t *testing.T) {
	svc := &mockSubscriptionService{
		updateFn: func(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
			t.Fatal("service.Update should not be called on invalid JSON")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Update_NotFound(t *testing.T) {
	svc := &mockSubscriptionService{
		updateFn: func(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
			return nil, repository.ErrNotFound
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	body := `{
        "service_name":"Netflix Premium",
        "price":1299,
        "user_id":"22222222-2222-2222-2222-222222222222",
        "start_date":"01-2026",
        "end_date":"04-2026"
    }`

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/missing", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", "missing")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_Update_InvalidRequest(t *testing.T) {
	svc := &mockSubscriptionService{
		updateFn: func(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
			return nil, errors.New("price must be greater than 0")
		},
	}

	h := NewSubscriptionHandler(svc, newTestLogger())

	body := `{
        "service_name":"Netflix Premium",
        "price":0,
        "user_id":"22222222-2222-2222-2222-222222222222",
        "start_date":"01-2026",
        "end_date":"04-2026"
    }`

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")

	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
