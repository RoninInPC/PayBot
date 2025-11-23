package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	repo "main/internal/database"
	"main/internal/model"
)

type ResourceRepository struct {
	tx pgx.Tx
}

func NewResourceRepository(tx pgx.Tx) repo.ResourceRepository {
	return &ResourceRepository{tx: tx}
}

func (r *ResourceRepository) Upsert(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("resources").
		Columns("id", "chat_id", "description")

	for _, resource := range resources {
		query = query.Values(
			resource.Id,
			resource.ChatID,
			resource.Description,
		)
	}

	sql, args, err := query.
		Suffix(`ON CONFLICT (chat_id) DO UPDATE SET
			description = EXCLUDED.description
			RETURNING id, chat_id, description
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

	upsertedResources := make([]model.Resource, 0, len(resources))

	for rows.Next() {
		var resource model.Resource

		err = rows.Scan(
			&resource.Id,
			&resource.ChatID,
			&resource.Description,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upsertedResources = append(upsertedResources, resource)
	}

	return upsertedResources, nil
}

func (r *ResourceRepository) SelectByChatID(ctx context.Context, chatTgIDs []int64) ([]model.Resource, error) {
	if len(chatTgIDs) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "chat_id", "description").
		From("resources").
		Where(squirrel.Eq{"chat_id": chatTgIDs}).
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

	var resources []model.Resource

	for rows.Next() {
		var resource model.Resource

		err = rows.Scan(
			&resource.Id,
			&resource.ChatID,
			&resource.Description,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		resources = append(resources, resource)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resources, nil
}

func (r *ResourceRepository) SelectByID(ctx context.Context, ids []int64) ([]model.Resource, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "chat_id", "description").
		From("resources").
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

	var resources []model.Resource

	for rows.Next() {
		var resource model.Resource

		err = rows.Scan(
			&resource.Id,
			&resource.ChatID,
			&resource.Description,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		resources = append(resources, resource)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resources, nil
}

func (r *ResourceRepository) SelectAll(ctx context.Context) ([]model.Resource, error) {
	sql, args, err := squirrel.Select("id", "chat_id", "description").
		From("resources").
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

	var resources []model.Resource

	for rows.Next() {
		var resource model.Resource

		err = rows.Scan(
			&resource.Id,
			&resource.ChatID,
			&resource.Description,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		resources = append(resources, resource)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return resources, nil
}

func (r *ResourceRepository) Delete(ctx context.Context, resources []model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(resources))

	for _, resource := range resources {
		ids = append(ids, resource.Id)
	}

	sql, args, err := squirrel.Delete("resources").
		Where(squirrel.Eq{"id": ids}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
