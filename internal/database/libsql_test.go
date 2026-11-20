// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	embeddingRepository "github.com/lin-snow/ech0/internal/repository/embedding"
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
		c.Close()
		return true
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
