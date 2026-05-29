package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/klyakssa/aggregation-sub/internal/domain/subs"
	"go.uber.org/zap"
)

func (r *PostgresStorage) CreateSubscription(ctx context.Context, sub subs.Subscription) (subs.Subscription, error) {
	query := `
        INSERT INTO subs (
            service_name, 
            price, 
            user_id, 
            start_date, 
            end_date,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, updated_at
    `

	now := time.Now()
	var created subs.Subscription

	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	err := r.QueryRowxContext(
		ctxTimeout,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		now,
	).StructScan(&created)

	if err != nil {
		r.l.Error("error creating subscription", zap.Error(err))
		return subs.Subscription{}, subs.ErrFailedToCreateSubscription
	}

	created.ServiceName = sub.ServiceName
	created.Price = sub.Price
	created.UserID = sub.UserID
	created.StartDate = sub.StartDate
	created.EndDate = sub.EndDate

	return created, nil
}

func (r *PostgresStorage) GetSubscription(ctx context.Context, id int) (subs.Subscription, error) {
	query := `
        SELECT 
            id, 
            service_name, 
            price, 
            user_id, 
            start_date, 
            end_date,
            updated_at
        FROM subs
        WHERE id = $1
    `

	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	var sub subs.Subscription
	err := r.GetContext(ctxTimeout, &sub, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return subs.Subscription{}, subs.ErrFailedToFoundSubscription
		}
		r.l.Error("error getting subscription", zap.Error(err))
		return subs.Subscription{}, subs.ErrFailedToGetSubscription
	}

	return sub, nil
}

func (r *PostgresStorage) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]subs.Subscription, error) {
	query := `
        SELECT 
            id, 
            service_name, 
            price, 
            user_id, 
            start_date, 
            end_date,
            updated_at
        FROM subs
        WHERE user_id = $1
        ORDER BY id DESC
    `

	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	var subList []subs.Subscription
	err := r.SelectContext(ctxTimeout, &subList, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, subs.ErrFailedToFoundSubscription
		}
		r.l.Error("error getting subscriptions", zap.Error(err))
		return nil, subs.ErrFailedToGetSubscription
	}

	return subList, nil
}

func (r *PostgresStorage) UpdateSubscription(ctx context.Context, sub subs.Subscription) (subs.Subscription, error) {

	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	query := `
        UPDATE subs
        SET 
            service_name = COALESCE(NULLIF($1, ''), service_name),
            price = COALESCE(NULLIF($2, 0), price),
            start_date = COALESCE($3, start_date),
            end_date = $4,
            updated_at = $5
        WHERE id = $6
        RETURNING 
            id, 
            service_name, 
            price, 
            user_id, 
            start_date, 
            end_date,
            updated_at
    `

	now := time.Now()
	var updated subs.Subscription

	serviceName := sub.ServiceName
	price := sub.Price
	startDate := sub.StartDate
	endDate := sub.EndDate

	if serviceName == "" {
		serviceName = ""
	}
	if price == 0 {
		price = 0
	}
	if startDate.IsZero() {
		startDate = time.Time{}
	}

	err := r.QueryRowxContext(
		ctxTimeout,
		query,
		serviceName,
		price,
		startDate,
		endDate,
		now,
		sub.ID,
	).StructScan(&updated)

	if err != nil {
		r.l.Error("error updating subscription", zap.Error(err))
		return subs.Subscription{}, subs.ErrFailedToUpdateSubscription
	}

	return updated, nil
}

func (r *PostgresStorage) DeleteSubscription(ctx context.Context, id int) error {

	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	query := `
        DELETE FROM subs
        WHERE id = $1
    `

	result, err := r.ExecContext(ctxTimeout, query, id)
	if err != nil {
		r.l.Error("error deleting subscription", zap.Error(err))
		return subs.ErrFailedToDeleteSubscription
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.l.Error("error getting rows affected", zap.Error(err))
		return subs.ErrFailedToGetRowsAffected
	}

	if rowsAffected == 0 {
		return subs.ErrFailedToFoundSubscription
	}

	return nil
}

func (r *PostgresStorage) ListSubscriptions(ctx context.Context, userID uuid.UUID, page int, pageSize int) ([]subs.Subscription, int, error) {
	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	baseQuery := `
        SELECT 
            id, 
            service_name, 
            price, 
            user_id, 
            start_date, 
            end_date, 
            updated_at
        FROM subs
    	WHERE ($1::uuid IS NULL OR user_id = $1)
		ORDER BY id ASC
		OFFSET $2
        LIMIT $3
    `

	countQuery := `
        SELECT COUNT(*)
        FROM subs
        WHERE ($1::uuid IS NULL OR user_id = $1)
    `

	if userID == uuid.Nil {
		return nil, 0, subs.ErrUserIDIsNil
	}

	var userIDParam any = nil
	if userID != uuid.Nil {
		userIDParam = userID
	}

	var total int
	err := r.GetContext(ctxTimeout, &total, countQuery, userIDParam)
	if err != nil {
		r.l.Error("failed to count subscriptions", zap.Error(err))
		return nil, 0, subs.ErrFailedToCountSubscriptions
	}

	var subsList []subs.Subscription
	err = r.SelectContext(ctxTimeout, &subsList, baseQuery, userIDParam, (page-1)*pageSize, pageSize+1)
	if err != nil {
		r.l.Error("failed to list subscriptions", zap.Error(err))
		return nil, 0, subs.ErrFailedToListSubscriptions
	}

	hasNext := len(subsList) > pageSize
	if hasNext {
		subsList = subsList[:pageSize]
	}

	return subsList, total, nil
}

func (r *PostgresStorage) GetTotalCostByPeriod(ctx context.Context, cost subs.CostCalculation) (int, error) {
	contextTimeOut := 5 * time.Second
	ctxTimeout, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	query := `
        SELECT COALESCE(SUM(price), 0) as total_cost
        FROM subs
        WHERE user_id = $1
          AND ($2::text IS NULL OR service_name ILIKE '%' || $2 || '%')
          AND start_date >= $3
          AND (end_date IS NULL OR end_date <= $4)
    `

	var userIDParam any = nil
	if cost.UserID != uuid.Nil {
		userIDParam = cost.UserID
	}

	var serviceNameParam any = nil
	if cost.ServiceName != "" {
		serviceNameParam = cost.ServiceName
	}

	var total int
	err := r.GetContext(ctxTimeout, &total, query, userIDParam, serviceNameParam, cost.StartDate, cost.EndDate)
	if err != nil {
		r.l.Error("failed to calculate total cost", zap.Error(err))
		return 0, subs.ErrFailedToCalculateTotalCost
	}

	return total, nil
}
