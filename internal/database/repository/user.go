package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	repo "main/internal/database"
	"main/internal/database/model"
)

type UserRepository struct {
	tx pgx.Tx
}

func NewUserRepository(tx pgx.Tx) repo.UserRepository {
	return &UserRepository{tx: tx}
}

func (r *UserRepository) Upsert(ctx context.Context, users []model.User) error {
	return nil
}

func (r *UserRepository) Select(ctx context.Context, tgIDs []int64) ([]model.User, error) {
	return nil, nil
}

func (r *UserRepository) SelectAll(ctx context.Context) ([]model.User, error) {
	return nil, nil
}

func (r *UserRepository) Delete(ctx context.Context, users []model.User) error {
	return nil
}
