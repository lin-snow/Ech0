// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package async

import (
	"sync"

	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type WorkerPool struct {
	workerCount int
	jobs        chan func() error
	wg          sync.WaitGroup
	mu          sync.RWMutex
	stopped     bool
	stopOnce    sync.Once
}

func NewWorkerPool(workerCount, jobQueueSize int) *WorkerPool {
	workerPool := &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func() error, jobQueueSize),
	}
	workerPool.start()
	return workerPool
}

func (p *WorkerPool) start() {
	for i := 0; i < p.workerCount; i++ {
		go func() {
			for job := range p.jobs {
				func() {
					defer p.wg.Done()
					if err := job(); err != nil {
						logUtil.GetLogger().
							Error("worker job failed", logUtil.Err(err))
					}
				}()
			}
		}()
	}
}

func (p *WorkerPool) Submit(job func() error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.stopped {
		return
	}
	p.wg.Add(1)
	p.jobs <- job
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

func (p *WorkerPool) Stop() {
	p.stopOnce.Do(func() {
		p.mu.Lock()
		p.stopped = true
		close(p.jobs)
		p.mu.Unlock()
	})
	p.Wait()
}
