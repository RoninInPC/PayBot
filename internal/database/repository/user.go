package repository

import (
	"context"
	repo "main/internal/database/repository/factory"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"main/internal/model"
)

type UserRepository struct {
	tx pgx.Tx
}

func NewUserRepository(tx pgx.Tx) repo.UserRepository {
	return &UserRepository{tx: tx}
}

func (r *UserRepository) Upsert(ctx context.Context, users []model.User) ([]model.User, error) {
	if len(users) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("users").
		Columns("tg_id", "username", "first_time", "total_sub", "contains_sub", "promocode_id")

	for _, user := range users {
		query = query.Values(
			user.TgID,
			user.Username,
			user.FirstTime,
			user.TotalSub,
			user.ContainsSub,
			user.PromocodeID,
		)
	}

	sql, args, err := query.
		Suffix(`ON CONFLICT (tg_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_time = EXCLUDED.first_time,
			total_sub = EXCLUDED.total_sub,
			contains_sub = EXCLUDED.contains_sub,
			promocode_id = EXCLUDED.promocode_id
			RETURNING id, tg_id, username, first_time, total_sub, contains_sub, promocode_id
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

	upsertedUsers := make([]model.User, 0, len(users))

	for rows.Next() {
		var u model.User

		err = rows.Scan(
			&u.Id,
			&u.TgID,
			&u.Username,
			&u.FirstTime,
			&u.TotalSub,
			&u.ContainsSub,
			&u.PromocodeID,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upsertedUsers = append(upsertedUsers, u)
	}

	return upsertedUsers, nil
}

func (r *UserRepository) SelectByTgID(ctx context.Context, tgIDs []int64) ([]model.User, error) {
	if len(tgIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "tg_id", "username", "first_time", "total_sub", "contains_sub", "promocode_id").
		From("users").
		Where(squirrel.Eq{"tg_id": tgIDs}).
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
			&u.PromocodeID,
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

func (r *UserRepository) SelectByUsername(ctx context.Context, userNames []string) ([]model.User, error) {
	if len(userNames) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "tg_id", "username", "first_time", "total_sub", "contains_sub", "promocode_id").
		From("users").
		Where(squirrel.Eq{"username": userNames}).
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
			&u.PromocodeID,
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

func (r *UserRepository) SelectAll(ctx context.Context) ([]model.User, error) {
	sql, args, err := squirrel.Select("id", "tg_id", "username", "first_time", "total_sub", "contains_sub", "promocode_id").
		From("users").
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
			&u.PromocodeID,
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

// SelectBySubscriptionStatus returns users filtered by subscription status
func (r *UserRepository) SelectBySubscriptionStatus(ctx context.Context, hasSubscription bool) ([]model.User, error) {
	sql, args, err := squirrel.Select("id", "tg_id", "username", "first_time", "total_sub", "contains_sub", "promocode_id").
		From("users").
		Where(squirrel.Eq{"contains_sub": hasSubscription}).
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
			&u.PromocodeID,
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

// SelectByPromocodeID returns users who used specific promocodes
func (r *UserRepository) SelectByPromocodeID(ctx context.Context, promocodeIDs []int64) ([]model.User, error) {
	if len(promocodeIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "tg_id", "username", "first_time", "total_sub", "contains_sub", "promocode_id").
		From("users").
		Where(squirrel.Eq{"promocode_id": promocodeIDs}).
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
			&u.PromocodeID,
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

func (r *UserRepository) Delete(ctx context.Context, users []model.User) error {
	if len(users) == 0 {
		return nil
	}

	tgIDs := make([]int64, 0, len(users))

	for _, user := range users {
		tgIDs = append(tgIDs, user.TgID)
	}

	sql, args, err := squirrel.Delete("users").
		Where(squirrel.Eq{"tg_id": tgIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
