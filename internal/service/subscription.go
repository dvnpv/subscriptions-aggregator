package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dvnpv/subscriptions-aggregator/internal/dto"
	"github.com/dvnpv/subscriptions-aggregator/internal/model"
	"github.com/dvnpv/subscriptions-aggregator/internal/repository"
	"github.com/dvnpv/subscriptions-aggregator/pkg/month"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
	sub, err := validateAndBuildSubscription(
		uuid.New(),
		req.ServiceName,
		req.Price,
		req.UserID,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, sub.ID)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid subscription id")
	}
	return s.repo.GetByID(ctx, uid)
}

func (s *SubscriptionService) Update(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
	subID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid subscription id")
	}

	sub, err := validateAndBuildSubscription(
		subID,
		req.ServiceName,
		req.Price,
		req.UserID,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, sub.ID)
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid subscription id")
	}
	return s.repo.Delete(ctx, uid)
}

func (s *SubscriptionService) List(ctx context.Context, userIDStr, serviceName string, limit, offset int) ([]model.Subscription, error) {
	var userID *uuid.UUID
	if userIDStr != "" {
		parsed, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user_id")
		}
		userID = &parsed
	}

	var serviceNamePtr *string
	if strings.TrimSpace(serviceName) != "" {
		serviceName = strings.TrimSpace(serviceName)
		serviceNamePtr = &serviceName
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, userID, serviceNamePtr, limit, offset)
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, fromStr, toStr, userIDStr, serviceName string) (*dto.TotalResponse, error) {
	from, err := month.ParseMonthYear(fromStr)
	if err != nil {
		return nil, errors.New("invalid from date")
	}

	to, err := month.ParseMonthYear(toStr)
	if err != nil {
		return nil, errors.New("invalid to date")
	}

	if err := month.ValidateRange(from, to); err != nil {
		return nil, err
	}

	var userID *uuid.UUID
	if userIDStr != "" {
		parsed, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user_id")
		}
		userID = &parsed
	}

	var serviceNamePtr *string
	if strings.TrimSpace(serviceName) != "" {
		serviceName = strings.TrimSpace(serviceName)
		serviceNamePtr = &serviceName
	}

	items, err := s.repo.ListForTotal(ctx, userID, serviceNamePtr, from, to)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, item := range items {
		months := month.MonthsIntersectionCount(item.StartDate, item.EndDate, from, to)
		total += months * item.Price
	}

	return &dto.TotalResponse{
		Total:       total,
		From:        fromStr,
		To:          toStr,
		UserID:      userIDStr,
		ServiceName: serviceName,
	}, nil
}

func validateAndBuildSubscription(
	id uuid.UUID,
	serviceName string,
	price int,
	userIDStr string,
	startDateStr string,
	endDateStr *string,
) (*model.Subscription, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, errors.New("service_name is required")
	}

	if price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user_id")
	}

	startDate, err := month.ParseMonthYear(startDateStr)
	if err != nil {
		return nil, errors.New("invalid start_date")
	}

	var endDate *time.Time
	if endDateStr != nil && strings.TrimSpace(*endDateStr) != "" {
		parsed, err := month.ParseMonthYear(*endDateStr)
		if err != nil {
			return nil, errors.New("invalid end_date")
		}
		if parsed.Before(startDate) {
			return nil, errors.New("end_date must be greater than or equal to start_date")
		}
		endDate = &parsed
	}

	return &model.Subscription{
		ID:          id,
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}
