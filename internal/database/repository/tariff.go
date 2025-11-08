package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type TariffRepository struct {
	tx pgx.Tx
}

func NewTariffRepository(tx pgx.Tx) repo.TariffRepository {
	return &TariffRepository{tx: tx}
}

func (r *TariffRepository) Upsert(ctx context.Context, tariffs []model.Tariff) error {
	return nil
}

func (r *TariffRepository) Select(ctx context.Context, names []string) ([]model.Tariff, error) {
	return nil, nil
}

func (r *TariffRepository) SelectAll(ctx context.Context) ([]model.Tariff, error) {
	return nil, nil
}

func (r *TariffRepository) Delete(ctx context.Context, tariffs []model.Tariff) error {
	return nil
}
