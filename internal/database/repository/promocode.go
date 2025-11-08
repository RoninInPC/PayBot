package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type PromocodeRepository struct {
	tx pgx.Tx
}

func NewPromocodeRepository(tx pgx.Tx) repo.PromocodeRepository {
	return &PromocodeRepository{tx: tx}
}

func (r *PromocodeRepository) Upsert(ctx context.Context, promocodes []model.Promocode) error {
	return nil
}

func (r *PromocodeRepository) Select(ctx context.Context, codes []string) ([]model.Promocode, error) {
	return nil, nil
}

func (r *PromocodeRepository) SelectAll(ctx context.Context) ([]model.Promocode, error) {
	return nil, nil
}

func (r *PromocodeRepository) Delete(ctx context.Context, promocodes []model.Promocode) error {
	return nil
}
