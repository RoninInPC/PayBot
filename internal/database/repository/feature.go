package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

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
	return nil, nil
}
