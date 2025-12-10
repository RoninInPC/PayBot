package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	repo "main/internal/database"
	"main/internal/model"
)

type PromocodeTariffRepository struct {
	tx pgx.Tx
}

func NewPromocodeTariffRepository(tx pgx.Tx) repo.PromocodeTariffRepository {
	return &PromocodeTariffRepository{tx: tx}
}

// Assign creates links between promocodes and tariffs
func (r *PromocodeTariffRepository) Assign(ctx context.Context, promocodeIDs []int64, tariffIDs []int64) error {
	if len(promocodeIDs) == 0 || len(tariffIDs) == 0 {
		return nil
	}

	query := squirrel.Insert("promocodes_tariffs").
		Columns("promocode_id", "tariff_id")

	for _, promocodeID := range promocodeIDs {
		for _, tariffID := range tariffIDs {
			query = query.Values(promocodeID, tariffID)
		}
	}

	sql, args, err := query.
		Suffix("ON CONFLICT (promocode_id, tariff_id) DO NOTHING").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "query.Suffix.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)
	return errors.Wrap(err, "tx.Exec")
}

// Unassign removes links between promocodes and tariffs
func (r *PromocodeTariffRepository) Unassign(ctx context.Context, promocodeIDs []int64, tariffIDs []int64) error {
	if len(promocodeIDs) == 0 || len(tariffIDs) == 0 {
		return nil
	}

	sql, args, err := squirrel.Delete("promocodes_tariffs").
		Where(squirrel.And{
			squirrel.Eq{"promocode_id": promocodeIDs},
			squirrel.Eq{"tariff_id": tariffIDs},
		}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)
	return errors.Wrap(err, "tx.Exec")
}

// SelectTariffsByPromocodeID returns all tariffs that a promocode can be applied to
func (r *PromocodeTariffRepository) SelectTariffsByPromocodeID(ctx context.Context, promocodeIDs []int64) ([]model.Tariff, error) {
	if len(promocodeIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("t.id", "t.name", "t.price", "t.duration_days").
		From("tariffs t").
		Join("promocodes_tariffs pt ON t.id = pt.tariff_id").
		Where(squirrel.Eq{"pt.promocode_id": promocodeIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "squirrel.Select.From.Join.Where.PlaceholderFormat.ToSql")
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

// SelectPromocodesByTariffID returns all promocodes that can be applied to given tariffs
func (r *PromocodeTariffRepository) SelectPromocodesByTariffID(ctx context.Context, tariffIDs []int64) ([]model.Promocode, error) {
	if len(tariffIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("p.id", "p.code", "p.discount", "p.expires_at", "p.used_count").
		From("promocodes p").
		Join("promocodes_tariffs pt ON p.id = pt.promocode_id").
		Where(squirrel.Eq{"pt.tariff_id": tariffIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "squirrel.Select.From.Join.Where.PlaceholderFormat.ToSql")
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
