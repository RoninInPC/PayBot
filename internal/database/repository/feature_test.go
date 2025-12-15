package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"main/internal/database"
	"main/internal/database/testcontainer"
	"main/internal/model"
)

func TestFeatureRepository_SelectUsersByTariff(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	var (
		user1TgID = timePoint.UnixNano() % 1000000
		user2TgID = user1TgID + 123
		user3TgID = user1TgID + 456
	)

	// Создаем тестовых пользователей
	prefillUsers(ctx, t, uowf, []model.User{
		{TgID: user1TgID, Username: ptr("user1"), FirstTime: timePoint, TotalSub: 1, ContainsSub: true},
		{TgID: user2TgID, Username: ptr("user2"), FirstTime: timePoint, TotalSub: 1, ContainsSub: true},
		{TgID: user3TgID, Username: ptr("user3"), FirstTime: timePoint, TotalSub: 1, ContainsSub: true},
	})

	// Создаем тестовые тарифы
	tariffIDs := prefillTariffs(ctx, t, uowf, []model.Tariff{
		{Name: "Premium", Price: 5000, DurationDays: 90},
		{Name: "Standard", Price: 2000, DurationDays: 30},
	})

	tests := []struct {
		name     string
		tariffID int64
		want     int
		setup    func(context.Context, database.UnitOfWork)
	}{
		{
			name:     "несуществующий tariff_id",
			tariffID: 999999,
			want:     0,
			setup:    nil,
		},
		{
			name:     "тариф без пользователей",
			tariffID: tariffIDs[1],
			want:     0,
			setup:    nil,
		},
		{
			name:     "один пользователь с тарифом",
			tariffID: tariffIDs[0],
			want:     1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(90 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
			},
		},
		{
			name:     "несколько пользователей с одним тарифом",
			tariffID: tariffIDs[0],
			want:     2,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(90 * 24 * time.Hour), Status: "active"},
					{UserTgID: user2TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(90 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
			},
		},
		{
			name:     "смешанные тарифы - выбираем только один",
			tariffID: tariffIDs[0],
			want:     2,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.SubscriptionRepo().Upsert(ctx, []model.Subscription{
					{UserTgID: user1TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(90 * 24 * time.Hour), Status: "active"},
					{UserTgID: user2TgID, TariffID: tariffIDs[0], StartDate: timePoint, EndDate: timePoint.Add(90 * 24 * time.Hour), Status: "active"},
					{UserTgID: user3TgID, TariffID: tariffIDs[1], StartDate: timePoint, EndDate: timePoint.Add(30 * 24 * time.Hour), Status: "active"},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []model.User

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				users, err = uow.FeatureRepo().SelectUsersByTariff(ctx, tt.tariffID)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, len(users))

			// Проверяем заполненность полей возвращенных пользователей
			for _, u := range users {
				assert.NotEqual(t, int64(0), u.Id)
				assert.NotEqual(t, int64(0), u.TgID)
				assert.NotNil(t, u.Username)
				assert.False(t, u.FirstTime.IsZero())
			}

			// Очищаем таблицу подписок после каждого теста
			_, err = db.Exec(ctx, "TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE;")
			require.NoError(t, err)
		})
	}
}
