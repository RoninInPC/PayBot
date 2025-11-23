package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	repo "main/internal/database"
	"main/internal/model"
)

type FeatureRepository struct {
	tx pgx.Tx
}

func NewFeatureRepository(tx pgx.Tx) repo.FeatureRepository {
	return &FeatureRepository{tx: tx}
}

func (r FeatureRepository) SelectUsersByTariff(ctx context.Context, tariffID int64) ([]model.User, error) {
	sql, args, err := squirrel.Select("u.id", "u.tg_id", "u.username", "u.first_time", "u.total_sub", "u.contains_sub").
		From("users u").
		Join("subscriptions s ON u.tg_id = s.user_tg_id").
		Where(squirrel.Eq{"s.tariff_id": tariffID}).
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

	var users []model.User

	for rows.Next() {
		var u model.User

		err = rows.Scan(
			&u.Id,
			&u.TgID,
			&u.Username,
			&u.FirstTime,
			&u.TotalSub,
			&u.ContainsSub,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return users, nil
}
