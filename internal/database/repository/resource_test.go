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
func truncateResources(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE resources RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func TestResourceRepository_Upsert(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now()

	tests := []struct {
		name      string
		resources []model.Resource
		wantErr   bool
	}{
		{
			name:      "пустой слайс",
			resources: []model.Resource{},
			wantErr:   false,
		},
		{
			name: "вставка одного ресурса",
			resources: []model.Resource{
				{
					ChatID:      timePoint.UnixNano() % 1000000,
					Description: "Test channel",
				},
			},
			wantErr: false,
		},
		{
			name: "вставка нескольких ресурсов",
			resources: []model.Resource{
				{
					ChatID:      timePoint.UnixNano()%1000000 + 1,
					Description: "Channel 1",
				},
				{
					ChatID:      timePoint.UnixNano()%1000000 + 2,
					Description: "Channel 2",
				},
			},
			wantErr: false,
		},
		{
			name: "вставка ресурса с пустым описанием",
			resources: []model.Resource{
				{
					ChatID:      timePoint.UnixNano()%1000000 + 100,
					Description: "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var upserted []model.Resource

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				upserted, err = uow.ResourceRepo().Upsert(ctx, tt.resources)
				if err != nil {
					return errors.Wrap(err, "uow.ResourceRepo.Upsert")
				}

				return nil
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, upserted, len(tt.resources))

			// проверяем все поля возвращенных ресурсов
			for i, resource := range upserted {
				require.NotEqual(t, int64(0), resource.Id)
				require.Equal(t, tt.resources[i].ChatID, resource.ChatID)
				require.Equal(t, tt.resources[i].Description, resource.Description)
			}

			truncateResources(ctx, t, db)
		})
	}
}

func TestResourceRepository_Upsert_Update(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	chatID := time.Now().UnixNano()

	// первая вставка
	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		_, err = uow.ResourceRepo().Upsert(ctx, []model.Resource{
			{ChatID: chatID, Description: "Old description"},
		})

		return err
	})
	require.NoError(t, err)

	// обновление (upsert с тем же chat_id)
	var updated []model.Resource
	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		updated, err = uow.ResourceRepo().Upsert(ctx, []model.Resource{
			{ChatID: chatID, Description: "New description"},
		})

		return err
	})
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, "New description", updated[0].Description)

	truncateResources(ctx, t, db)
}

func TestResourceRepository_SelectByChatID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now()

	tests := []struct {
		name    string
		chatIDs []int64
		want    int
		setup   func(context.Context, database.UnitOfWork) []int64 // возвращает chat_ids
	}{
		{
			name:    "пустой слайс chat_ids",
			chatIDs: nil,
			want:    0,
			setup:   nil,
		},
		{
			name:    "несуществующие chat_ids",
			chatIDs: []int64{999999999},
			want:    0,
			setup:   nil,
		},
		{
			name: "один ресурс",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				chatID := timePoint.UnixNano() % 1000000
				_, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: chatID, Description: "Channel"},
				})
				require.NoError(t, err)
				return []int64{chatID}
			},
		},
		{
			name: "несколько ресурсов",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				chatID1 := timePoint.UnixNano() % 1000000
				chatID2 := chatID1 + 1000
				_, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: chatID1, Description: "Channel 1"},
					{ChatID: chatID2, Description: "Channel 2"},
				})
				require.NoError(t, err)
				return []int64{chatID1, chatID2}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources []model.Resource

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.chatIDs = tt.setup(ctx, uow)
				}

				resources, err = uow.ResourceRepo().SelectByChatID(ctx, tt.chatIDs)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(resources))

			// проверяем заполненность полей
			for _, r := range resources {
				assert.NotEqual(t, int64(0), r.Id)
				assert.NotEqual(t, int64(0), r.ChatID)
			}

			truncateResources(ctx, t, db)
		})
	}
}

func TestResourceRepository_SelectByID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now()

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
			name: "один ресурс по id",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				resources, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano() % 1000000, Description: "Test"},
				})
				require.NoError(t, err)
				return []int64{resources[0].Id}
			},
		},
		{
			name: "несколько ресурсов по ids",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				resources, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano() % 1000000, Description: "Test 1"},
					{ChatID: timePoint.UnixNano()%1000000 + 1, Description: "Test 2"},
				})
				require.NoError(t, err)
				return []int64{resources[0].Id, resources[1].Id}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources []model.Resource

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.ids = tt.setup(ctx, uow)
				}

				resources, err = uow.ResourceRepo().SelectByID(ctx, tt.ids)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(resources))

			truncateResources(ctx, t, db)
		})
	}
}

func TestResourceRepository_SelectAll(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now()

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
			name: "один ресурс",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano(), Description: "Channel"},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "несколько ресурсов",
			want: 3,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano(), Description: "Channel 1"},
					{ChatID: timePoint.UnixNano() + 1, Description: "Channel 2"},
					{ChatID: timePoint.UnixNano() + 2, Description: "Channel 3"},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources []model.Resource

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				resources, err = uow.ResourceRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(resources))

			truncateResources(ctx, t, db)
		})
	}
}

func TestResourceRepository_Delete(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now()

	tests := []struct {
		name     string
		wantLeft int
		setup    func(context.Context, database.UnitOfWork) []model.Resource
	}{
		{
			name:     "удаление из пустой таблицы",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Resource {
				return []model.Resource{}
			},
		},
		{
			name:     "удаление одного ресурса",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Resource {
				resources, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano() % 1000000, Description: "To delete"},
				})
				require.NoError(t, err)
				return resources
			},
		},
		{
			name:     "удаление части ресурсов",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Resource {
				resources, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano() % 1000000, Description: "To delete"},
					{ChatID: timePoint.UnixNano()%1000000 + 1, Description: "To keep"},
				})
				require.NoError(t, err)
				// удаляем только первый
				return resources[:1]
			},
		},
		{
			name:     "удаление несуществующих ресурсов (не должно быть ошибки)",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Resource {
				_, err := uow.ResourceRepo().Upsert(ctx, []model.Resource{
					{ChatID: timePoint.UnixNano() % 1000000, Description: "Keep"},
				})
				require.NoError(t, err)
				// возвращаем несуществующий ресурс
				return []model.Resource{{Id: 999999}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources []model.Resource

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				toDelete := tt.setup(ctx, uow)

				err = uow.ResourceRepo().Delete(ctx, toDelete)
				require.NoError(t, err)

				resources, err = uow.ResourceRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantLeft, len(resources))

			truncateResources(ctx, t, db)
		})
	}
}
