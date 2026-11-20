// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"errors"
	"fmt"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authRepository "github.com/lin-snow/ech0/internal/repository/auth"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	"gorm.io/driver/sqlite"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	embeddingRepository "github.com/lin-snow/ech0/internal/repository/embedding"
	oauthserverRepository "github.com/lin-snow/ech0/internal/repository/oauthserver"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Set ECH0_TEST_SQLD to the sqld executable. The test always owns its
// temporary database and only connects to that server over loopback.
func TestLibSQLPersistence(t *testing.T) {
	binary := os.Getenv("ECH0_TEST_SQLD")
	if binary == "" {
		t.Skip("ECH0_TEST_SQLD is unset")
	}
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	address := ln.Addr().String()
	require.NoError(t, ln.Close())
	dir := t.TempDir()
	cmd := exec.CommandContext(t.Context(), binary, "--db-path", dir+"/db", "--http-listen-addr", address)
	cmd.Dir = dir
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	require.Eventually(t, func() bool {
		c, err := net.DialTimeout("tcp", address, time.Millisecond*100)
		if err != nil {
			return false
		}
		return c.Close() == nil
	}, 10*time.Second, 50*time.Millisecond)
	cfg := config.DatabaseConfig{URL: "http://" + address}
	open := func() *gorm.DB {
		db, err := openLibSQL(cfg, logger.Silent)
		require.NoError(t, err)
		pool, err := db.DB()
		require.NoError(t, err)
		t.Cleanup(func() { _ = pool.Close() })
		return db
	}
	a, b := open(), open()
	ctx := t.Context()
	results := make(chan error, 2)
	for _, db := range []*gorm.DB{a, b} {
		go func() { results <- prepareRemoteSchema(db.WithContext(ctx)) }()
	}
	for range 2 {
		require.NoError(t, <-results)
	}
	require.NoError(t, checkRemoteSchema(b))
	visitsA := visitorRepository.NewVisitorRepository(func() *gorm.DB { return a })
	visitsB := visitorRepository.NewVisitorRepository(func() *gorm.DB { return b })
	beforeMidnight := time.Date(2026, 1, 1, 23, 58, 0, 0, time.UTC)
	for _, repo := range []*visitorRepository.VisitorRepository{visitsA, visitsB} {
		go func() { results <- repo.RecordVisit(ctx, "same", beforeMidnight, 5*time.Minute) }()
	}
	for range 2 {
		require.NoError(t, <-results)
	}
	require.NoError(t, visitsB.RecordVisit(ctx, "same", beforeMidnight.Add(3*time.Minute), 5*time.Minute))
	require.NoError(t, visitsA.RecordVisit(ctx, "same", beforeMidnight.Add(6*time.Minute), 5*time.Minute))
	days, err := visitsB.GetRecentDays(ctx, 7)
	require.NoError(t, err)
	require.Len(t, days, 2)
	for _, day := range days {
		require.EqualValues(t, 1, day.PV)
		require.EqualValues(t, 1, day.UV)
	}
	revokerA := authRepository.NewAuthRepository(func() *gorm.DB { return a })
	revokerB := authRepository.NewAuthRepository(func() *gorm.DB { return b })
	session := authModel.PasskeySession{Origin: "https://example.test"}
	require.NoError(t, revokerA.StorePasskeySession("register:test-nonce", session, time.Minute))
	restored, err := revokerB.ConsumePasskeySession("register:test-nonce")
	require.NoError(t, err)
	require.Equal(t, session, restored)
	_, err = revokerA.ConsumePasskeySession("register:test-nonce")
	require.Error(t, err)
	require.False(t, revokerB.IsTokenRevoked("session"))
	require.NoError(t, revokerA.RevokeToken("session", time.Minute))
	require.True(t, revokerB.IsTokenRevoked("session"))
	accepted := make(chan bool, 2)
	for _, db := range []*gorm.DB{a, b} {
		repo := oauthserverRepository.NewClientRepository(func() *gorm.DB { return db })
		go func() {
			inserted, err := repo.ConsumeCode(ctx, "single-use", 10, 20)
			accepted <- inserted
			results <- err
		}()
	}
	require.NotEqual(t, <-accepted, <-accepted)
	for range 2 {
		require.NoError(t, <-results)
	}
	abort := errors.New("rollback")
	require.ErrorIs(t, a.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&echoModel.Echo{ID: "rollback", Content: "sample", UserID: "owner"}).Error; err != nil {
			return err
		}
		return abort
	}), abort)
	var count int64
	require.NoError(t, b.Model(&echoModel.Echo{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, a.Create(&echoModel.Echo{ID: "entry", Content: "sample", UserID: "owner"}).Error)
	var indexes int
	require.NoError(t, b.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_echos_private_created_id'").Scan(&indexes).Error)
	require.Equal(t, 1, indexes)
	for n, db := range []*gorm.DB{a, b} {
		go func() {
			results <- db.Create(&echoModel.Echo{ID: fmt.Sprintf("extra-%d", n), Content: "keyword", UserID: "owner"}).Error
		}()
	}
	for range 2 {
		require.NoError(t, <-results)
	}
	require.NoError(t, a.Model(&echoModel.Echo{}).Where("content LIKE ?", "%word%").Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, b.Model(&echoModel.Echo{}).Where("id = ?", "extra-0").Update("content", "edited").Error)
	require.NoError(t, a.Where("content = ?", "edited").Delete(&echoModel.Echo{}).Error)
	require.NoError(t, b.Model(&echoModel.Echo{}).Where("id = ?", "extra-0").Count(&count).Error)
	require.Zero(t, count)
	writer := embeddingRepository.NewEmbeddingRepository(func() *gorm.DB { return a })
	reader := embeddingRepository.NewEmbeddingRepository(func() *gorm.DB { return b })
	require.NoError(t, writer.EnsureVecTable(ctx, 2))
	meta := &embeddingModel.EchoEmbedding{EchoID: "entry", Dim: 2}
	require.NoError(t, writer.Upsert(ctx, meta, []float32{3, 4}))
	require.Error(t, writer.Upsert(ctx, meta, []float32{3}))
	hits, err := reader.Search(ctx, []float32{0, 0}, 1, "owner")
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, 5.0, hits[0].Distance)
	hits, err = reader.Search(ctx, []float32{0, 0}, 1, "another")
	require.NoError(t, err)
	require.Empty(t, hits)
	backup := filepath.Join(t.TempDir(), "ech0.db")
	require.NoError(t, writeRemoteSnapshot(ctx, b, backup))
	local, err := gorm.Open(sqlite.Open(backup), buildGormConfig(logger.Silent))
	require.NoError(t, err)
	localPool, err := local.DB()
	require.NoError(t, err)
	defer func() { require.NoError(t, localPool.Close()) }()
	native := embeddingRepository.NewEmbeddingRepository(func() *gorm.DB { return local })
	restoredHits, err := native.Search(ctx, []float32{0, 0}, 1, "owner")
	require.NoError(t, err)
	require.Len(t, restoredHits, 1)
	require.Equal(t, 5.0, restoredHits[0].Distance)
	restoredDB := filepath.Join(t.TempDir(), "restored.db")
	require.NoError(t, ConvertSQLite(ctx, backup, restoredDB))
	require.NoError(t, writer.Delete(ctx, "entry"))
	hits, err = reader.Search(ctx, []float32{0, 0}, 1, "")
	require.NoError(t, err)
	require.Empty(t, hits)
}

func TestLibSQLConfiguration(t *testing.T) {
	for _, endpoint := range []string{"file:test.db", "http://example.com", "libsql://secret@example.com", "https://example.com?authToken=secret"} {
		_, err := openLibSQL(config.DatabaseConfig{URL: endpoint}, logger.Silent)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "secret")
	}
	_, err := openLibSQL(config.DatabaseConfig{URL: "libsql://example.com"}, logger.Silent)
	require.ErrorContains(t, err, "TURSO_AUTH_TOKEN")
}
