package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	repo "main/internal/database"
	"main/internal/model"
)

type SubscriptionRepository struct {
	tx pgx.Tx
}

func NewSubscriptionRepository(tx pgx.Tx) repo.SubscriptionRepository {
	return &SubscriptionRepository{tx: tx}
}

func (r *SubscriptionRepository) Upsert(ctx context.Context, subscriptions []model.Subscription) ([]model.Subscription, error) {
	if len(subscriptions) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("subscriptions").
		Columns("user_tg_id", "tariff_id", "start_date", "end_date", "status")

	for _, user := range subscriptions {
		query = query.Values(
			user.UserTgID,
			user.TariffID,
			user.StartDate,
			user.EndDate,
			user.Status,
		)
	}

	sql, args, err := query.
		Suffix(`ON CONFLICT (user_tg_id) DO UPDATE SET
			tariff_id = EXCLUDED.tariff_id,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			status = EXCLUDED.status
			RETURNING id, user_tg_id, tariff_id, start_date, end_date, status
`).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "query.Suffix.PlaceholderFormat.ToSql")
	}

	rows, err := r.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}

	upsertedSubscriptions := make([]model.Subscription, 0, len(subscriptions))

	for rows.Next() {
		var s model.Subscription

		err = rows.Scan(
			&s.Id,
			&s.UserTgID,
			&s.TariffID,
			&s.StartDate,
			&s.EndDate,
			&s.Status,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upsertedSubscriptions = append(upsertedSubscriptions, s)
	}

	return upsertedSubscriptions, nil
}

func (r *SubscriptionRepository) SelectByUserID(ctx context.Context, userTgIDs []int64) ([]model.Subscription, error) {
	if len(userTgIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "user_tg_id", "tariff_id", "start_date", "end_date", "status").
		From("subscriptions").
		Where(squirrel.Eq{"user_tg_id": userTgIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "squirrel.Select.From.Where.PlaceholderFormat.ToSql")
	}

	rows, err := r.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var s model.Subscription

		err = rows.Scan(
			&s.Id,
			&s.UserTgID,
			&s.TariffID,
			&s.StartDate,
			&s.EndDate,
			&s.Status,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		subscriptions = append(subscriptions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) SelectByTariffID(ctx context.Context, tariffIDs []int64) ([]model.Subscription, error) {
	if len(tariffIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "user_tg_id", "tariff_id", "start_date", "end_date", "status").
		From("subscriptions").
		Where(squirrel.Eq{"tariff_id": tariffIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "squirrel.Select.From.Where.PlaceholderFormat.ToSql")
	}

	rows, err := r.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var s model.Subscription

		err = rows.Scan(
			&s.Id,
			&s.UserTgID,
			&s.TariffID,
			&s.StartDate,
			&s.EndDate,
			&s.Status,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		subscriptions = append(subscriptions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) SelectAll(ctx context.Context) ([]model.Subscription, error) {
	sql, args, err := squirrel.Select("id", "user_tg_id", "tariff_id", "start_date", "end_date", "status").
		From("subscriptions").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "squirrel.Select.From.PlaceholderFormat.ToSql")
	}

	rows, err := r.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var s model.Subscription

		err = rows.Scan(
			&s.Id,
			&s.UserTgID,
			&s.TariffID,
			&s.StartDate,
			&s.EndDate,
			&s.Status,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		subscriptions = append(subscriptions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, subscriptions []model.Subscription) error {
	if len(subscriptions) == 0 {
		return nil
	}

	tgIDs := make([]int64, 0, len(subscriptions))

	for _, sub := range subscriptions {
		tgIDs = append(tgIDs, sub.UserTgID)
	}

	sql, args, err := squirrel.Delete("users").
		Where(squirrel.Eq{"user_tg_id": tgIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
