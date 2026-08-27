// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	jobModel "github.com/lin-snow/ech0/internal/model/job"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const logModule = "job"

var (
	ErrNoRunner       = errors.New("no runner registered for job type")
	ErrAlreadyRunning = errors.New("a job of this type is already running")
)

type Manager struct {
	repo JobRepository

	mu      sync.Mutex
	runners map[string]Runner
	live    map[string]*Progress
	cancels map[string]context.CancelFunc
}

func NewManager(repo JobRepository) *Manager {
	return &Manager{
		repo:    repo,
		runners: make(map[string]Runner),
		live:    make(map[string]*Progress),
		cancels: make(map[string]context.CancelFunc),
	}
}

func (m *Manager) Register(jobType string, r Runner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runners[jobType] = r
}

func (m *Manager) Submit(ctx context.Context, jobType string, payload []byte) (jobModel.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	runner, ok := m.runners[jobType]
	if !ok {
		return jobModel.Job{}, fmt.Errorf("%w: %s", ErrNoRunner, jobType)
	}

	if existing, err := m.repo.GetByType(ctx, jobType); err == nil {
		if !existing.Status.IsTerminal() {
			return jobModel.Job{}, ErrAlreadyRunning
		}
	} else if !errors.Is(err, ErrNotFound) {
		return jobModel.Job{}, err
	}

	now := time.Now().UTC().Unix()
	pending := jobModel.Job{
		Type:      jobType,
		Status:    jobModel.StatusPending,
		Payload:   string(payload),
		StartedAt: &now,
	}
	if err := m.repo.Upsert(ctx, &pending); err != nil {
		return jobModel.Job{}, err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	m.cancels[jobType] = cancel
	delete(m.live, jobType)

	go m.run(runCtx, jobType, runner, pending)

	logUtil.GetLogger().Info("job submitted", slog.String("module", logModule), slog.String("type", jobType))
	return pending, nil
}

func (m *Manager) run(runCtx context.Context, jobType string, runner Runner, base jobModel.Job) {
	dbCtx := context.Background()
	report := func(phase string, snapshot any) { m.setLive(jobType, phase, snapshot) }

	base.Status = jobModel.StatusRunning
	if err := m.repo.Upsert(dbCtx, &base); err != nil {
		logUtil.GetLogger().Error("job mark running failed",
			slog.String("module", logModule), slog.String("type", jobType), logUtil.Err(err))
	}

	result, runErr := runner.Run(runCtx, []byte(base.Payload), report)

	now := time.Now().UTC().Unix()
	base.FinishedAt = &now
	base.Phase = m.takeLivePhase(jobType)

	switch {
	case errors.Is(runCtx.Err(), context.Canceled):
		base.Status = jobModel.StatusCancelled
		base.Error = ""
		logUtil.GetLogger().Warn("job cancelled", slog.String("module", logModule), slog.String("type", jobType))
	case runErr != nil:
		base.Status = jobModel.StatusFailed
		base.Error = runErr.Error()
		logUtil.GetLogger().Error("job failed",
			slog.String("module", logModule), slog.String("type", jobType), logUtil.Err(runErr))
	default:
		base.Status = jobModel.StatusSuccess
		base.Error = ""
		if result != nil {
			base.Payload = mustJSON(result)
		}
		logUtil.GetLogger().Info("job succeeded", slog.String("module", logModule), slog.String("type", jobType))
	}

	if err := m.repo.Upsert(dbCtx, &base); err != nil {
		logUtil.GetLogger().Error("job persist terminal failed",
			slog.String("module", logModule), slog.String("type", jobType),
			slog.String("status", string(base.Status)), logUtil.Err(err))
	}

	m.clear(jobType)
}

func (m *Manager) Get(ctx context.Context, jobType string) (jobModel.Job, error) {
	row, err := m.repo.GetByType(ctx, jobType)
	if err != nil {
		return row, err
	}
	m.mu.Lock()
	p := m.live[jobType]
	m.mu.Unlock()
	if p != nil {
		row.Phase = p.Phase
		if p.Snapshot != nil {
			row.Payload = mustJSON(p.Snapshot)
		}
	}
	return row, nil
}

func (m *Manager) Delete(ctx context.Context, jobType string) error {
	if err := m.repo.Delete(ctx, jobType); err != nil {
		return err
	}
	m.clear(jobType)
	return nil
}

func (m *Manager) Cancel(jobType string) error {
	m.mu.Lock()
	cancel := m.cancels[jobType]
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

func (m *Manager) setLive(jobType, phase string, snapshot any) {
	m.mu.Lock()
	m.live[jobType] = &Progress{Phase: phase, Snapshot: snapshot}
	m.mu.Unlock()
}

func (m *Manager) takeLivePhase(jobType string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p := m.live[jobType]; p != nil {
		return p.Phase
	}
	return ""
}

func (m *Manager) clear(jobType string) {
	m.mu.Lock()
	delete(m.live, jobType)
	delete(m.cancels, jobType)
	m.mu.Unlock()
}

func (m *Manager) Name() string { return "job" }

func (m *Manager) Start(ctx context.Context) error {
	if err := m.repo.SweepRunning(ctx, "interrupted by restart"); err != nil {
		logUtil.GetLogger().Error("sweep orphan jobs failed", slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	return nil
}

func (m *Manager) Stop(context.Context) error {
	m.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(m.cancels))
	for _, c := range m.cancels {
		cancels = append(cancels, c)
	}
	m.mu.Unlock()
	for _, c := range cancels {
		c()
	}
	return nil
}

func mustJSON(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
