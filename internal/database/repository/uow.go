package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"main/internal/database"
)

type unitOfWork struct {
	tx       pgx.Tx
	userRepo database.UserRepository
}

func (u unitOfWork) UserRepo() database.UserRepository {
	return u.userRepo
}

func (u unitOfWork) Commit(ctx context.Context) error {
	if err := u.tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}

	return nil
}

func (u unitOfWork) Rollback(ctx context.Context) error {
	if err := u.tx.Rollback(ctx); err != nil {
		return errors.Wrap(err, "tx.Rollback")
	}

	return nil
}

type UnitOfWorkFactory struct {
	pool *pgxpool.Pool
}

func NewUnitOfWorkFactory(pool *pgxpool.Pool) UnitOfWorkFactory {
	return UnitOfWorkFactory{
		pool: pool,
	}
}

func (f UnitOfWorkFactory) New(ctx context.Context, level pgx.TxIsoLevel, fn func(uow database.UnitOfWork) error) error {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: level})
	if err != nil {
		return errors.Wrap(err, "pool.BeginTx")
	}

	uow := f.createUnitOfWork(tx)

	if err = fn(uow); err != nil {
		if rollbackErr := uow.Rollback(ctx); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "uow.Rollback")
		}

		return errors.Wrap(err, "fn(uow)")
	}

	err = uow.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "uow.Commit")
	}

	return nil
}

func (f UnitOfWorkFactory) createUnitOfWork(tx pgx.Tx) database.UnitOfWork {
	return unitOfWork{
		tx:       tx,
		userRepo: NewUserRepository(tx),
	}
}
