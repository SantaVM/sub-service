package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"sub-service/internal/model"

	"github.com/lib/pq"
)

type SubscriptionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func New(db *sql.DB, logger *slog.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, input model.CreateSubscription) (*model.Subscription, error) {
	r.logger.InfoContext(ctx, "creating subscription", "service_name", input.ServiceName, "user_id", input.UserID)

	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	sub := &model.Subscription{}
	err := r.db.QueryRowContext(ctx, query,
		input.ServiceName,
		input.Price,
		input.UserID,
		input.StartDate,
		input.EndDate,
	).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		r.logger.ErrorContext(ctx, "failed to create subscription", "error", err)
		return nil, mapPostgresError(err)
	}

	r.logger.InfoContext(ctx, "subscription created successfully", "subscription_id", sub.ID)
	return sub, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uint) (*model.Subscription, error) {
	r.logger.InfoContext(ctx, "getting subscription", "subscription_id", id)

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	sub := &model.Subscription{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.WarnContext(ctx, "subscription not found", "subscription_id", id)

			return nil, model.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "failed to get subscription", "error", err)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	r.logger.InfoContext(ctx, "subscription retrieved successfully", "subscription_id", id)
	return sub, nil
}

// TODO: add fields for SORT and ORDER??
func (r *SubscriptionRepository) List(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error) {
	r.logger.InfoContext(ctx, "listing subscriptions", "user_id", query.UserID, "service_name", query.ServiceName)

	offset := (query.Page - 1) * query.Size

	baseQuery := `
		FROM subscriptions
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if query.UserID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argNum)
		args = append(args, *query.UserID)
		argNum++
	}

	if query.ServiceName != nil {
		baseQuery += fmt.Sprintf(" AND service_name ILIKE $%d", argNum)
		args = append(args, "%"+*query.ServiceName+"%")
		argNum++
	}

	// --- COUNT ---
	countQuery := "SELECT COUNT(*) " + baseQuery

	var totalElements int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalElements)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to count subscriptions", "error", err)
		return nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// stop if totalElements < offset
	if totalElements < int64(offset) {
		r.logger.InfoContext(ctx, "no more pages",
			"total", totalElements,
			"page", query.Page,
			"size", query.Size,
			"offset", offset,
		)

		sub := []*model.Subscription{}
		page := model.NewPage(sub, query.Size, query.Page, totalElements)

		return page, nil
	}

	// --- DATA ---
	dataQuery := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	` + baseQuery +
		fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)

	argsWithPagination := append(args, query.Size, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, argsWithPagination...)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to list subscriptions", "error", err)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		sub := &model.Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			r.logger.ErrorContext(ctx, "failed to scan subscription", "error", err)
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	numberOfElements := len(subscriptions)

	page := model.NewPage(subscriptions, query.Size, query.Page, totalElements)

	r.logger.InfoContext(ctx, "subscriptions retrieved successfully",
		"count", numberOfElements,
		"total", totalElements,
		"page", query.Page,
		"size", query.Size,
	)

	return page, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, id uint, input model.UpdateSubscription) (*model.Subscription, error) {
	r.logger.InfoContext(ctx, "updating subscription", "subscription_id", id)

	// TODO: refactor to func Named(name string, value any) ?

	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	if input.ServiceName != nil {
		setClauses = append(setClauses, fmt.Sprintf("service_name = $%d", argNum))
		args = append(args, *input.ServiceName)
		argNum++
	}

	if input.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argNum))
		args = append(args, *input.Price)
		argNum++
	}

	if input.StartDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_date = $%d", argNum))
		args = append(args, *input.StartDate)
		argNum++
	}

	if input.EndDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_date = $%d", argNum))
		args = append(args, *input.EndDate)
		argNum++
	}

	if len(setClauses) == 0 {
		r.logger.WarnContext(ctx, "no fields to update", "subscription_id", id)
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE subscriptions
		SET %s
		WHERE id = $%d
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`, strings.Join(setClauses, ", "), argNum)

	sub := &model.Subscription{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.WarnContext(ctx, "subscription not found", "subscription_id", id)
			return nil, model.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "failed to update subscription", "error", err)
		return nil, mapPostgresError(err)
	}

	r.logger.InfoContext(ctx, "subscription updated successfully", "subscription_id", id)
	return sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uint) error {
	r.logger.InfoContext(ctx, "deleting subscription", "subscription_id", id)

	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to delete subscription", "error", err)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.WarnContext(ctx, "subscription not found", "subscription_id", id)
		return fmt.Errorf("subscription with id: %d not found %w", id, model.ErrNotFound)
	}

	r.logger.InfoContext(ctx, "subscription deleted successfully", "subscription_id", id)
	return nil
}

func (r *SubscriptionRepository) GetTotalCost(ctx context.Context, query model.TotalCostReq) (int, error) {
	r.logger.InfoContext(
		ctx,
		"calculating total cost",
		"user_id", query.UserID,
		"service_name", query.ServiceName,
		"start_date", query.StartDate,
		"end_date", query.EndDate,
	)

	sqlQuery := `
		SELECT COALESCE(SUM(
			price * GREATEST(
				0,
				(
					DATE_PART('year', age(upper(r), lower(r))) * 12 +
					DATE_PART('month', age(upper(r), lower(r)))
				)::int
			)
		), 0)
		FROM (
			SELECT 
				price,
				daterange(start_date, COALESCE(end_date, 'infinity')) 
				* daterange($1, $2) AS r
			FROM subscriptions
			WHERE 1=1
	`

	args := []interface{}{query.StartDate, query.EndDate}
	argNum := 3

	// Фильтр по user_id
	if query.UserID != nil {
		sqlQuery += fmt.Sprintf(" AND user_id = $%d", argNum)
		args = append(args, *query.UserID)
		argNum++
	}

	// Фильтр по service_name
	if query.ServiceName != nil {
		sqlQuery += fmt.Sprintf(" AND service_name ILIKE $%d", argNum)
		args = append(args, "%"+*query.ServiceName+"%")
		argNum++
	}

	// ВАЖНО: используем оператор пересечения диапазонов (под GiST индекс)
	sqlQuery += `
			AND daterange(start_date, COALESCE(end_date, 'infinity')) 
			    && daterange($1, $2)
		) t
		WHERE r IS NOT NULL
	`

	var totalCost int
	err := r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&totalCost)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to calculate total cost", "error", err)
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	r.logger.InfoContext(ctx, "total cost calculated successfully", "total_cost", totalCost)
	return totalCost, nil
}

func mapPostgresError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}

	switch pqErr.Code {
	case "23514":
		if pqErr.Constraint == "check_end_date_after_start_date" {
			return model.ErrInvalidDateRange
		}
	case "23P01":
		if pqErr.Constraint == "no_overlapping_subscriptions" {
			return model.ErrSubscriptionOverlap
		}
	}

	return err
}
