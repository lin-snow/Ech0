// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package async

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool_SubmitWaitRunsAllJobs(t *testing.T) {
	cases := []struct {
		name        string
		workerCount int
		queueSize   int
		jobs        int
	}{
		{"single worker", 1, 1, 50},
		{"multi worker", 4, 8, 200},
		{"more workers than jobs", 8, 16, 3},
		{"unbuffered queue", 2, 0, 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewWorkerPool(tc.workerCount, tc.queueSize)
			defer pool.Stop()

			var done atomic.Int32
			for range tc.jobs {
				pool.Submit(func() error {
					done.Add(1)
					return nil
				})
			}
			pool.Wait()

			assert.Equal(t, int32(tc.jobs), done.Load(), "全部任务都应被执行")
		})
	}
}

func TestWorkerPool_JobErrorsDoNotBlockWait(t *testing.T) {
	pool := NewWorkerPool(3, 4)
	defer pool.Stop()

	const total = 60
	var executed atomic.Int32
	for i := range total {
		failing := i%2 == 0
		pool.Submit(func() error {
			executed.Add(1)
			if failing {
				return errors.New("boom")
			}
			return nil
		})
	}
	pool.Wait()

	assert.Equal(t, int32(total), executed.Load(), "返回错误的任务也应计入执行并不阻塞 Wait")
}

func TestWorkerPool_SubmitAfterStopIsNoop(t *testing.T) {
	pool := NewWorkerPool(2, 4)

	var before atomic.Int32
	for range 10 {
		pool.Submit(func() error {
			before.Add(1)
			return nil
		})
	}
	pool.Stop()
	require.Equal(t, int32(10), before.Load(), "Stop 前提交的任务应全部完成")

	var after atomic.Int32
	require.NotPanics(t, func() {
		for range 10 {
			pool.Submit(func() error {
				after.Add(1)
				return nil
			})
		}
	}, "Stop 之后 Submit 不应 panic")

	pool.Wait()
	assert.Zero(t, after.Load(), "Stop 之后提交的任务应被丢弃，不被执行")
}

func TestWorkerPool_StopIsIdempotent(t *testing.T) {
	pool := NewWorkerPool(2, 2)

	var done atomic.Int32
	for range 5 {
		pool.Submit(func() error {
			done.Add(1)
			return nil
		})
	}

	require.NotPanics(t, func() {
		pool.Stop()
		pool.Stop()
		pool.Stop()
	}, "重复 Stop 不应 panic")
	assert.Equal(t, int32(5), done.Load())
}

func TestWorkerPool_StopWaitsForInFlightJobs(t *testing.T) {
	pool := NewWorkerPool(1, 8)

	release := make(chan struct{})
	started := make(chan struct{})
	var finished atomic.Int32

	pool.Submit(func() error {
		close(started)
		<-release
		finished.Add(1)
		return nil
	})
	<-started

	const queued = 5
	for range queued {
		pool.Submit(func() error {
			finished.Add(1)
			return nil
		})
	}

	stopReturned := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopReturned)
	}()

	select {
	case <-stopReturned:
		t.Fatal("Stop 不应在仍有在途任务时返回")
	default:
	}

	close(release)
	<-stopReturned

	assert.Equal(t, int32(queued+1), finished.Load(), "Stop 返回时所有入队任务都应已执行")
}

func TestWorkerPool_ConcurrentSubmitAndStop(t *testing.T) {
	pool := NewWorkerPool(4, 8)

	const submitters = 16
	const perSubmitter = 50

	var executed atomic.Int32
	var submittersWg sync.WaitGroup
	submittersWg.Add(submitters)
	for range submitters {
		go func() {
			defer submittersWg.Done()
			for range perSubmitter {
				pool.Submit(func() error {
					executed.Add(1)
					return nil
				})
			}
		}()
	}

	pool.Stop()

	require.NotPanics(t, func() {
		submittersWg.Wait()
		pool.Wait()
	}, "并发 Submit/Stop 不应 panic 或死锁")

	got := executed.Load()
	assert.GreaterOrEqual(t, got, int32(0))
	assert.LessOrEqual(t, got, int32(submitters*perSubmitter), "执行数不应超过提交总数")
}
