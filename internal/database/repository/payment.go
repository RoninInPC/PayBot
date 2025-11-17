package repository

import (
	"context"
	repo "main/internal/database/repository/factory"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"main/internal/model"
)

type PaymentRepository struct {
	tx pgx.Tx
}

func NewPaymentRepository(tx pgx.Tx) repo.PaymentRepository {
	return &PaymentRepository{tx: tx}
}

func (r *PaymentRepository) Upsert(ctx context.Context, payments []model.Payment) ([]model.Payment, error) {
	if len(payments) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("payments").
		Columns("user_tg_id", "amount", "timestamp", "status", "receipt_photo")

	for _, payment := range payments {
		query = query.Values(
			payment.UserTgID,
			payment.Amount,
			payment.Timestamp,
			payment.Status,
			payment.ReceiptPhoto,
		)
	}

	sql, args, err := query.
		Suffix(`
		ON CONFLICT (user_tg_id) DO UPDATE SET
			amount = EXCLUDED.amount,
			timestamp = EXCLUDED.timestamp,
			status = EXCLUDED.status,
			receipt_photo = EXCLUDED.receipt_photo
		RETURNING id, user_tg_id, amount, timestamp, status, receipt_photo`).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "query.Suffix.PlaceholderFormat.ToSql")
	}

	rows, err := r.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	var upserted []model.Payment

	for rows.Next() {
		var p model.Payment

		err = rows.Scan(
			&p.Id,
			&p.UserTgID,
			&p.Amount,
			&p.Timestamp,
			&p.Status,
			&p.ReceiptPhoto,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upserted = append(upserted, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return upserted, nil
}

func (r *PaymentRepository) SelectByUserTgID(ctx context.Context, userTgIDs []int64) ([]model.Payment, error) {
	if len(userTgIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "user_id", "amount", "timestamp", "status", "receipt_photo").
		From("payments").
		Where(squirrel.Eq{"user_id": userTgIDs}).
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

	var payments []model.Payment
	for rows.Next() {
		var p model.Payment

		err = rows.Scan(
			&p.Id,
			&p.UserTgID,
			&p.Amount,
			&p.Timestamp,
			&p.Status,
			&p.ReceiptPhoto,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		payments = append(payments, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return payments, nil
}

func (r *PaymentRepository) SelectAll(ctx context.Context) ([]model.Payment, error) {
	sql, args, err := squirrel.Select("id", "user_id", "amount", "timestamp", "status", "receipt_photo").
		From("payments").
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

	var payments []model.Payment
	for rows.Next() {
		var p model.Payment

		err = rows.Scan(
			&p.Id,
			&p.UserTgID,
			&p.Amount,
			&p.Timestamp,
			&p.Status,
			&p.ReceiptPhoto,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		payments = append(payments, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return payments, nil
}

func (r *PaymentRepository) Delete(ctx context.Context, payments []model.Payment) error {
	if len(payments) == 0 {
		return nil
	}

	userIDs := make([]int64, 0, len(payments))
	for _, payment := range payments {
		userIDs = append(userIDs, payment.UserTgID)
	}

	sql, args, err := squirrel.Delete("payments").
		Where(squirrel.Eq{"user_id": userIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
