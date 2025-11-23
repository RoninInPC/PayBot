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

// helper для очистки таблиц после теста
func truncatePayments(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE payments, users RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

// helper для создания тестового пользователя
func createTestUser(ctx context.Context, t *testing.T, uow database.UnitOfWork, tgID int64) model.User {
	users, err := uow.UserRepo().Upsert(ctx, []model.User{
		{
			TgID:        tgID,
			Username:    ptr("test_user"),
			FirstTime:   time.Now().UTC(),
			TotalSub:    0,
			ContainsSub: false,
		},
	})
	require.NoError(t, err)
	require.Len(t, users, 1)
	return users[0]
}

func TestPaymentRepository_Upsert(t *testing.T) {
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
		name     string
		payments []model.Payment
		wantErr  bool
		setup    func(context.Context, database.UnitOfWork) int64 // возвращает user tg_id
	}{
		{
			name:     "пустой слайс",
			payments: []model.Payment{},
			wantErr:  false,
			setup:    nil,
		},
		{
			name: "вставка одного платежа",
			payments: []model.Payment{
				{
					UserTgID:     0, // будет установлен в setup
					Amount:       10000,
					Timestamp:    timePoint,
					Status:       "pending",
					ReceiptPhoto: []byte("photo_data"),
				},
			},
			wantErr: false,
			setup: func(ctx context.Context, uow database.UnitOfWork) int64 {
				user := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				return user.TgID
			},
		},
		{
			name: "вставка нескольких платежей для одного пользователя",
			payments: []model.Payment{
				{
					UserTgID:     0,
					Amount:       5000,
					Timestamp:    timePoint,
					Status:       "completed",
					ReceiptPhoto: nil,
				},
				{
					UserTgID:     0,
					Amount:       7500,
					Timestamp:    timePoint.Add(time.Hour),
					Status:       "pending",
					ReceiptPhoto: []byte("receipt"),
				},
			},
			wantErr: false,
			setup: func(ctx context.Context, uow database.UnitOfWork) int64 {
				user := createTestUser(ctx, t, uow, timePoint.Add(time.Second).UnixNano()%1000000)
				return user.TgID
			},
		},
		{
			name: "вставка платежа без существующего пользователя (FK violation)",
			payments: []model.Payment{
				{
					UserTgID:  999999999,
					Amount:    1000,
					Timestamp: timePoint,
					Status:    "pending",
				},
			},
			wantErr: true,
			setup:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var upserted []model.Payment

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				var userTgID int64
				if tt.setup != nil {
					userTgID = tt.setup(ctx, uow)
					// устанавливаем user_tg_id для всех платежей
					for i := range tt.payments {
						tt.payments[i].UserTgID = userTgID
					}
				}

				upserted, err = uow.PaymentRepo().Upsert(ctx, tt.payments)
				if err != nil {
					return errors.Wrap(err, "uow.PaymentRepo.Upsert")
				}

				return nil
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, upserted, len(tt.payments))

			// проверяем все поля возвращенных платежей
			for i, payment := range upserted {
				require.NotEqual(t, int64(0), payment.Id)
				require.Equal(t, tt.payments[i].UserTgID, payment.UserTgID)
				require.Equal(t, tt.payments[i].Amount, payment.Amount)
				require.Equal(t, tt.payments[i].Timestamp.Unix(), payment.Timestamp.Unix())
				require.Equal(t, tt.payments[i].Status, payment.Status)
				assert.Equal(t, tt.payments[i].ReceiptPhoto, payment.ReceiptPhoto)
			}

			truncatePayments(ctx, t, db)
		})
	}
}

func TestPaymentRepository_SelectByUserTgID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)
	timePoint := time.Now().UTC()

	tests := []struct {
		name      string
		userTgIDs []int64
		want      int
		setup     func(context.Context, database.UnitOfWork) []int64 // возвращает tg_ids пользователей
	}{
		{
			name:      "пустой слайс tg_ids",
			userTgIDs: nil,
			want:      0,
			setup:     nil,
		},
		{
			name:      "несуществующие tg_ids",
			userTgIDs: []int64{999999, 888888},
			want:      0,
			setup:     nil,
		},
		{
			name: "один платеж пользователя",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				user := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				_, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{
						UserTgID:  user.TgID,
						Amount:    5000,
						Timestamp: timePoint,
						Status:    "completed",
					},
				})
				require.NoError(t, err)
				return []int64{user.TgID}
			},
		},
		{
			name: "несколько платежей нескольких пользователей",
			want: 3,
			setup: func(ctx context.Context, uow database.UnitOfWork) []int64 {
				user1 := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				user2 := createTestUser(ctx, t, uow, timePoint.Add(1251*time.Microsecond).UnixNano()%1000000)

				_, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{
						UserTgID:  user1.TgID,
						Amount:    1000,
						Timestamp: timePoint,
						Status:    "pending",
					},
					{
						UserTgID:  user1.TgID,
						Amount:    2000,
						Timestamp: timePoint.Add(time.Hour),
						Status:    "completed",
					},
					{
						UserTgID:  user2.TgID,
						Amount:    3000,
						Timestamp: timePoint,
						Status:    "pending",
					},
				})
				require.NoError(t, err)
				return []int64{user1.TgID, user2.TgID}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payments []model.Payment

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				var tgIDs []int64
				if tt.setup != nil {
					tgIDs = tt.setup(ctx, uow)
					tt.userTgIDs = tgIDs
				}

				payments, err = uow.PaymentRepo().SelectByUserTgID(ctx, tt.userTgIDs)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(payments))

			// проверяем что все поля заполнены
			for _, p := range payments {
				assert.NotEqual(t, int64(0), p.Id)
				assert.NotEqual(t, int64(0), p.UserTgID)
				assert.NotEqual(t, int64(0), p.Amount)
				assert.NotEmpty(t, p.Status)
				assert.False(t, p.Timestamp.IsZero())
			}

			truncatePayments(ctx, t, db)
		})
	}
}

func TestPaymentRepository_SelectAll(t *testing.T) {
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
			name: "один платеж",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				user := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				_, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{
						UserTgID:  user.TgID,
						Amount:    1000,
						Timestamp: timePoint,
						Status:    "pending",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "несколько платежей",
			want: 3,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				user1 := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				user2 := createTestUser(ctx, t, uow, timePoint.Add(1251*time.Microsecond).UnixNano()%1000000)

				_, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{UserTgID: user1.TgID, Amount: 1000, Timestamp: timePoint, Status: "completed"},
					{UserTgID: user1.TgID, Amount: 2000, Timestamp: timePoint, Status: "pending"},
					{UserTgID: user2.TgID, Amount: 3000, Timestamp: timePoint, Status: "completed"},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payments []model.Payment

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				payments, err = uow.PaymentRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(payments))

			truncatePayments(ctx, t, db)
		})
	}
}

func TestPaymentRepository_Delete(t *testing.T) {
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
		setup    func(context.Context, database.UnitOfWork) []model.Payment
	}{
		{
			name:     "удаление из пустой таблицы",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Payment {
				return []model.Payment{}
			},
		},
		{
			name:     "удаление одного платежа",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Payment {
				user := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				payments, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{UserTgID: user.TgID, Amount: 1000, Timestamp: timePoint, Status: "pending"},
				})
				require.NoError(t, err)
				return payments
			},
		},
		{
			name:     "удаление части платежей",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Payment {
				user1 := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				user2 := createTestUser(ctx, t, uow, timePoint.Add(1251*time.Microsecond).UnixNano()%1000000)

				payments, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{UserTgID: user1.TgID, Amount: 1000, Timestamp: timePoint, Status: "completed"},
					{UserTgID: user2.TgID, Amount: 2000, Timestamp: timePoint, Status: "pending"},
				})
				require.NoError(t, err)

				// возвращаем только первый для удаления
				return payments[:1]
			},
		},
		{
			name:     "удаление несуществующих платежей (не должно быть ошибки)",
			wantLeft: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Payment {
				user := createTestUser(ctx, t, uow, timePoint.UnixNano()%1000000)
				_, err := uow.PaymentRepo().Upsert(ctx, []model.Payment{
					{UserTgID: user.TgID, Amount: 1000, Timestamp: timePoint, Status: "pending"},
					{UserTgID: user.TgID, Amount: 2000, Timestamp: timePoint, Status: "completed"},
				})
				require.NoError(t, err)

				// возвращаем несуществующие
				return []model.Payment{{UserTgID: 999999}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payments []model.Payment

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				toDelete := tt.setup(ctx, uow)

				err = uow.PaymentRepo().Delete(ctx, toDelete)
				require.NoError(t, err)

				payments, err = uow.PaymentRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantLeft, len(payments))

			truncatePayments(ctx, t, db)
		})
	}
}
