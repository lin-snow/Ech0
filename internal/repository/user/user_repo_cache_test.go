// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newUserRepo(t *testing.T) (*UserRepository, *gorm.DB, cache.ICache[string, any]) {
	t.Helper()
	db := helpers.NewTestDB(t)
	c := helpers.NewTestCache()
	return NewUserRepository(func() *gorm.DB { return db }, c), db, c
}

func cacheGetUser(t *testing.T, c cache.ICache[string, any], key string) (any, bool) {
	t.Helper()
	v, ok, err := c.Get(key)
	require.NoError(t, err)
	return v, ok
}

func seedUser(t *testing.T, db *gorm.DB, u userModel.User) userModel.User {
	t.Helper()
	require.NoError(t, db.Create(&u).Error)
	var got userModel.User
	require.NoError(t, db.Where("id = ?", u.ID).First(&got).Error)
	return got
}

func TestUserRepository_UpdateUser_CacheCoordination(t *testing.T) {
	t.Run("rename deletes the old username key and sets the new one", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		existing := seedUser(t, db, userModel.User{ID: "u1", Username: "old"})
		c.Set(GetUsernameKey("old"), existing, 1)

		updated := existing
		updated.Username = "new"
		require.NoError(t, repo.UpdateUser(context.Background(), &updated))

		_, ok := cacheGetUser(t, c, GetUsernameKey("old"))
		assert.False(t, ok, "stale username key must be deleted on rename")

		gotNew, ok := cacheGetUser(t, c, GetUsernameKey("new"))
		require.True(t, ok)
		assert.Equal(t, updated, gotNew)

		gotID, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		require.True(t, ok)
		assert.Equal(t, updated, gotID)

		var dbUser userModel.User
		require.NoError(t, db.Where("id = ?", "u1").First(&dbUser).Error)
		assert.Equal(t, "new", dbUser.Username)
	})

	t.Run("admin downgrade deletes the admin key", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		existing := seedUser(t, db, userModel.User{ID: "u1", Username: "a", IsAdmin: true})
		c.Set(GetAdminKey("u1"), existing, 1)

		updated := existing
		updated.IsAdmin = false
		require.NoError(t, repo.UpdateUser(context.Background(), &updated))

		_, ok := cacheGetUser(t, c, GetAdminKey("u1"))
		assert.False(t, ok, "admin key must be deleted when demoting from admin")

		gotID, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		require.True(t, ok)
		assert.Equal(t, updated, gotID)
	})

	t.Run("owner downgrade deletes the owner key", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		existing := seedUser(t, db, userModel.User{ID: "u1", Username: "o", IsAdmin: true, IsOwner: true})
		c.Set(GetOwnerKey(), existing, 1)

		updated := existing
		updated.IsOwner = false
		require.NoError(t, repo.UpdateUser(context.Background(), &updated))

		_, ok := cacheGetUser(t, c, GetOwnerKey())
		assert.False(t, ok, "owner key must be deleted when demoting from owner")
	})

	t.Run("promotion to admin and owner sets id/username/admin/owner keys", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		existing := seedUser(t, db, userModel.User{ID: "u1", Username: "x"})

		updated := existing
		updated.IsAdmin = true
		updated.IsOwner = true
		require.NoError(t, repo.UpdateUser(context.Background(), &updated))

		gotID, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		require.True(t, ok)
		assert.Equal(t, updated, gotID)

		gotName, ok := cacheGetUser(t, c, GetUsernameKey("x"))
		require.True(t, ok)
		assert.Equal(t, updated, gotName)

		gotAdmin, ok := cacheGetUser(t, c, GetAdminKey("u1"))
		require.True(t, ok)
		assert.Equal(t, updated, gotAdmin)

		gotOwner, ok := cacheGetUser(t, c, GetOwnerKey())
		require.True(t, ok)
		assert.Equal(t, updated, gotOwner)
	})

	t.Run("updating a missing user errors and writes no cache", func(t *testing.T) {
		repo, _, c := newUserRepo(t)

		missing := userModel.User{ID: "ghost", Username: "ghost"}
		err := repo.UpdateUser(context.Background(), &missing)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		_, ok := cacheGetUser(t, c, GetUserIDKey("ghost"))
		assert.False(t, ok, "failed update must not populate the cache")
	})
}

func TestUserRepository_CreateUser_KeyFanout(t *testing.T) {
	t.Run("non-owner sets only id and username keys", func(t *testing.T) {
		repo, db, c := newUserRepo(t)

		user := userModel.User{ID: "u1", Username: "alice", IsAdmin: true}
		require.NoError(t, repo.CreateUser(context.Background(), &user))

		gotID, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		require.True(t, ok)
		assert.Equal(t, user, gotID)

		gotName, ok := cacheGetUser(t, c, GetUsernameKey("alice"))
		require.True(t, ok)
		assert.Equal(t, user, gotName)

		_, ok = cacheGetUser(t, c, GetOwnerKey())
		assert.False(t, ok, "create must not set owner key for a non-owner")
		_, ok = cacheGetUser(t, c, GetAdminKey("u1"))
		assert.False(t, ok, "create does not populate the admin key")

		var n int64
		require.NoError(t, db.Model(&userModel.User{}).Where("id = ?", "u1").Count(&n).Error)
		assert.Equal(t, int64(1), n)
	})

	t.Run("owner also sets the owner key", func(t *testing.T) {
		repo, _, c := newUserRepo(t)

		user := userModel.User{ID: "u1", Username: "boss", IsAdmin: true, IsOwner: true}
		require.NoError(t, repo.CreateUser(context.Background(), &user))

		gotOwner, ok := cacheGetUser(t, c, GetOwnerKey())
		require.True(t, ok)
		assert.Equal(t, user, gotOwner)

		_, ok = cacheGetUser(t, c, GetUserIDKey("u1"))
		assert.True(t, ok)
		_, ok = cacheGetUser(t, c, GetUsernameKey("boss"))
		assert.True(t, ok)
	})
}

func TestUserRepository_DeleteUser_KeyFanout(t *testing.T) {
	t.Run("admin owner deletes id/username/admin/owner keys", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		u := seedUser(t, db, userModel.User{ID: "u1", Username: "boss", IsAdmin: true, IsOwner: true})
		c.Set(GetUserIDKey("u1"), u, 1)
		c.Set(GetUsernameKey("boss"), u, 1)
		c.Set(GetAdminKey("u1"), u, 1)
		c.Set(GetOwnerKey(), u, 1)

		require.NoError(t, repo.DeleteUser(context.Background(), "u1"))

		for _, key := range []string{GetUserIDKey("u1"), GetUsernameKey("boss"), GetAdminKey("u1"), GetOwnerKey()} {
			_, ok := cacheGetUser(t, c, key)
			assert.Falsef(t, ok, "key %q must be invalidated", key)
		}

		var n int64
		require.NoError(t, db.Model(&userModel.User{}).Where("id = ?", "u1").Count(&n).Error)
		assert.Equal(t, int64(0), n)
	})

	t.Run("regular user deletes only id/username keys and leaves admin/owner keys", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		u := seedUser(t, db, userModel.User{ID: "u1", Username: "bob"})
		c.Set(GetUserIDKey("u1"), u, 1)
		c.Set(GetUsernameKey("bob"), u, 1)
		c.Set(GetAdminKey("u1"), u, 1)
		c.Set(GetOwnerKey(), u, 1)

		require.NoError(t, repo.DeleteUser(context.Background(), "u1"))

		_, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		assert.False(t, ok)
		_, ok = cacheGetUser(t, c, GetUsernameKey("bob"))
		assert.False(t, ok)

		_, ok = cacheGetUser(t, c, GetAdminKey("u1"))
		assert.True(t, ok, "admin key must survive deleting a non-admin user")
		_, ok = cacheGetUser(t, c, GetOwnerKey())
		assert.True(t, ok, "owner key must survive deleting a non-owner user")
	})

	t.Run("deleting a missing user errors", func(t *testing.T) {
		repo, _, _ := newUserRepo(t)

		err := repo.DeleteUser(context.Background(), "ghost")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestUserRepository_MarkInitialized_Idempotent(t *testing.T) {
	repo, db, _ := newUserRepo(t)
	ctx := context.Background()

	got, err := repo.IsInitialized(ctx)
	require.NoError(t, err)
	assert.False(t, got, "fresh db should not be initialized")

	require.NoError(t, repo.MarkInitialized(ctx))

	var kv commonModel.KeyValue
	require.NoError(t, db.Where("key = ?", commonModel.InstallInitializedKey).First(&kv).Error)
	assert.Equal(t, "true", kv.Value)

	got, err = repo.IsInitialized(ctx)
	require.NoError(t, err)
	assert.True(t, got)

	require.NoError(t, repo.MarkInitialized(ctx))

	var n int64
	require.NoError(t, db.Model(&commonModel.KeyValue{}).
		Where("key = ?", commonModel.InstallInitializedKey).Count(&n).Error)
	assert.Equal(t, int64(1), n, "MarkInitialized must be idempotent (no duplicate row)")
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	t.Run("empty returns nil without error", func(t *testing.T) {
		repo, _, _ := newUserRepo(t)
		got, err := repo.GetAllUsers(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("returns every inserted user", func(t *testing.T) {
		repo, db, _ := newUserRepo(t)
		seedUser(t, db, userModel.User{ID: "u1", Username: "a"})
		seedUser(t, db, userModel.User{ID: "u2", Username: "b"})

		got, err := repo.GetAllUsers(context.Background())
		require.NoError(t, err)
		require.Len(t, got, 2)
	})
}

func TestUserCacheKeyBuilders(t *testing.T) {
	assert.Equal(t, "id:u1", GetUserIDKey("u1"))
	assert.Equal(t, "username:alice", GetUsernameKey("alice"))
	assert.Equal(t, "admin:u1", GetAdminKey("u1"))
	assert.Equal(t, OwnerKey, GetOwnerKey())
	assert.Equal(t, "passkey:reg:nonce1", GetPasskeyRegisterSessionKey("nonce1"))
	assert.Equal(t, "passkey:login:nonce2", GetPasskeyLoginSessionKey("nonce2"))
}

func TestUserRepository_GetUserByID_ReadThrough(t *testing.T) {
	t.Run("cache hit returns cached without db", func(t *testing.T) {
		repo, _, c := newUserRepo(t)
		sentinel := userModel.User{ID: "u1", Username: "cached"}
		c.Set(GetUserIDKey("u1"), sentinel, 1)

		got, err := repo.GetUserByID(context.Background(), "u1")
		require.NoError(t, err)
		assert.Equal(t, sentinel, got)
	})

	t.Run("cache miss reads db and backfills", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		want := seedUser(t, db, userModel.User{ID: "u1", Username: "dbuser"})

		got, err := repo.GetUserByID(context.Background(), "u1")
		require.NoError(t, err)
		assert.Equal(t, want, got)

		cached, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		require.True(t, ok)
		assert.Equal(t, want, cached)
	})

	t.Run("transaction context bypasses cache", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		want := seedUser(t, db, userModel.User{ID: "u1", Username: "dbuser"})
		stale := userModel.User{ID: "u1", Username: "stale"}
		c.Set(GetUserIDKey("u1"), stale, 1)

		ctxTx := context.WithValue(context.Background(), transaction.TxKey, db)
		got, err := repo.GetUserByID(ctxTx, "u1")
		require.NoError(t, err)
		assert.Equal(t, want, got)

		cached, ok := cacheGetUser(t, c, GetUserIDKey("u1"))
		require.True(t, ok)
		assert.Equal(t, stale, cached, "tx read must not mutate the cache")
	})

	t.Run("missing id returns record-not-found", func(t *testing.T) {
		repo, _, _ := newUserRepo(t)
		_, err := repo.GetUserByID(context.Background(), "ghost")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestUserRepository_GetUserByUsername_ReadThrough(t *testing.T) {
	t.Run("cache hit returns cached without db", func(t *testing.T) {
		repo, _, c := newUserRepo(t)
		sentinel := userModel.User{ID: "u1", Username: "alice"}
		c.Set(GetUsernameKey("alice"), sentinel, 1)

		got, err := repo.GetUserByUsername(context.Background(), "alice")
		require.NoError(t, err)
		assert.Equal(t, sentinel, got)
	})

	t.Run("cache miss reads db and backfills", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		want := seedUser(t, db, userModel.User{ID: "u1", Username: "alice"})

		got, err := repo.GetUserByUsername(context.Background(), "alice")
		require.NoError(t, err)
		assert.Equal(t, want, got)

		cached, ok := cacheGetUser(t, c, GetUsernameKey("alice"))
		require.True(t, ok)
		assert.Equal(t, want, cached)
	})

	t.Run("transaction context bypasses cache", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		want := seedUser(t, db, userModel.User{ID: "u1", Username: "alice"})
		stale := userModel.User{ID: "u9", Username: "alice"}
		c.Set(GetUsernameKey("alice"), stale, 1)

		ctxTx := context.WithValue(context.Background(), transaction.TxKey, db)
		got, err := repo.GetUserByUsername(ctxTx, "alice")
		require.NoError(t, err)
		assert.Equal(t, want, got)

		cached, ok := cacheGetUser(t, c, GetUsernameKey("alice"))
		require.True(t, ok)
		assert.Equal(t, stale, cached)
	})

	t.Run("missing username returns record-not-found", func(t *testing.T) {
		repo, _, _ := newUserRepo(t)
		_, err := repo.GetUserByUsername(context.Background(), "ghost")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestUserRepository_GetOwner_ReadThrough(t *testing.T) {
	t.Run("cache hit returns cached without db", func(t *testing.T) {
		repo, _, c := newUserRepo(t)
		sentinel := userModel.User{ID: "u1", Username: "owner", IsOwner: true}
		c.Set(GetOwnerKey(), sentinel, 1)

		got, err := repo.GetOwner(context.Background())
		require.NoError(t, err)
		assert.Equal(t, sentinel, got)
	})

	t.Run("cache miss reads db and backfills", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		seedUser(t, db, userModel.User{ID: "u2", Username: "plain"})
		want := seedUser(t, db, userModel.User{ID: "u1", Username: "owner", IsAdmin: true, IsOwner: true})

		got, err := repo.GetOwner(context.Background())
		require.NoError(t, err)
		assert.Equal(t, want, got)

		cached, ok := cacheGetUser(t, c, GetOwnerKey())
		require.True(t, ok)
		assert.Equal(t, want, cached)
	})

	t.Run("transaction context bypasses cache", func(t *testing.T) {
		repo, db, c := newUserRepo(t)
		want := seedUser(t, db, userModel.User{ID: "u1", Username: "owner", IsAdmin: true, IsOwner: true})
		stale := userModel.User{ID: "u9", Username: "stale-owner", IsOwner: true}
		c.Set(GetOwnerKey(), stale, 1)

		ctxTx := context.WithValue(context.Background(), transaction.TxKey, db)
		got, err := repo.GetOwner(ctxTx)
		require.NoError(t, err)
		assert.Equal(t, want, got)

		cached, ok := cacheGetUser(t, c, GetOwnerKey())
		require.True(t, ok)
		assert.Equal(t, stale, cached)
	})

	t.Run("no owner returns record-not-found", func(t *testing.T) {
		repo, _, _ := newUserRepo(t)
		_, err := repo.GetOwner(context.Background())
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}
