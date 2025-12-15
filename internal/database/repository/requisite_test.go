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
func truncateRequisites(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE requisites RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func TestRequisiteRepository_Upsert(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	uowf := NewUnitOfWorkFactory(db)

	tests := []struct {
		name       string
		requisites []model.Requisite
		wantErr    bool
	}{
		{
			name:       "пустой слайс",
			requisites: []model.Requisite{},
			wantErr:    false,
		},
		{
			name: "вставка одного реквизита",
			requisites: []model.Requisite{
				{
					Name:    "Bank Transfer",
					Link:    "https://bank.com/transfer",
					Content: "Account: 1234567890",
					Photo:   []byte("qr_code_data"),
				},
			},
			wantErr: false,
		},
		{
			name: "вставка нескольких реквизитов",
			requisites: []model.Requisite{
				{
					Name:    "Card Payment",
					Link:    "https://payment.com/card",
					Content: "Card: **** **** **** 1234",
					Photo:   nil,
				},
				{
					Name:    "Crypto",
					Link:    "https://crypto.com/wallet",
					Content: "BTC Address: 1A2B3C...",
					Photo:   []byte("wallet_qr"),
				},
			},
			wantErr: false,
		},
		{
			name: "вставка реквизита без фото",
			requisites: []model.Requisite{
				{
					Name:    "Cash",
					Link:    "https://cash.com/pay",
					Content: "Meet at location",
					Photo:   nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var upserted []model.Requisite

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				upserted, err = uow.RequisiteRepo().Upsert(ctx, tt.requisites)
				if err != nil {
					return errors.Wrap(err, "uow.RequisiteRepo.Upsert")
				}

				return nil
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, upserted, len(tt.requisites))

			// проверяем все поля возвращенных реквизитов
			for i, requisite := range upserted {
				require.NotEqual(t, int64(0), requisite.Id)
				require.Equal(t, tt.requisites[i].Name, requisite.Name)
				require.Equal(t, tt.requisites[i].Link, requisite.Link)
				require.Equal(t, tt.requisites[i].Content, requisite.Content)
				assert.Equal(t, tt.requisites[i].Photo, requisite.Photo)
			}

			truncateRequisites(ctx, t, db)
		})
	}
}

func TestRequisiteRepository_Upsert_Update(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	link := "https://update-test.com/payment"

	// первая вставка
	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
			{Name: "Old Name", Link: link, Content: "Old content", Photo: []byte("old")},
		})
		return err
	})
	require.NoError(t, err)

	// обновление (upsert с тем же link)
	var updated []model.Requisite
	err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
		updated, err = uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
			{Name: "New Name", Link: link, Content: "New content", Photo: []byte("new")},
		})
		return err
	})
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, "New Name", updated[0].Name)
	require.Equal(t, "New content", updated[0].Content)

	truncateRequisites(ctx, t, db)
}

func TestRequisiteRepository_SelectByName(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	tests := []struct {
		name  string
		names []string
		want  int
		setup func(context.Context, database.UnitOfWork) []string // возвращает names
	}{
		{
			name:  "пустой слайс names",
			names: nil,
			want:  0,
			setup: nil,
		},
		{
			name:  "несуществующие names",
			names: []string{"Nonexistent"},
			want:  0,
			setup: nil,
		},
		{
			name: "один реквизит",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []string {
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "SingleReq", Link: "https://single.com/" + time.Now().String(), Content: "Content", Photo: nil},
				})
				require.NoError(t, err)
				return []string{"SingleReq"}
			},
		},
		{
			name: "несколько реквизитов",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []string {
				timestamp := time.Now().UnixNano()
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "FirstReq", Link: "https://first.com/" + string(rune(timestamp)), Content: "Content 1", Photo: nil},
					{Name: "SecondReq", Link: "https://second.com/" + string(rune(timestamp+1)), Content: "Content 2", Photo: nil},
				})
				require.NoError(t, err)
				return []string{"FirstReq", "SecondReq"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requisites []model.Requisite

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.names = tt.setup(ctx, uow)
				}

				requisites, err = uow.RequisiteRepo().SelectByName(ctx, tt.names)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(requisites))

			// проверяем заполненность полей
			for _, r := range requisites {
				assert.NotEqual(t, int64(0), r.Id)
				assert.NotEmpty(t, r.Name)
				assert.NotEmpty(t, r.Link)
				assert.NotEmpty(t, r.Content)
			}

			truncateRequisites(ctx, t, db)
		})
	}
}

func TestRequisiteRepository_SelectByLink(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	tests := []struct {
		name  string
		links []string
		want  int
		setup func(context.Context, database.UnitOfWork) []string // возвращает links
	}{
		{
			name:  "пустой слайс links",
			links: nil,
			want:  0,
			setup: nil,
		},
		{
			name:  "несуществующие links",
			links: []string{"https://nonexistent.com/link"},
			want:  0,
			setup: nil,
		},
		{
			name: "один реквизит по link",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []string {
				link := "https://link-test.com/pay/" + time.Now().String()
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "LinkTest", Link: link, Content: "Test content", Photo: nil},
				})
				require.NoError(t, err)
				return []string{link}
			},
		},
		{
			name: "несколько реквизитов по links",
			want: 2,
			setup: func(ctx context.Context, uow database.UnitOfWork) []string {
				timestamp := time.Now().UnixNano()
				link1 := "https://link1.com/pay/" + string(rune(timestamp))
				link2 := "https://link2.com/pay/" + string(rune(timestamp+1))
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "Link1", Link: link1, Content: "Content 1", Photo: nil},
					{Name: "Link2", Link: link2, Content: "Content 2", Photo: nil},
				})
				require.NoError(t, err)
				return []string{link1, link2}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requisites []model.Requisite

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.links = tt.setup(ctx, uow)
				}

				requisites, err = uow.RequisiteRepo().SelectByLink(ctx, tt.links)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(requisites))

			truncateRequisites(ctx, t, db)
		})
	}
}

func TestRequisiteRepository_SelectAll(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

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
			name: "один реквизит",
			want: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "AllOne", Link: "https://all-one.com/" + time.Now().String(), Content: "Content", Photo: nil},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "несколько реквизитов",
			want: 3,
			setup: func(ctx context.Context, uow database.UnitOfWork) {
				timestamp := time.Now().UnixNano()
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "All1", Link: "https://all1.com/" + string(rune(timestamp)), Content: "Content 1", Photo: nil},
					{Name: "All2", Link: "https://all2.com/" + string(rune(timestamp+1)), Content: "Content 2", Photo: nil},
					{Name: "All3", Link: "https://all3.com/" + string(rune(timestamp+2)), Content: "Content 3", Photo: []byte("data")},
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requisites []model.Requisite

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if tt.setup != nil {
					tt.setup(ctx, uow)
				}

				requisites, err = uow.RequisiteRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(requisites))

			truncateRequisites(ctx, t, db)
		})
	}
}

func TestRequisiteRepository_Delete(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	tests := []struct {
		name     string
		wantLeft int
		setup    func(context.Context, database.UnitOfWork) []model.Requisite
	}{
		{
			name:     "удаление из пустой таблицы",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Requisite {
				return []model.Requisite{}
			},
		},
		{
			name:     "удаление одного реквизита",
			wantLeft: 0,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Requisite {
				requisites, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "DelOne", Link: "https://del-one.com/" + time.Now().String(), Content: "To delete", Photo: nil},
				})
				require.NoError(t, err)
				return requisites
			},
		},
		{
			name:     "удаление части реквизитов",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Requisite {
				timestamp := time.Now().UnixNano()
				requisites, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "DelFirst", Link: "https://del-first.com/" + string(rune(timestamp)), Content: "To delete", Photo: nil},
					{Name: "DelKeep", Link: "https://del-keep.com/" + string(rune(timestamp+1)), Content: "To keep", Photo: nil},
				})
				require.NoError(t, err)
				// удаляем только первый
				return requisites[:1]
			},
		},
		{
			name:     "удаление несуществующих реквизитов (не должно быть ошибки)",
			wantLeft: 1,
			setup: func(ctx context.Context, uow database.UnitOfWork) []model.Requisite {
				_, err := uow.RequisiteRepo().Upsert(ctx, []model.Requisite{
					{Name: "KeepMe", Link: "https://keep-me.com/" + time.Now().String(), Content: "Keep", Photo: nil},
				})
				require.NoError(t, err)
				// возвращаем несуществующий реквизит
				return []model.Requisite{{Link: "https://nonexistent.com/link"}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requisites []model.Requisite

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				toDelete := tt.setup(ctx, uow)

				err = uow.RequisiteRepo().Delete(ctx, toDelete)
				require.NoError(t, err)

				requisites, err = uow.RequisiteRepo().SelectAll(ctx)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantLeft, len(requisites))

			truncateRequisites(ctx, t, db)
		})
	}
}
