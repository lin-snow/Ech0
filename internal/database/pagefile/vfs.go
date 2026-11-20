// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:build cgo

package pagefile

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/lin-snow/ech0/internal/util/uuid"
	_ "github.com/mattn/go-sqlite3"
	vfs "github.com/psanford/sqlite3vfs"
)

const vfsName = "ech0_snapshot_blocks"

type registry struct{ files sync.Map }

var files registry
var register = sync.OnceValue(func() error { return vfs.RegisterVFS(vfsName, &files) })

// UseSQLite opens one isolated connection on the overlay. A failed callback
// invalidates the overlay: its caller must discard it without checkpointing.
func UseSQLite(ctx context.Context, file *File, fn func(*sql.DB) error) error {
	if err := register(); err != nil {
		return err
	}
	name := uuid.NewV7()
	files.files.Store(name, file)
	defer files.files.Delete(name)
	db, err := sql.Open("sqlite3", "file:"+name+"?vfs="+vfsName+"&_journal_mode=MEMORY&_sync=OFF")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, "PRAGMA temp_store=MEMORY")
	if err == nil {
		err = fn(db)
	}
	return errors.Join(err, db.Close())
}

func (r *registry) Open(name string, flags vfs.OpenFlag) (vfs.File, vfs.OpenFlag, error) {
	file, ok := r.files.Load(name)
	if !ok || flags&vfs.OpenMainDB == 0 {
		return nil, 0, vfs.CantOpenError
	}
	return &handle{File: file.(*File)}, flags, nil
}
func (r *registry) Access(name string, _ vfs.AccessFlag) (bool, error) {
	_, ok := r.files.Load(name)
	return ok, nil
}
func (r *registry) Delete(string, bool) error       { return vfs.PermError }
func (r *registry) FullPathname(name string) string { return name }

type handle struct{ *File }

func (h *handle) Close() error                                    { return nil }
func (h *handle) Sync(vfs.SyncType) error                         { return nil }
func (h *handle) FileSize() (int64, error)                        { return h.Size, nil }
func (h *handle) Lock(vfs.LockType) error                         { return nil }
func (h *handle) Unlock(vfs.LockType) error                       { return nil }
func (h *handle) CheckReservedLock() (bool, error)                { return false, nil }
func (h *handle) SectorSize() int64                               { return 4096 }
func (h *handle) DeviceCharacteristics() vfs.DeviceCharacteristic { return 0 }
