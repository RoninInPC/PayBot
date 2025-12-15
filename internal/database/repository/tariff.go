package repository

import (
	"context"
	repo "main/internal/database/repository/factory"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"main/internal/model"
)

type TariffRepository struct {
	tx pgx.Tx
}

func NewTariffRepository(tx pgx.Tx) repo.TariffRepository {
	return &TariffRepository{tx: tx}
}

func (r *TariffRepository) Upsert(ctx context.Context, tariffs []model.Tariff) ([]model.Tariff, error) {
	if len(tariffs) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("tariffs").
		Columns("name", "price", "duration_days")

	for _, tariff := range tariffs {
		query = query.Values(
			tariff.Name,
			tariff.Price,
			tariff.DurationDays,
		)
	}

	sql, args, err := query.
		Suffix(`ON CONFLICT (name) DO UPDATE SET
			name = EXCLUDED.name,
			price = EXCLUDED.price,
			duration_days = EXCLUDED.duration_days
			RETURNING id, name, price, duration_days
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

	upsertedTariffs := make([]model.Tariff, 0, len(tariffs))

	for rows.Next() {
		var t model.Tariff

		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Price,
			&t.DurationDays,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upsertedTariffs = append(upsertedTariffs, t)
	}

	return upsertedTariffs, nil
}

func (r *TariffRepository) SelectByName(ctx context.Context, names []string) ([]model.Tariff, error) {
	if len(names) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "name", "price", "duration_days").
		From("tariffs").
		Where(squirrel.Eq{"name": names}).
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

	var tariffs []model.Tariff

	for rows.Next() {
		var t model.Tariff

		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Price,
			&t.DurationDays,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		tariffs = append(tariffs, t)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return tariffs, nil
}

func (r *TariffRepository) SelectByID(ctx context.Context, ids []string) ([]model.Tariff, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "name", "price", "duration_days").
		From("tariffs").
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

	var tariffs []model.Tariff

	for rows.Next() {
		var t model.Tariff

		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Price,
			&t.DurationDays,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		tariffs = append(tariffs, t)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return tariffs, nil
}

func (r *TariffRepository) SelectAll(ctx context.Context) ([]model.Tariff, error) {
	sql, args, err := squirrel.Select("id", "name", "price", "duration_days").
		From("tariffs").
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

	var tariffs []model.Tariff

	for rows.Next() {
		var t model.Tariff

		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Price,
			&t.DurationDays,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		tariffs = append(tariffs, t)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return tariffs, nil
}

func (r *TariffRepository) Delete(ctx context.Context, tariffs []model.Tariff) error {
	if len(tariffs) == 0 {
		return nil
	}

	names := make([]string, 0, len(tariffs))

	for _, tariff := range tariffs {
		names = append(names, tariff.Name)
	}

	sql, args, err := squirrel.Delete("tariffs").
		Where(squirrel.Eq{"name": names}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
