package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type RequisiteRepository struct {
	tx pgx.Tx
}

func NewRequisiteRepository(tx pgx.Tx) repo.RequisiteRepository {
	return &RequisiteRepository{tx: tx}
}

func (r *RequisiteRepository) Upsert(ctx context.Context, requisites []model.Requisite) error {
	return nil
}

func (r *RequisiteRepository) Select(ctx context.Context, names []string) ([]model.Requisite, error) {
	return nil, nil
}

func (r *RequisiteRepository) SelectAll(ctx context.Context) ([]model.Requisite, error) {
	return nil, nil
}

func (r *RequisiteRepository) Delete(ctx context.Context, requisites []model.Requisite) error {
	return nil
}
