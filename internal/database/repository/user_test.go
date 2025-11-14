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

// helper for truncating table after each case
func truncateUsers(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func TestUserRepository_Upsert(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	uowf := NewUnitOfWorkFactory(db)

	timePoint1 := time.Now()

	tests := []struct {
		name    string
		users   []model.User
		wantErr bool
	}{
		{
			name:    "empty slice",
			users:   []model.User{},
			wantErr: false,
		},
		{
			name: "insert one full user",
			users: []model.User{
				{
					TgID:        time.Now().UnixNano(),
					Username:    ptr("Ivan"),
					FirstTime:   time.Now().UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			wantErr: false,
		},
		{
			name: "insert with nil username",
			users: []model.User{
				{
					TgID:        time.Now().UnixNano(),
					Username:    nil,
					FirstTime:   time.Now().UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			wantErr: false,
		},
		{
			name: "insert multiple users",
			users: []model.User{
				{
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("Ivan"),
					FirstTime:   timePoint1.UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
				{
					TgID:        timePoint1.Add(time.Second).UnixNano(),
					Username:    ptr("Petr"),
					FirstTime:   timePoint1.Add(time.Second).UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			wantErr: false,
		},
		{
			name: "insert multiple users with duplicates",
			users: []model.User{
				{
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("Ivan"),
					FirstTime:   timePoint1.UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
				{
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("Ivan"),
					FirstTime:   timePoint1.Add(time.Second).UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
				{
					TgID:        timePoint1.Add(2 * time.Second).UnixNano(),
					Username:    ptr("Petr"),
					FirstTime:   timePoint1.Add(2 * time.Second).UTC(),
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, err)

			var upserted []model.User

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				upserted, err = uow.UserRepo().Upsert(ctx, tt.users)
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

			require.Len(t, upserted, len(tt.users))

			for _, user := range upserted {
				require.NotEqual(t, user.Id, 0)
			}

			truncateUsers(ctx, t, db)
		})
	}
}

func TestUserRepository_SelectByTgID(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)

	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	timePoint1 := time.Now()

	tests := []struct {
		name     string
		insert   []model.User
		selectID []int64
		want     int
	}{
		{
			name:     "empty tgIDs",
			insert:   nil,
			selectID: nil,
			want:     0,
		},
		{
			name: "single user",
			insert: []model.User{
				{
					Id:          1,
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("single"),
					FirstTime:   timePoint1,
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			selectID: []int64{timePoint1.UnixNano()},
			want:     1,
		},
		{
			name: "multiple users",
			insert: []model.User{
				{
					Id:          1,
					TgID:        timePoint1.Add(time.Second).UnixNano(),
					Username:    ptr("second"),
					FirstTime:   timePoint1.Add(time.Second),
					TotalSub:    2,
					ContainsSub: false,
				},
				{
					Id:          2,
					TgID:        timePoint1.Add(2 * time.Second).UnixNano(),
					Username:    ptr("third"),
					FirstTime:   timePoint1.Add(2 * time.Second),
					TotalSub:    3,
					ContainsSub: true,
				},
			},
			selectID: []int64{timePoint1.Add(time.Second).UnixNano(), timePoint1.Add(2 * time.Second).UnixNano()},
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []model.User

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.insert) > 0 {
					upserted, err := uow.UserRepo().Upsert(ctx, tt.insert)
					require.NoError(t, err)
					require.Len(t, upserted, len(tt.insert))
				}

				users, err = uow.UserRepo().SelectByTgID(ctx, tt.selectID)
				require.NoError(t, err)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(users))

			for i, u := range users {
				assert.Equal(t, u.Id, tt.insert[i].Id)
				assert.Equal(t, u.TotalSub, tt.insert[i].TotalSub)
				assert.Equal(t, u.ContainsSub, tt.insert[i].ContainsSub)
				assert.Equal(t, u.FirstTime.Second(), tt.insert[i].FirstTime.Second())
				assert.Equal(t, u.TgID, tt.insert[i].TgID)
				assert.Equal(t, u.Username, tt.insert[i].Username)
			}

			truncateUsers(ctx, t, db)
		})
	}
}

func TestUserRepository_SelectByUsername(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	timePoint1 := time.Now().UTC()

	tests := []struct {
		name        string
		insert      []model.User
		selectNames []string
		want        int
	}{
		{
			name:        "empty tgIDs",
			insert:      nil,
			selectNames: nil,
			want:        0,
		},
		{
			name: "single user",
			insert: []model.User{
				{
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("single"),
					FirstTime:   timePoint1,
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			selectNames: []string{"single"},
			want:        1,
		},
		{
			name: "multiple users",
			insert: []model.User{
				{
					TgID:        timePoint1.Add(time.Second).UnixNano(),
					Username:    ptr("second"),
					FirstTime:   timePoint1.Add(time.Second),
					TotalSub:    2,
					ContainsSub: false,
				},
				{
					TgID:        timePoint1.Add(2 * time.Second).UnixNano(),
					Username:    ptr("third"),
					FirstTime:   timePoint1.Add(2 * time.Second),
					TotalSub:    3,
					ContainsSub: true,
				},
			},
			selectNames: []string{"second", "third"},
			want:        2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []model.User

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.insert) > 0 {
					var upserted []model.User
					upserted, err = uow.UserRepo().Upsert(ctx, tt.insert)
					if err != nil {
						return errors.Wrap(err, "uow.UserRepo.Upsert")
					}
					require.Len(t, upserted, len(tt.insert))

					for i, u := range upserted {
						tt.insert[i].Id = u.Id
					}
				}

				users, err = uow.UserRepo().SelectByUsername(ctx, tt.selectNames)
				if err != nil {
					return errors.Wrap(err, "uow.UserRepo.SelectByTgID")
				}

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(users))

			for i, u := range users {
				assert.Equal(t, u.Id, tt.insert[i].Id)
				assert.Equal(t, u.TotalSub, tt.insert[i].TotalSub)
				assert.Equal(t, u.ContainsSub, tt.insert[i].ContainsSub)
				assert.Equal(t, u.FirstTime.Second(), tt.insert[i].FirstTime.Second())
				assert.Equal(t, u.TgID, tt.insert[i].TgID)
				assert.Equal(t, u.Username, tt.insert[i].Username)
			}

			truncateUsers(ctx, t, db)
		})
	}
}

func TestUserRepository_SelectAll(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()

	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)

	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	timePoint1 := time.Now().UTC()

	tests := []struct {
		name   string
		insert []model.User
		want   int
	}{
		{
			name:   "empty table",
			insert: nil,
			want:   0,
		},
		{
			name: "one user",
			insert: []model.User{
				{
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("single"),
					FirstTime:   timePoint1,
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			want: 1,
		},
		{
			name: "multiple users",
			insert: []model.User{
				{
					TgID:        timePoint1.Add(time.Second).UnixNano(),
					Username:    ptr("second"),
					FirstTime:   timePoint1.Add(time.Second),
					TotalSub:    2,
					ContainsSub: false,
				},
				{
					TgID:        timePoint1.Add(2 * time.Second).UnixNano(),
					Username:    ptr("third"),
					FirstTime:   timePoint1.Add(2 * time.Second),
					TotalSub:    3,
					ContainsSub: true,
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []model.User

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				if len(tt.insert) > 0 {
					var upserted []model.User

					upserted, err = uow.UserRepo().Upsert(ctx, tt.insert)
					if err != nil {
						return errors.Wrap(err, "uow.UserRepo.Upsert")
					}
					require.Len(t, upserted, len(tt.insert))

					for i, u := range upserted {
						tt.insert[i].Id = u.Id
					}
				}

				users, err = uow.UserRepo().SelectAll(ctx)
				if err != nil {
					return errors.Wrap(err, "uow.UserRepo.SelectAll")
				}

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(users))

			for i, u := range users {
				assert.Equal(t, u.Id, tt.insert[i].Id)
				assert.Equal(t, u.TotalSub, tt.insert[i].TotalSub)
				assert.Equal(t, u.ContainsSub, tt.insert[i].ContainsSub)
				assert.Equal(t, u.FirstTime.Second(), tt.insert[i].FirstTime.Second())
				assert.Equal(t, u.TgID, tt.insert[i].TgID)
				assert.Equal(t, u.Username, tt.insert[i].Username)
			}

			truncateUsers(ctx, t, db)
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	ctx := t.Context()
	logger := logrus.New()
	db, container, err := testcontainer.CreatePostgresStorage(ctx, logger)
	require.NoError(t, err)

	defer func() { _ = container.Terminate(ctx) }()

	uowf := NewUnitOfWorkFactory(db)

	timePoint1 := time.Now().UTC()

	tests := []struct {
		name     string
		toInsert []model.User
		toDelete []model.User
		wantLeft int
	}{
		{
			name: "delete none (empty)",
			toInsert: []model.User{
				{
					TgID:        timePoint1.UnixNano(),
					Username:    ptr("single"),
					FirstTime:   timePoint1,
					TotalSub:    1,
					ContainsSub: true,
				},
			},
			toDelete: nil,
			wantLeft: 1,
		},
		{
			name: "delete one",
			toInsert: []model.User{
				{
					TgID:        timePoint1.Add(time.Second).UnixNano(),
					Username:    ptr("second"),
					FirstTime:   timePoint1.Add(time.Second),
					TotalSub:    2,
					ContainsSub: false,
				},
			},
			toDelete: []model.User{
				{TgID: timePoint1.Add(time.Second).UnixNano()},
			},
			wantLeft: 0,
		},
		{
			name: "delete multiple",
			toInsert: []model.User{
				{
					TgID:        timePoint1.Add(2 * time.Second).UnixNano(),
					Username:    ptr("third"),
					FirstTime:   timePoint1.Add(2 * time.Second),
					TotalSub:    3,
					ContainsSub: true,
				},
				{
					TgID:        timePoint1.Add(3 * time.Second).UnixNano(),
					Username:    ptr("fourth"),
					FirstTime:   timePoint1.Add(3 * time.Second),
					TotalSub:    4,
					ContainsSub: false,
				},
			},
			toDelete: []model.User{
				{TgID: timePoint1.Add(2 * time.Second).UnixNano()},
				{TgID: timePoint1.Add(3 * time.Second).UnixNano()},
			},
			wantLeft: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []model.User

			err = uowf.New(ctx, pgx.ReadCommitted, func(uow database.UnitOfWork) error {
				expected := make(map[int64]model.User)

				if len(tt.toInsert) > 0 {
					var upserted []model.User
					upserted, err = uow.UserRepo().Upsert(ctx, tt.toInsert)
					if err != nil {
						return errors.Wrap(err, "uow.UserRepo.Upsert")
					}
					require.Len(t, upserted, len(tt.toInsert))

					for i, u := range upserted {
						tt.toInsert[i].Id = u.Id
						expected[u.TgID] = u
					}
				}

				if len(tt.toDelete) > 0 {
					err = uow.UserRepo().Delete(ctx, tt.toDelete)
					if err != nil {
						return errors.Wrap(err, "uow.UserRepo.Delete")
					}

					for _, del := range tt.toDelete {
						delete(expected, del.TgID)
					}
				}

				users, err = uow.UserRepo().SelectAll(ctx)
				if err != nil {
					return errors.Wrap(err, "uow.UserRepo.SelectAll")
				}

				expectedSlice := make([]model.User, 0, len(expected))
				for _, u := range expected {
					expectedSlice = append(expectedSlice, u)
				}

				assert.ElementsMatch(t, expectedSlice, users)

				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantLeft, len(users))

			truncateUsers(ctx, t, db)
		})
	}
}

func ptr(s string) *string { return &s }
