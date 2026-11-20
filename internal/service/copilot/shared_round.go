// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"time"
)

type RoundStore interface {
	Register(ctx context.Context, id, owner string, deadline time.Time) error
	Reply(ctx context.Context, id, owner string, payload []byte) (bool, error)
	Read(ctx context.Context, id string) ([]byte, error)
	Close(ctx context.Context, id string) error
}

func (r *askRegistry) register(ctx context.Context, id, owner string) (wait func() ([]AskAnswer, error), release func(), err error) {
	if r.shared == nil {
		pending := r.open(id, owner)
		return func() ([]AskAnswer, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case answers := <-pending.answers:
				return answers, nil
			}
		}, func() { r.close(id) }, nil
	}
	deadline, _ := ctx.Deadline()
	if err := r.shared.Register(ctx, id, owner, deadline); err != nil {
		return nil, nil, err
	}
	return func() ([]AskAnswer, error) { return r.waitForReply(ctx, id) }, func() {
		// Expired rows are swept at registration even if the worker died.
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = r.shared.Close(cleanup, id)
	}, nil
}

func (r *askRegistry) waitForReply(ctx context.Context, id string) ([]AskAnswer, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			payload, err := r.shared.Read(ctx, id)
			if err != nil {
				return nil, err
			}
			if payload != nil {
				var answers []AskAnswer
				if err := json.Unmarshal(payload, &answers); err != nil {
					return nil, err
				}
				return answers, nil
			}
		}
	}
}
