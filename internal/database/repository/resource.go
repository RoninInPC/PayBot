package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type ResourceRepository struct {
	tx pgx.Tx
}

func NewResourceRepository(tx pgx.Tx) repo.ResourceRepository {
	return &ResourceRepository{tx: tx}
}

func (r *ResourceRepository) Upsert(ctx context.Context, resources []model.Resource) error {
	return nil
}

func (r *ResourceRepository) Select(ctx context.Context, chatTgIDs []int64) ([]model.Resource, error) {
	return nil, nil
}

func (r *ResourceRepository) SelectAll(ctx context.Context) ([]model.Resource, error) {
	return nil, nil
}

func (r *ResourceRepository) Delete(ctx context.Context, resources []model.Resource) error {
	return nil
}
