package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	repo "main/internal/database"
	"main/internal/model"
)

type RequisiteRepository struct {
	tx pgx.Tx
}

func NewRequisiteRepository(tx pgx.Tx) repo.RequisiteRepository {
	return &RequisiteRepository{tx: tx}
}

func (r *RequisiteRepository) Upsert(ctx context.Context, requisites []model.Requisite) ([]model.Requisite, error) {
	if len(requisites) == 0 {
		return nil, nil
	}

	query := squirrel.Insert("requisites").
		Columns("id", "name", "link", "content", "photo")

	for _, requisite := range requisites {
		query = query.Values(
			requisite.Id,
			requisite.Name,
			requisite.Link,
			requisite.Content,
			requisite.Photo,
		)
	}

	sql, args, err := query.
		Suffix(`ON CONFLICT (link) DO UPDATE SET
			name = EXCLUDED.name,
			content = EXCLUDED.content,
			photo = EXCLUDED.photo
			RETURNING id, name, link, content, photo
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

	upsertedRequisites := make([]model.Requisite, 0, len(requisites))

	for rows.Next() {
		var requisite model.Requisite

		err = rows.Scan(
			&requisite.Id,
			&requisite.Name,
			&requisite.Link,
			&requisite.Content,
			&requisite.Photo,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		upsertedRequisites = append(upsertedRequisites, requisite)
	}

	return upsertedRequisites, nil
}

func (r *RequisiteRepository) SelectByName(ctx context.Context, names []string) ([]model.Requisite, error) {
	if len(names) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "name", "link", "content", "photo").
		From("requisites").
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

	var requisites []model.Requisite

	for rows.Next() {
		var requisite model.Requisite

		err = rows.Scan(
			&requisite.Id,
			&requisite.Name,
			&requisite.Link,
			&requisite.Content,
			&requisite.Photo,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		requisites = append(requisites, requisite)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return requisites, nil
}

func (r *RequisiteRepository) SelectByLink(ctx context.Context, links []string) ([]model.Requisite, error) {
	if len(links) == 0 {
		return nil, nil
	}

	sql, args, err := squirrel.Select("id", "name", "link", "content", "photo").
		From("requisites").
		Where(squirrel.Eq{"link": links}).
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

	var requisites []model.Requisite

	for rows.Next() {
		var requisite model.Requisite

		err = rows.Scan(
			&requisite.Id,
			&requisite.Name,
			&requisite.Link,
			&requisite.Content,
			&requisite.Photo,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		requisites = append(requisites, requisite)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return requisites, nil
}

func (r *RequisiteRepository) SelectAll(ctx context.Context) ([]model.Requisite, error) {
	sql, args, err := squirrel.Select("id", "name", "link", "content", "photo").
		From("requisites").
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

	var requisites []model.Requisite

	for rows.Next() {
		var requisite model.Requisite

		err = rows.Scan(
			&requisite.Id,
			&requisite.Name,
			&requisite.Link,
			&requisite.Content,
			&requisite.Photo,
		)

		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		requisites = append(requisites, requisite)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return requisites, nil
}

func (r *RequisiteRepository) Delete(ctx context.Context, requisites []model.Requisite) error {
	if len(requisites) == 0 {
		return nil
	}

	links := make([]string, 0, len(requisites))

	for _, requisite := range requisites {
		links = append(links, requisite.Link)
	}

	sql, args, err := squirrel.Delete("requisites").
		Where(squirrel.Eq{"link": links}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "squirrel.Delete.Where.PlaceholderFormat.ToSql")
	}

	_, err = r.tx.Exec(ctx, sql, args...)

	return errors.Wrap(err, "tx.Exec")
}
