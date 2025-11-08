package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type SubscriptionRepository struct {
	tx pgx.Tx
}

func NewSubscriptionRepository(tx pgx.Tx) repo.SubscriptionRepository {
	return &SubscriptionRepository{tx: tx}
}

func (r *SubscriptionRepository) Upsert(ctx context.Context, subscriptions []model.Subscription) error {
	return nil
}

func (r *SubscriptionRepository) Select(ctx context.Context, userTgIDs []int64) ([]model.Subscription, error) {
	return nil, nil
}

func (r *SubscriptionRepository) SelectAll(ctx context.Context) ([]model.Subscription, error) {
	return nil, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, subscriptions []model.Subscription) error {
	return nil
}
