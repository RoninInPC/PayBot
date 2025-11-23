package repository

import (
	"context"
	"testing"
	"time"

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

// helper для очистки таблицы после теста
func truncatePromocodes(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE promocodes RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func TestPromocodeRepository_Upsert(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	tests := []struct {
		name       string
		promocodes []model.Promocode
		wantErr    bool
	}{
		{
			name:       "пустой слайс",
			promocodes: []model.Promocode{},
			wantErr:    false,
		},
		{
			name: "вставка одного промокода",
			promocodes: []model.Promocode{
				{
					Code:      "SUMMER2024",
					Discount:  20,
					ExpiresAt: timePoint.Add(30 * 24 * time.Hour),
					UsedCount: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "вставка нескольких промокодов",
			promocodes: []model.Promocode{
				{
					Code:      "WINTER10",
					Discount:  10,
					ExpiresAt: timePoint.Add(60 * 24 * time.Hour),
					UsedCount: 5,
				},
				{
					Code:      "SPRING50",
					Discount:  50,
					ExpiresAt: timePoint.Add(90 * 24 * time.Hour),
					UsedCount: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "вставка промокода с максимальной скидкой (100%)",
			promocodes: []model.Promocode{
				{
					Code:      "FREE100",
					Discount:  100,
					ExpiresAt: timePoint.Add(7 * 24 * time.Hour),
					UsedCount: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "вставка промокода со скидкой > 100 (CHECK constraint)",
			promocodes: []model.Promocode{
				{
					Code:      "INVALID",
					Discount:  150,
					ExpiresAt: timePoint,
					UsedCount: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "вставка промокода с отрицательной скидкой (CHECK constraint)",
			promocodes: []model.Promocode{
				{
					Code:      "NEGATIVE",
					Discount:  -10,
					ExpiresAt: timePoint,
					UsedCount: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var upserted []model.Promocode

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				upserted, err = uow.PromocodeRepo().Upsert(ctx, tt.promocodes)
				if err != nil {
					return errors.Wrap(err, "uow.PromocodeRepo.Upsert")
				}

				return nil
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, upserted, len(tt.promocodes))

			// проверяем все поля возвращенных промокодов
			for i, promocode := range upserted {
				require.NotEqual(t, int64(0), promocode.Id)
				require.Equal(t, tt.promocodes[i].Code, promocode.Code)
				require.Equal(t, tt.promocodes[i].Discount, promocode.Discount)
				require.Equal(t, tt.promocodes[i].ExpiresAt.Unix(), promocode.ExpiresAt.Unix())
				require.Equal(t, tt.promocodes[i].UsedCount, promocode.UsedCount)
			}

			truncatePromocodes(ctx, t, db)
		})
	}
}

func TestPromocodeRepository_Upsert_Update(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	// первая вставка
	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		_, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
			{Code: "UPDATE_TEST", Discount: 10, ExpiresAt: timePoint, UsedCount: 5},
		})
		return err
	})
	require.NoError(t, err)

	// обновление (upsert с тем же code)
	var updated []model.Promocode
	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		updated, err = uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
			{Code: "UPDATE_TEST", Discount: 30, ExpiresAt: timePoint.Add(24 * time.Hour), UsedCount: 10},
		})
		return err
	})
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, int64(30), updated[0].Discount)
	require.Equal(t, 10, updated[0].UsedCount)

	truncatePromocodes(ctx, t, db)
}

func TestPromocodeRepository_SelectByCode(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	tests := []struct {
		name  string
		codes []string
		want  int
		setup func(context.Context, database.UnitOfWork) []string // возвращает codes
	}{
		{
			name:  "пустой слайс codes",
			codes: nil,
			want:  0,
			setup: nil,
		},
		{
			name:  "несуществующие codes",
			codes: []string{"NONEXISTENT"},
			want:  0,
			setup: nil,
		},
		{
			name: "один промокод",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []string {
				_, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "SINGLE", Discount: 15, ExpiresAt: timePoint, UsedCount: 0},
				})
				require.NoError(t, err)
				return []string{"SINGLE"}
			},
		},
		{
			name: "несколько промокодов",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []string {
				_, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "FIRST", Discount: 10, ExpiresAt: timePoint, UsedCount: 0},
					{Code: "SECOND", Discount: 20, ExpiresAt: timePoint, UsedCount: 3},
				})
				require.NoError(t, err)
				return []string{"FIRST", "SECOND"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var promocodes []model.Promocode

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.codes = tt.setup(ctx, uow)
				}

				promocodes, err = uow.PromocodeRepo().SelectByCode(ctx, tt.codes)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(promocodes))

			// проверяем заполненность полей
			for _, p := range promocodes {
				assert.NotEqual(t, int64(0), p.Id)
				assert.NotEmpty(t, p.Code)
				assert.GreaterOrEqual(t, p.Discount, int64(0))
				assert.LessOrEqual(t, p.Discount, int64(100))
			}

			truncatePromocodes(ctx, t, db)
		})
	}
}

func TestPromocodeRepository_SelectByID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	tests := []struct {
		name  string
		ids   []int64
		want  int
		setup func(context.Context, database.UnitOfWork) []int64 // возвращает ids
	}{
		{
			name:  "пустой слайс ids",
			ids:   nil,
			want:  0,
			setup: nil,
		},
		{
			name:  "несуществующие ids",
			ids:   []int64{999999},
			want:  0,
			setup: nil,
		},
		{
			name: "один промокод по id",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				promocodes, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "ID_TEST", Discount: 25, ExpiresAt: timePoint, UsedCount: 0},
				})
				require.NoError(t, err)
				return []int64{promocodes[0].Id}
			},
		},
		{
			name: "несколько промокодов по ids",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				promocodes, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "ID_FIRST", Discount: 10, ExpiresAt: timePoint, UsedCount: 0},
					{Code: "ID_SECOND", Discount: 30, ExpiresAt: timePoint, UsedCount: 5},
				})
				require.NoError(t, err)
				return []int64{promocodes[0].Id, promocodes[1].Id}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var promocodes []model.Promocode

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.ids = tt.setup(ctx, uow)
				}

				promocodes, err = uow.PromocodeRepo().SelectByID(ctx, tt.ids)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(promocodes))

			truncatePromocodes(ctx, t, db)
		})
	}
}

func TestPromocodeRepository_SelectAll(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	tests := []struct {
		name  string
		want  int
		setup func(context.Context, database.UnitOfWork)
	}{
		{
			name:  "пустая таблица",
			want:  0,
			setup: nil,
		},
		{
			name: "один промокод",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "ALL_ONE", Discount: 15, ExpiresAt: timePoint, UsedCount: 0},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "несколько промокодов",
			want: 3,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "ALL_FIRST", Discount: 10, ExpiresAt: timePoint, UsedCount: 1},
					{Code: "ALL_SECOND", Discount: 20, ExpiresAt: timePoint, UsedCount: 2},
					{Code: "ALL_THIRD", Discount: 30, ExpiresAt: timePoint, UsedCount: 3},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var promocodes []model.Promocode

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				promocodes, err = uow.PromocodeRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(promocodes))

			truncatePromocodes(ctx, t, db)
		})
	}
}

func TestPromocodeRepository_Delete(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	tests := []struct {
		name     string
		wantLeft int
		setup    func(context.Context, database.UnitOfWork) []model.Promocode
	}{
		{
			name:     "удаление из пустой таблицы",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Promocode {
				return []model.Promocode{}
			},
		},
		{
			name:     "удаление одного промокода",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Promocode {
				promocodes, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "DEL_ONE", Discount: 10, ExpiresAt: timePoint, UsedCount: 0},
				})
				require.NoError(t, err)
				return promocodes
			},
		},
		{
			name:     "удаление части промокодов",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Promocode {
				promocodes, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "DEL_FIRST", Discount: 15, ExpiresAt: timePoint, UsedCount: 0},
					{Code: "DEL_KEEP", Discount: 25, ExpiresAt: timePoint, UsedCount: 5},
				})
				require.NoError(t, err)
				// удаляем только первый
				return promocodes[:1]
			},
		},
		{
			name:     "удаление несуществующих промокодов (не должно быть ошибки)",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Promocode {
				_, err := uow.PromocodeRepo().Upsert(ctx, []model.Promocode{
					{Code: "KEEP_ME", Discount: 50, ExpiresAt: timePoint, UsedCount: 10},
				})
				require.NoError(t, err)
				// возвращаем несуществующий промокод
				return []model.Promocode{{Code: "NONEXISTENT"}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var promocodes []model.Promocode

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				toDelete := tt.setup(ctx, uow)

				err = uow.PromocodeRepo().Delete(ctx, toDelete)
				require.NoError(t, err)

				promocodes, err = uow.PromocodeRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantLeft, len(promocodes))

			truncatePromocodes(ctx, t, db)
		})
	}
}
