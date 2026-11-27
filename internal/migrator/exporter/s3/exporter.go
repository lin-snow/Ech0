// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package s3

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/snapshot"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const (
	uploadTimeout = 60 * time.Minute
	snapshotKey   = "snapshots/"
)

// uploader is the one thing this exporter needs from the storage layer: the
// default instance an artifact is uploaded to. The registry owns driver choice
// and configuration; there is no second decision to make here.
type uploader interface {
	Default() (string, storage.Store, error)
}

type Exporter struct {
	uploader uploader
}

func New(up uploader) *Exporter {
	return &Exporter{uploader: up}
}

func (e *Exporter) Export(_ context.Context, req spec.ExportRequest) (spec.ExportResult, error) {
	emit(req, migratorModel.ExportPhasePacking)
	path, fileName, err := snapshot.Create(snapshot.WithConsistentDB(database.SnapshotTo))
	if err != nil {
		return spec.ExportResult{}, err
	}

	var size int64
	if info, statErr := os.Stat(path); statErr == nil {
		size = info.Size()
	}

	go e.upload(path, fileName)

	emit(req, migratorModel.ExportPhaseCompleted)
	return spec.ExportResult{ArtifactPath: path, FileName: fileName, Size: size}, nil
}

func (e *Exporter) upload(artifactPath, fileName string) {
	if e.uploader == nil {
		return
	}

	uploadCtx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
	defer cancel()

	if err := e.put(uploadCtx, artifactPath, fileName); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			logUtil.GetLogger().Warn("Failed to upload snapshot: upload timeout reached",
				slog.Duration("timeout", uploadTimeout), logUtil.Err(err))
		case errors.Is(err, context.Canceled):
			logUtil.GetLogger().Warn("Failed to upload snapshot: upload context canceled", logUtil.Err(err))
		default:
			logUtil.GetLogger().Warn("Failed to upload snapshot", logUtil.Err(err))
		}
	}
}

func (e *Exporter) put(ctx context.Context, artifactPath, fileName string) error {
	f, err := os.Open(artifactPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	_, store, err := e.uploader.Default()
	if err != nil {
		return err
	}

	key := snapshotKey + fileName
	if _, err := store.Put(ctx, key, f, info.Size()); err != nil {
		return err
	}

	logUtil.GetLogger().Info("Snapshot uploaded to storage", slog.String("key", key))
	return nil
}

func emit(req spec.ExportRequest, phase string) {
	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.ExportProgress{CurrentPhase: phase})
	}
}
