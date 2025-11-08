package database

import (
	"context"

	"github.com/jackc/pgx/v5"

	"main/internal/database/model"
)

type UserRepository interface {
	Upsert(ctx context.Context, users []model.User) error
	Select(ctx context.Context, tgIDs []int64) ([]model.User, error)
	SelectAll(ctx context.Context) ([]model.User, error)
	Delete(ctx context.Context, users []model.User) error
}

type PaymentRepository interface {
	Upsert(ctx context.Context, payments []model.Payment) error
	Select(ctx context.Context, userTgIDs []int64) ([]model.Payment, error)
	SelectAll(ctx context.Context) ([]model.Payment, error)
	Delete(ctx context.Context, payments []model.Payment) error
}

type SubscriptionRepository interface {
	Upsert(ctx context.Context, subscriptions []model.Subscription) error
	Select(ctx context.Context, userTgIDs []int64) ([]model.Subscription, error)
	SelectAll(ctx context.Context) ([]model.Subscription, error)
	Delete(ctx context.Context, subscriptions []model.Subscription) error
}

type TariffRepository interface {
	Upsert(ctx context.Context, tariffs []model.Tariff) error
	Select(ctx context.Context, names []string) ([]model.Tariff, error)
	SelectAll(ctx context.Context) ([]model.Tariff, error)
	Delete(ctx context.Context, tariffs []model.Tariff) error
}

type ResourceRepository interface {
	Upsert(ctx context.Context, resources []model.Resource) error
	Select(ctx context.Context, chatTgIDs []int64) ([]model.Resource, error)
	SelectAll(ctx context.Context) ([]model.Resource, error)
	Delete(ctx context.Context, resources []model.Resource) error
}

type PromocodeRepository interface {
	Upsert(ctx context.Context, promocodes []model.Promocode) error
	Select(ctx context.Context, codes []string) ([]model.Promocode, error)
	SelectAll(ctx context.Context) ([]model.Promocode, error)
	Delete(ctx context.Context, promocodes []model.Promocode) error
}

type RequisiteRepository interface {
	Upsert(ctx context.Context, requisites []model.Requisite) error
	Select(ctx context.Context, names []string) ([]model.Requisite, error)
	SelectAll(ctx context.Context) ([]model.Requisite, error)
	Delete(ctx context.Context, requisites []model.Requisite) error
}

type UnitOfWork interface {
	UserRepo() UserRepository

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWorkFactory interface {
	New(ctx context.Context, isoLevel pgx.TxIsoLevel) (UnitOfWork, error)
}
