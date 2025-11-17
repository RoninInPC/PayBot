package repository

import (
	"context"
	repo "main/internal/database/repository/factory"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"main/internal/model"
)

type PromocodeRepository struct {
	tx pgx.Tx
}

func NewPromocodeRepository(tx pgx.Tx) repo.PromocodeRepository {
	return &PromocodeRepository{tx: tx}
}

func (r *PromocodeRepository) Upsert(ctx context.Context, promocodes []model.Promocode) ([]model.Promocode, error) {
	if len(promocodes) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("promocodes").
		Columns("id", "code", "discount", "expires_at", "used_count")

	for _, promocode := range promocodes {
		query = query.Values(
			promocode.Id,
			promocode.Code,
			promocode.Discount,
			promocode.ExpiresAt,
			promocode.UsedCount,
		)
	}

	sql, args, err := query.
		Suffix(`ON CONFLICT (code) DO UPDATE SET
			code = EXCLUDED.code,
			discount = EXCLUDED.discount,
			expires_at = EXCLUDED.expires_at,
			used_count = EXCLUDED.used_count
			RETURNING id, code, discount, expires_at, used_count
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

	upsertedPromocodes := make([]model.Promocode, 0, len(promocodes))

	for rows.Next() {
		var p model.Promocode

		err = rows.Scan(
			&p.Id,
			&p.Code,
			&p.Discount,
			&p.ExpiresAt,
			&p.UsedCount,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upsertedPromocodes = append(upsertedPromocodes, p)
	}

	return upsertedPromocodes, nil
}

func (r *PromocodeRepository) SelectByCode(ctx context.Context, codes []string) ([]model.Promocode, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "code", "discount", "expires_at", "used_count").
		From("promocodes").
		Where(squirrel.Eq{"code": codes}).
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

	var promocodes []model.Promocode

	for rows.Next() {
		var p model.Promocode

		err = rows.Scan(
			&p.Id,
			&p.Code,
			&p.Discount,
			&p.ExpiresAt,
			&p.UsedCount,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		promocodes = append(promocodes, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return promocodes, nil
}

func (r *PromocodeRepository) SelectByID(ctx context.Context, ids []int64) ([]model.Promocode, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "code", "discount", "expires_at", "used_count").
		From("promocodes").
		Where(squirrel.Eq{"id": ids}).
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

	var promocodes []model.Promocode

	for rows.Next() {
		var p model.Promocode

		err = rows.Scan(
			&p.Id,
			&p.Code,
			&p.Discount,
			&p.ExpiresAt,
			&p.UsedCount,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		promocodes = append(promocodes, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return promocodes, nil
}

func (r *PromocodeRepository) SelectAll(ctx context.Context) ([]model.Promocode, error) {
	sql, args, err := squirrel.Select("id", "code", "discount", "expires_at", "used_count").
		From("promocodes").
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

	var promocodes []model.Promocode

	for rows.Next() {
		var p model.Promocode

		err = rows.Scan(
			&p.Id,
			&p.Code,
			&p.Discount,
			&p.ExpiresAt,
			&p.UsedCount,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		promocodes = append(promocodes, p)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return promocodes, nil
}

func (r *PromocodeRepository) Delete(ctx context.Context, promocodes []model.Promocode) error {
	if len(promocodes) == 0 {
		return nil
	}

	codes := make([]string, 0, len(promocodes))

	for _, promocode := range promocodes {
		codes = append(codes, promocode.Code)
	}

	sql, args, err := squirrel.Delete("promocodes").
		Where(squirrel.Eq{"code": codes}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
