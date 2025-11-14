package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"main/internal/database"
	"main/internal/database/testcontainer"
	"main/internal/model"
)

// helper for truncating table after each case
func truncateTariffs(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE tariffs RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func TestTariffRepository_Upsert(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	dbPool, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(dbPool)

	tests := []struct {
		name    string
		tariffs []model.Tariff
		wantErr bool
	}{
		{
			name:    "empty slice",
			tariffs: []model.Tariff{},
			wantErr: false,
		},
		{
			name: "insert single",
			tariffs: []model.Tariff{
				{Name: "basic", Price: 100, DurationDays: 30},
			},
			wantErr: false,
		},
		{
			name: "insert multiple",
			tariffs: []model.Tariff{
				{Name: "silver", Price: 200, DurationDays: 60},
				{Name: "gold", Price: 300, DurationDays: 90},
			},
			wantErr: false,
		},
		{
			name: "insert duplicates",
			tariffs: []model.Tariff{
				{Name: "premium", Price: 500, DurationDays: 120},
				{Name: "premium", Price: 700, DurationDays: 180},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var upserted []model.Tariff

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				upserted, err = uow.TariffRepo().Upsert(ctx, tt.tariffs)
				if err != nil {
					return errors.Wrap(err, "uow.TariffRepo.Upsert")
				}

				return nil
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			require.Len(t, upserted, len(tt.tariffs))

			for i, tariff := range upserted {
				require.NotZero(t, tariff.Id)
				require.Equal(t, tariff.Name, upserted[i].Name)
				require.Equal(t, tariff.Price, upserted[i].Price)
				require.Equal(t, tariff.DurationDays, upserted[i].DurationDays)
			}

			// cleanup
			truncateTariffs(ctx, t, dbPool)
		})
	}
}

func TestTariffRepository_SelectByName(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	dbPool, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(dbPool)

	tests := []struct {
		name       string
		insert     []model.Tariff
		selectName []string
		want       int
	}{
		{
			name:       "empty",
			insert:     nil,
			selectName: nil,
			want:       0,
		},
		{
			name: "single",
			insert: []model.Tariff{
				{Name: "basic", Price: 100, DurationDays: 30},
				{Name: "basic2", Price: 100, DurationDays: 30},
			},
			selectName: []string{"basic"},
			want:       1,
		},
		{
			name: "multiple",
			insert: []model.Tariff{
				{Name: "silver", Price: 200, DurationDays: 60},
				{Name: "gold", Price: 300, DurationDays: 90},
			},
			selectName: []string{"silver", "gold"},
			want:       2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var selected []model.Tariff

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.insert) > 0 {
					upserted, err := uow.TariffRepo().Upsert(ctx, tt.insert)
					require.NoError(t, err)
					require.Len(t, upserted, len(tt.insert))
				}

				selected, err = uow.TariffRepo().SelectByName(ctx, tt.selectName)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			require.Equal(t, tt.want, len(selected))

			for i, trf := range selected {
				assert.Equal(t, tt.insert[i].Name, trf.Name)
				assert.Equal(t, tt.insert[i].Price, trf.Price)
				assert.Equal(t, tt.insert[i].DurationDays, trf.DurationDays)
			}

			// cleanup
			truncateTariffs(ctx, t, dbPool)
		})
	}
}

func TestTariffRepository_SelectByID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	dbPool, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(dbPool)

	tests := []struct {
		name       string
		insert     []model.Tariff
		selectName []string
		want       int
	}{
		{
			name:       "empty",
			insert:     nil,
			selectName: nil,
			want:       0,
		},
		{
			name: "single",
			insert: []model.Tariff{
				{Name: "basic", Price: 100, DurationDays: 30},
				{Name: "basic2", Price: 100, DurationDays: 30},
			},
			selectName: []string{"basic"},
			want:       1,
		},
		{
			name: "multiple",
			insert: []model.Tariff{
				{Name: "silver", Price: 200, DurationDays: 60},
				{Name: "gold", Price: 300, DurationDays: 90},
			},
			selectName: []string{"silver", "gold"},
			want:       2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var selected []model.Tariff

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.insert) > 0 {
					upserted, err := uow.TariffRepo().Upsert(ctx, tt.insert)
					require.NoError(t, err)
					require.Len(t, upserted, len(tt.insert))

					for i, u := range upserted {
						tt.insert[i].Id = u.Id
					}
				}

				selected, err = uow.TariffRepo().SelectByName(ctx, tt.selectName)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			require.Equal(t, tt.want, len(selected))

			for i, trf := range selected {
				assert.Equal(t, tt.insert[i].Name, trf.Name)
				assert.Equal(t, tt.insert[i].Price, trf.Price)
				assert.Equal(t, tt.insert[i].DurationDays, trf.DurationDays)
			}

			// cleanup
			truncateTariffs(ctx, t, dbPool)
		})
	}
}

func TestTariffRepository_SelectAll(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)

	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	tests := []struct {
		name   string
		insert []model.Tariff
		want   int
	}{
		{
			name:   "empty table",
			insert: nil,
			want:   0,
		},
		{
			name: "one tariff",
			insert: []model.Tariff{
				{
					Name:         "basic",
					Price:        100,
					DurationDays: 30,
				},
			},
			want: 1,
		},
		{
			name: "multiple tariffs",
			insert: []model.Tariff{
				{
					Name:         "first",
					Price:        100,
					DurationDays: 30,
				},
				{
					Name:         "second",
					Price:        90,
					DurationDays: 20,
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tariffs []model.Tariff

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.insert) > 0 {
					var upserted []model.Tariff

					upserted, err = uow.TariffRepo().Upsert(ctx, tt.insert)
					require.NoError(t, err)
					require.Len(t, upserted, len(tt.insert))
				}

				tariffs, err = uow.TariffRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(tariffs))

			for i, trf := range tariffs {
				assert.Equal(t, tt.insert[i].Name, trf.Name)
				assert.Equal(t, tt.insert[i].Price, trf.Price)
				assert.Equal(t, tt.insert[i].DurationDays, trf.DurationDays)
			}

			truncateTariffs(ctx, t, db)
		})
	}
}

func TestTariffRepository_Delete(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()
	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)

	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	tests := []struct {
		name     string
		toInsert []model.Tariff
		toDelete []model.Tariff
		wantLeft int
	}{
		{
			name: "delete none",
			toInsert: []model.Tariff{
				{
					Name:         "first",
					Price:        10,
					DurationDays: 10,
				},
			},
			toDelete: []model.Tariff{
				{
					Name:         "second",
					Price:        10,
					DurationDays: 10,
				},
			},
			wantLeft: 1,
		},
		{
			name: "delete one",
			toInsert: []model.Tariff{
				{
					Name:         "first",
					Price:        10,
					DurationDays: 10,
				},
			},
			toDelete: []model.Tariff{
				{
					Name:         "first",
					Price:        10,
					DurationDays: 10,
				},
			},
			wantLeft: 0,
		},
		{
			name: "delete multiple",
			toInsert: []model.Tariff{
				{
					Name:         "first",
					Price:        10,
					DurationDays: 10,
				},
				{
					Name:         "second",
					Price:        20,
					DurationDays: 10,
				},
				{
					Name:         "third",
					Price:        30,
					DurationDays: 10,
				},
			},
			toDelete: []model.Tariff{
				{Name: "first"},
				{Name: "second"},
			},
			wantLeft: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tariffs []model.Tariff

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.toInsert) > 0 {
					upserted, err := uow.TariffRepo().Upsert(ctx, tt.toInsert)
					require.NoError(t, err)
					require.Len(t, upserted, len(tt.toInsert))

					for i, u := range upserted {
						tt.toInsert[i].Id = u.Id
					}
				}

				if len(tt.toDelete) > 0 {
					err = uow.TariffRepo().Delete(ctx, tt.toDelete)
					require.NoError(t, err)
				}

				tariffs, err = uow.TariffRepo().SelectAll(ctx)
				require.NoError(t, err)

				assert.Len(t, tariffs, tt.wantLeft)

				return nil
			})
			require.NoError(t, err)

			truncateTariffs(ctx, t, db)
		})
	}
}
