// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	copilotModel "github.com/lin-snow/ech0/internal/model/copilot"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	"github.com/lin-snow/ech0/internal/securitystate"
	"github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const remoteSchemaVersion = 9

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

// OpenRemoteDatabase returns a pool only after verifying the provisioned schema.
func OpenRemoteDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := openLibSQL(cfg, logger.Silent)
	if err != nil {
		return nil, err
	}
	if err := checkRemoteSchema(db); err != nil {
		pool, _ := db.DB()
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
	defer func() { _ = pool.Close() }()
	return prepareRemoteSchema(db.WithContext(ctx))
}

// Read the exact table and index DDL produced by the upstream GORM models.
// Compiling on a private in-memory connection keeps ORM introspection outside
// Turso's short interactive transaction window.
var modelDDL = sync.OnceValues(func() (string, error) {
	if err := enableLocalSQLite(); err != nil {
		return "", nil
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), buildGormConfig(logger.Silent))
	if err != nil {
		return "", err
	}
	pool, err := db.DB()
	if err != nil {
		return "", err
	}
	pool.SetMaxOpenConns(1)
	defer func() { _ = pool.Close() }()
	if err := migrateModels(db); err != nil {
		return "", err
	}
	var statements []string
	if err := db.Raw("SELECT sql FROM sqlite_schema WHERE type IN ('table','index') AND name NOT GLOB 'sqlite_*' AND sql IS NOT NULL ORDER BY CASE type WHEN 'table' THEN 0 ELSE 1 END,name").Scan(&statements).Error; err != nil {
		return "", err
	}
	return strings.Join(statements, ";"), nil
})

func prepareRemoteSchema(db *gorm.DB) error {
	ddl, err := modelDDL()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("CREATE TABLE IF NOT EXISTS ech0_schema (id INTEGER PRIMARY KEY CHECK(id = 1), version INTEGER NOT NULL)").Error; err != nil {
			return err
		}
		var version int
		if err := tx.Raw("SELECT coalesce(max(version),0) FROM ech0_schema").Scan(&version).Error; err != nil {
			return err
		}
		if version == 1 {
			if err := tx.AutoMigrate(&oauthserverModel.ConsumedCode{}); err != nil {
				return err
			}
			version = 2
		}
		if version == 2 {
			if err := tx.AutoMigrate(&fileModel.UploadIntent{}); err != nil {
				return err
			}
			version = 3
		}
		if version == 3 {
			if err := tx.AutoMigrate(&authModel.RevokedToken{}); err != nil {
				return err
			}
			version = 4
		}
		if version == 4 {
			if err := tx.AutoMigrate(&authModel.Challenge{}); err != nil {
				return err
			}
			version = 5
		}
		if version == 5 {
			if err := tx.AutoMigrate(&visitorModel.VisitorSeen{}); err != nil {
				return err
			}
			version = 6
		}
		if version == 6 {
			if err := tx.AutoMigrate(&securitystate.Entry{}); err != nil {
				return err
			}
			version = 7
		}
		if version == 7 {
			if err := tx.AutoMigrate(&authModel.DeviceGrant{}); err != nil {
				return err
			}
			version = 8
		}
		if version == 8 {
			if err := tx.AutoMigrate(&copilotModel.Round{}); err != nil {
				return err
			}
			return tx.Exec("UPDATE ech0_schema SET version = ? WHERE id = 1", remoteSchemaVersion).Error
		}
		if version != 0 {
			return checkRemoteSchema(tx)
		}
		var existing int
		if err := tx.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name <> 'ech0_schema' AND name NOT GLOB 'sqlite_*' AND name NOT GLOB '_litestream_*' AND name NOT GLOB 'libsql_*'").Scan(&existing).Error; err != nil {
			return err
		}
		if existing != 0 {
			return errors.New("remote schema preparation requires an empty database")
		}
		if ddl != "" {
			if err := tx.Exec(ddl).Error; err != nil {
				return err
			}
		} else if err := migrateModels(tx); err != nil {
			return err
		}
		return tx.Exec("INSERT INTO ech0_schema(id, version) VALUES (1, ?)", remoteSchemaVersion).Error
	})
}
