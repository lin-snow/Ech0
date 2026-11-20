// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const remoteSchemaVersion = 1

func openLibSQL(cfg config.DatabaseConfig, level logger.LogLevel) (*gorm.DB, error) {
	endpoint, err := url.Parse(cfg.URL)
	if err != nil || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, errors.New("TURSO_DATABASE_URL must be a secure libSQL endpoint without credentials or query parameters")
	}
	local := endpoint.Scheme == "http" && (endpoint.Hostname() == "127.0.0.1" || endpoint.Hostname() == "::1")
	if !local && endpoint.Scheme != "https" && endpoint.Scheme != "libsql" {
		return nil, errors.New("libSQL requires TLS outside loopback")
	}
	var opts []libsql.Option
	if cfg.AuthToken != "" {
		opts = append(opts, libsql.WithAuthToken(cfg.AuthToken))
	} else if !local {
		return nil, errors.New("TURSO_AUTH_TOKEN is required")
	}
	connector, err := libsql.NewConnector(endpoint.String(), opts...)
	if err != nil {
		return nil, errors.New("invalid libSQL connection configuration")
	}
	pool := sql.OpenDB(connector)
	pool.SetMaxOpenConns(4)
	pool.SetMaxIdleConns(1)
	pool.SetConnMaxIdleTime(time.Minute)
	db, err := gorm.Open(sqlite.New(sqlite.Config{DriverName: "libsql", Conn: pool}), buildGormConfig(level))
	if err != nil {
		_ = pool.Close()
		return nil, err
	}
	return db, nil
}

func checkRemoteSchema(db *gorm.DB) error {
	var version int
	if err := db.Raw("SELECT version FROM ech0_schema WHERE id = 1").Row().Scan(&version); err != nil {
		return err
	}
	if version != remoteSchemaVersion {
		return errors.New("unsupported Ech0 remote schema version")
	}
	return nil
}

// PrepareRemoteSchema provisions a new database. Older Ech0 databases are
// migrated locally before their rows are copied to the provisioned target.
func PrepareRemoteSchema(ctx context.Context, cfg config.DatabaseConfig) error {
	db, err := openLibSQL(cfg, logger.Silent)
	if err != nil {
		return err
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	return prepareRemoteSchema(db.WithContext(ctx))
}

func prepareRemoteSchema(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("CREATE TABLE IF NOT EXISTS ech0_schema (id INTEGER PRIMARY KEY CHECK(id = 1), version INTEGER NOT NULL)").Error; err != nil {
			return err
		}
		var marked int
		if err := tx.Raw("SELECT COUNT(*) FROM ech0_schema").Scan(&marked).Error; err != nil {
			return err
		}
		if marked != 0 {
			return checkRemoteSchema(tx)
		}
		var existing int
		if err := tx.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name <> 'ech0_schema' AND name NOT GLOB 'sqlite_*' AND name NOT GLOB '_litestream_*' AND name NOT GLOB 'libsql_*'").Scan(&existing).Error; err != nil {
			return err
		}
		if existing != 0 {
			return errors.New("remote schema preparation requires an empty database")
		}
		if err := migrateModels(tx); err != nil {
			return err
		}
		return tx.Exec("INSERT INTO ech0_schema(id, version) VALUES (1, ?)", remoteSchemaVersion).Error
	})
}
