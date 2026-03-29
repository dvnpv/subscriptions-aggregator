package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dvnpv/subscriptions-aggregator/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *model.Subscription) error {
	query := `
        INSERT INTO subscriptions (
            id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    `
	_, err := r.db.Exec(ctx, query, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	return err
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
    `

	var s model.Subscription
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.ServiceName,
		&s.Price,
		&s.UserID,
		&s.StartDate,
		&s.EndDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &s, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *model.Subscription) error {
	query := `
        UPDATE subscriptions
        SET service_name = $2,
            price = $3,
            user_id = $4,
            start_date = $5,
            end_date = $6,
            updated_at = NOW()
        WHERE id = $1
    `
	tag, err := r.db.Exec(ctx, query, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE ($1::uuid IS NULL OR user_id = $1)
          AND ($2::text IS NULL OR service_name = $2)
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.Query(ctx, query, userID, serviceName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, s)
	}

	return items, rows.Err()
}

func (r *SubscriptionRepository) ListForTotal(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	periodFrom time.Time,
	periodTo time.Time,
) ([]model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE ($1::uuid IS NULL OR user_id = $1)
          AND ($2::text IS NULL OR service_name = $2)
          AND start_date <= $4
          AND (end_date IS NULL OR end_date >= $3)
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(ctx, query, userID, serviceName, periodFrom, periodTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, s)
	}

	return items, rows.Err()
}
