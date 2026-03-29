package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dvnpv/subscriptions-aggregator/internal/dto"
	"github.com/dvnpv/subscriptions-aggregator/internal/httpresponse"
	"github.com/dvnpv/subscriptions-aggregator/internal/model"
	"github.com/dvnpv/subscriptions-aggregator/internal/repository"
	"github.com/dvnpv/subscriptions-aggregator/internal/service"
	"github.com/dvnpv/subscriptions-aggregator/pkg/month"

	"github.com/go-chi/chi/v5"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *slog.Logger
}

func NewSubscriptionHandler(service *service.SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// Create godoc
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body dto.CreateSubscriptionRequest true "Subscription payload"
// @Success 201 {object} dto.SubscriptionResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /subscriptions/ [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "failed to decode request body")
		return
	}

	sub, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create subscription", slog.String("error", err.Error()))
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	httpresponse.JSON(w, http.StatusCreated, toSubscriptionResponse(sub))
}

// GetByID godoc
// @Summary Get subscription by id
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 404 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httpresponse.Error(w, http.StatusNotFound, "not found", "subscription not found")
			return
		}
		if err.Error() == "invalid subscription id" {
			httpresponse.Error(w, http.StatusBadRequest, "invalid request", err.Error())
			return
		}

		h.logger.Error("failed to get subscription", slog.String("error", err.Error()))
		httpresponse.Error(w, http.StatusInternalServerError, "internal error", "failed to get subscription")
		return
	}

	httpresponse.JSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

// Update godoc
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body dto.UpdateSubscriptionRequest true "Subscription payload"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 404 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "failed to decode request body")
		return
	}

	sub, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httpresponse.Error(w, http.StatusNotFound, "not found", "subscription not found")
			return
		}
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	httpresponse.JSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

// Delete godoc
// @Summary Delete subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 404 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httpresponse.Error(w, http.StatusNotFound, "not found", "subscription not found")
			return
		}
		if err.Error() == "invalid subscription id" {
			httpresponse.Error(w, http.StatusBadRequest, "invalid request", err.Error())
			return
		}

		h.logger.Error("failed to delete subscription", slog.String("error", err.Error()))
		httpresponse.Error(w, http.StatusInternalServerError, "internal error", "failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.ListSubscriptionsResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /subscriptions/ [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	limit := 10
	offset := 0

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "invalid request", "limit must be integer")
			return
		}
		limit = parsed
	}

	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpresponse.Error(w, http.StatusBadRequest, "invalid request", "offset must be integer")
			return
		}
		offset = parsed
	}

	items, err := h.service.List(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	respItems := make([]dto.SubscriptionResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, toSubscriptionResponse(&item))
	}

	httpresponse.JSON(w, http.StatusOK, dto.ListSubscriptionsResponse{
		Items: respItems,
		Count: len(respItems),
	})
}

// GetTotal godoc
// @Summary Calculate total subscription cost for a selected period
// @Description Calculates total monthly subscription cost by counting month intersections with the requested period
// @Tags subscriptions
// @Produce json
// @Param from query string true "Start of calculation period in MM-YYYY format"
// @Param to query string true "End of calculation period in MM-YYYY format"
// @Param user_id query string false "User ID"
// @Param service_name query string false "Service name"
// @Success 200 {object} dto.TotalResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) GetTotal(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	if from == "" || to == "" {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "from and to are required")
		return
	}

	resp, err := h.service.CalculateTotal(r.Context(), from, to, userID, serviceName)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	httpresponse.JSON(w, http.StatusOK, resp)
}

func toSubscriptionResponse(sub *model.Subscription) dto.SubscriptionResponse {
	var endDate *string
	if sub.EndDate != nil {
		v := month.FormatMonthYear(*sub.EndDate)
		endDate = &v
	}

	return dto.SubscriptionResponse{
		ID:          sub.ID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   month.FormatMonthYear(sub.StartDate),
		EndDate:     endDate,
		CreatedAt:   sub.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   sub.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
