package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type PaymentRepository struct {
	tx pgx.Tx
}

func NewPaymentRepository(tx pgx.Tx) repo.PaymentRepository {
	return &PaymentRepository{tx: tx}
}

func (r *PaymentRepository) Upsert(ctx context.Context, payments []model.Payment) error {
	return nil
}

func (r *PaymentRepository) Select(ctx context.Context, userTgIDs []int64) ([]model.Payment, error) {
	return nil, nil
}

func (r *PaymentRepository) SelectAll(ctx context.Context) ([]model.Payment, error) {
	return nil, nil
}

func (r *PaymentRepository) Delete(ctx context.Context, users []model.Payment) error {
	return nil
}
