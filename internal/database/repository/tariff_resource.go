package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	repo "main/internal/database"
	"main/internal/model"
)

type TariffResourceRepository struct {
	tx pgx.Tx
}

func NewTariffResourceRepository(tx pgx.Tx) repo.TariffResourceRepository {
	return &TariffResourceRepository{tx: tx}
}

// Assign creates links between tariffs and resources
func (r *TariffResourceRepository) Assign(ctx context.Context, tariffIDs []int64, resourceIDs []int64) error {
	if len(tariffIDs) == 0 || len(resourceIDs) == 0 {
		return nil
	}

	query := squirrel.Insert("tariffs_resources").
		Columns("tariff_id", "resource_id")

	for _, tariffID := range tariffIDs {
		for _, resourceID := range resourceIDs {
			query = query.Values(tariffID, resourceID)
		}
	}

	sql, args, err := query.
		Suffix("ON CONFLICT (tariff_id, resource_id) DO NOTHING").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "query.Suffix.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)
	return errors.Wrap(err, "tx.Exec")
}

// Unassign removes links between tariffs and resources
func (r *TariffResourceRepository) Unassign(ctx context.Context, tariffIDs []int64, resourceIDs []int64) error {
	if len(tariffIDs) == 0 || len(resourceIDs) == 0 {
		return nil
	}

	sql, args, err := squirrel.Delete("tariffs_resources").
		Where(squirrel.And{
			squirrel.Eq{"tariff_id": tariffIDs},
			squirrel.Eq{"resource_id": resourceIDs},
		}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)
	return errors.Wrap(err, "tx.Exec")
}

// SelectResourcesByTariffID returns all resources linked to the given tariffs
func (r *TariffResourceRepository) SelectResourcesByTariffID(ctx context.Context, tariffIDs []int64) ([]model.Resource, error) {
	if len(tariffIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("r.id", "r.chat_id", "r.description").
		From("resources r").
		Join("tariffs_resources tr ON r.id = tr.resource_id").
		Where(squirrel.Eq{"tr.tariff_id": tariffIDs}).
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

	var resources []model.Resource

	for rows.Next() {
		var res model.Resource

		err = rows.Scan(
			&res.Id,
			&res.ChatID,
			&res.Description,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		resources = append(resources, res)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resources, nil
}

// SelectTariffsByResourceID returns all tariffs that grant access to the given resources
func (r *TariffResourceRepository) SelectTariffsByResourceID(ctx context.Context, resourceIDs []int64) ([]model.Tariff, error) {
	if len(resourceIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("t.id", "t.name", "t.price", "t.duration_days").
		From("tariffs t").
		Join("tariffs_resources tr ON t.id = tr.tariff_id").
		Where(squirrel.Eq{"tr.resource_id": resourceIDs}).
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
