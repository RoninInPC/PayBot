package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"main/internal/database"
	"main/internal/database/testcontainer"
	"main/internal/model"
)

// helper for truncating table after each case
func truncateSubscriptions(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func prefillUsers(ctx context.Context, t *testing.T, uowf UnitOfWorkFactory, users []model.User) {
	t.Helper()

	if len(users) == 0 {
		return
	}

	err := uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		upserted, err := uow.UserRepo().Upsert(ctx, users)
		if err != nil {
			return errors.Wrap(err, "uow.UserRepo.Upsert")
		}

		require.Len(t, upserted, len(users))

		return nil
	})
	require.NoError(t, err)
}

func prefillTariffs(ctx context.Context, t *testing.T, uowf UnitOfWorkFactory, tariffs []model.Tariff) (tariffIDs []int64) {
	t.Helper()

	if len(tariffs) == 0 {
		return nil
	}

	var (
		upserted []model.Tariff
		err      error
	)

	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		upserted, err = uow.TariffRepo().Upsert(ctx, tariffs)
		if err != nil {
			return errors.Wrap(err, "uow.TariffRepo.Upsert")
		}

		require.Len(t, upserted, len(tariffs))

		return nil
	})
	require.NoError(t, err)

	ids := make([]int64, 0, len(upserted))
	for _, item := range upserted {
		ids = append(ids, item.Id)
	}

	return ids
}

func TestSubscriptionRepository_Upsert(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	uowf := NewUnitOfWorkFactory(db)

	timePoint1 := time.Now()

	var (
		IvanTgID = timePoint1.UnixNano() % 1000000
		PetrTgID = IvanTgID + 123
	)

	prefillUsers(ctx, t, uowf, []model.User{
		{
			TgID:        IvanTgID,
			Username:    ptr("Ivan"),
			FirstTime:   timePoint1.UTC(),
			TotalSub:    1,
			ContainsSub: true,
		},
		{
			TgID:        PetrTgID,
			Username:    ptr("Petr"),
			FirstTime:   timePoint1.UTC(),
			TotalSub:    1,
			ContainsSub: true,
		},
	})
	require.NoError(t, err)

	prefilledTariffIDs := prefillTariffs(ctx, t, uowf, []model.Tariff{
		{
			Name:         "Tariff 1",
			Price:        100,
			DurationDays: 10,
		},
		{
			Name:         "Tariff 2",
			Price:        200,
			DurationDays: 20,
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name                 string
		prefillSubscriptions []model.Subscription
		subscriptions        []model.Subscription
		wantErr              bool
	}{
		{
			name:          "empty slice",
			subscriptions: []model.Subscription{},
			wantErr:       false,
		},
		{
			name: "insert one full subscription",
			subscriptions: []model.Subscription{
				{
					UserTgID:  IvanTgID,
					TariffID:  prefilledTariffIDs[0],
					StartDate: timePoint1,
					EndDate:   timePoint1.Add(time.Hour * 24 * 60),
					Status:    "test",
				},
			},
			wantErr: false,
		},
		{
			name: "insert two full subscription",
			subscriptions: []model.Subscription{
				{
					UserTgID:  IvanTgID,
					TariffID:  prefilledTariffIDs[0],
					StartDate: timePoint1,
					EndDate:   timePoint1.Add(time.Hour * 24 * 60),
					Status:    "test",
				},
				{
					UserTgID:  PetrTgID,
					TariffID:  prefilledTariffIDs[1],
					StartDate: timePoint1.Add(time.Hour * 24 * 10),
					EndDate:   timePoint1.Add(time.Hour * 24 * 60),
					Status:    "test2",
				},
			},
			wantErr: false,
		},
		{
			name: "update one subscription",
			prefillSubscriptions: []model.Subscription{
				{
					UserTgID:  IvanTgID,
					TariffID:  prefilledTariffIDs[0],
					StartDate: timePoint1,
					EndDate:   timePoint1.Add(time.Hour * 24 * 60),
					Status:    "test",
				},
			},
			subscriptions: []model.Subscription{
				{
					UserTgID:  IvanTgID,
					TariffID:  prefilledTariffIDs[0],
					StartDate: timePoint1.Add(time.Hour * 24 * 10),
					EndDate:   timePoint1.Add(time.Hour * 24 * 30),
					Status:    "UpdTest",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, err)

			var upserted []model.Subscription

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				upserted, err = uow.SubscriptionRepo().Upsert(ctx, tt.subscriptions)
				if err != nil {
					return errors.Wrap(err, "uow.UserRepo.Upsert")
				}

				return nil
			})
			if tt.wantErr {
				require.Error(t, err)

				return
			} else {
				require.NoError(t, err)
			}

			require.Len(t, upserted, len(tt.subscriptions))

			for i, item := range upserted {
				require.NotEqual(t, item.Id, 0)
				require.Equal(t, item.UserTgID, tt.subscriptions[i].UserTgID)
				require.Equal(t, item.TariffID, tt.subscriptions[i].TariffID)
				require.Equal(t, item.StartDate.UnixMicro(), tt.subscriptions[i].StartDate.UnixMicro())
				require.Equal(t, item.EndDate.UnixMicro(), tt.subscriptions[i].EndDate.UnixMicro())
				require.Equal(t, item.Status, tt.subscriptions[i].Status)
			}

			truncateSubscriptions(ctx, t, db)
		})
	}
}

func TestSubscriptionRepository_SelectByUserID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	var (
		user1TgID = timePoint.UnixNano()
		user2TgID = user1TgID + 123
	)

	// подготовка тестовых данных
	prefillUsers(ctx, t, uowf, []model.User{
		{TgID: user1TgID, Username: ptr("user1"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
		{TgID: user2TgID, Username: ptr("user2"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
	})

	tariffIDs := prefillTariffs(ctx, t, uowf, []model.Tariff{
		{Name: "Basic", Price: 1000, DurationDays: 30},
	})

	tests := []struct {
		name      string
		userTgIDs []int64
		want      int
		setup     func(context.Context, database.UnitOfWork)
	}{
		{
			name:      "пустой слайс user_tg_ids",
			userTgIDs: nil,
			want:      0,
			setup:     nil,
		},
		{
			name:      "несуществующие user_tg_ids",
			userTgIDs: []int64{999999},
			want:      0,
			setup:     nil,
		},
		{
			name:      "одна подписка",
			userTgIDs: []int64{user1TgID},
			want:      1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
			},
		},
		{
			name:      "несколько подписок",
			userTgIDs: []int64{user1TgID, user2TgID},
			want:      2,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(24 * time.Hour), Status: "active"},
					{UserTgID: user2TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(48 * time.Hour), Status: "pending"},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subscriptions []model.Subscription

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				subscriptions, err = uow.SubscriptionRepo().SelectByUserID(ctx, tt.userTgIDs)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.want, len(subscriptions))

			// проверяем заполненность полей
			for _, s := range subscriptions {
				require.NotEqual(t, int64(0), s.Id)
				require.NotEqual(t, int64(0), s.UserTgID)
				require.NotEqual(t, int64(0), s.TariffID)
				require.NotEmpty(t, s.Status)
				require.False(t, s.StartDate.IsZero())
			}

			truncateSubscriptions(ctx, t, db)
		})
	}
}

func TestSubscriptionRepository_SelectByTariffID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	var userTgID = timePoint.UnixNano() % 1000000

	prefillUsers(ctx, t, uowf, []model.User{
		{TgID: userTgID, Username: ptr("user"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
	})

	tariffIDs := prefillTariffs(ctx, t, uowf, []model.Tariff{
		{Name: "Premium", Price: 5000, DurationDays: 90},
		{Name: "Standard", Price: 2000, DurationDays: 30},
	})

	tests := []struct {
		name      string
		tariffIDs []int64
		want      int
		setup     func(context.Context, database.UnitOfWork)
	}{
		{
			name:      "пустой слайс tariff_ids",
			tariffIDs: nil,
			want:      0,
			setup:     nil,
		},
		{
			name:      "несуществующие tariff_ids",
			tariffIDs: []int64{999999},
			want:      0,
			setup:     nil,
		},
		{
			name:      "одна подписка с конкретным тарифом",
			tariffIDs: []int64{tariffIDs[0]},
			want:      1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: userTgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(90 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subscriptions []model.Subscription

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				subscriptions, err = uow.SubscriptionRepo().SelectByTariffID(ctx, tt.tariffIDs)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.want, len(subscriptions))

			truncateSubscriptions(ctx, t, db)
		})
	}
}

func TestSubscriptionRepository_SelectAll(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	var (
		user1TgID = timePoint.UnixNano() % 1000000
		user2TgID = user1TgID + 123
	)

	prefillUsers(ctx, t, uowf, []model.User{
		{TgID: user1TgID, Username: ptr("user1"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
		{TgID: user2TgID, Username: ptr("user2"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
	})

	tariffIDs := prefillTariffs(ctx, t, uowf, []model.Tariff{
		{Name: "Test", Price: 1000, DurationDays: 30},
	})

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
			name: "одна подписка",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(30 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "несколько подписок",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(30 * 24 * time.Hour), Status: "active"},
					{UserTgID: user2TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(60 * 24 * time.Hour), Status: "pending"},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subscriptions []model.Subscription

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				subscriptions, err = uow.SubscriptionRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.want, len(subscriptions))

			truncateSubscriptions(ctx, t, db)
		})
	}
}

func TestSubscriptionRepository_Delete(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	var (
		user1TgID = timePoint.UnixNano() % 1000000
		user2TgID = user1TgID + 123
	)

	prefillUsers(ctx, t, uowf, []model.User{
		{TgID: user1TgID, Username: ptr("user1"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
		{TgID: user2TgID, Username: ptr("user2"), FirstTime: timePoint, TotalSub: 0, ContainsSub: false},
	})

	tariffIDs := prefillTariffs(ctx, t, uowf, []model.Tariff{
		{Name: "DeleteTest", Price: 1000, DurationDays: 30},
	})

	tests := []struct {
		name     string
		wantLeft int
		setup    func(context.Context, database.UnitOfWork) []model.Subscription
	}{
		{
			name:     "удаление из пустой таблицы",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Subscription {
				return []model.Subscription{}
			},
		},
		{
			name:     "удаление одной подписки",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Subscription {
				subs, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(30 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
				return subs
			},
		},
		{
			name:     "удаление части подписок",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Subscription {
				subs, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(30 * 24 * time.Hour), Status: "active"},
					{UserTgID: user2TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(60 * 24 * time.Hour), Status: "pending"},
				})
				require.NoError(t, err)
				// удаляем только первую
				return subs[:1]
			},
		},
		{
			name:     "удаление несуществующих подписок (не должно быть ошибки)",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Subscription {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(30 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
				// возвращаем несуществующую подписку
				return []model.Subscription{{UserTgID: 999999}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subscriptions []model.Subscription

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				toDelete := tt.setup(ctx, uow)

				err = uow.SubscriptionRepo().Delete(ctx, toDelete)
				require.NoError(t, err)

				subscriptions, err = uow.SubscriptionRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantLeft, len(subscriptions))

			truncateSubscriptions(ctx, t, db)
		})
	}
}
