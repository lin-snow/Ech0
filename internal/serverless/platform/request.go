// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type requestKey struct{}
type identity struct{ token string }

func WithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestKey{}, identity{token: r.Header.Get("x-vercel-oidc-token")})
}
func IsFunction(ctx context.Context) bool { _, ok := ctx.Value(requestKey{}).(identity); return ok }

var regionName = regexp.MustCompile(`^[a-z]{3}[0-9]+$`)

func Wake(ctx context.Context, id string, revision int64) error {
	return Publish(ctx, map[string]any{"run": id, "revision": revision}, "snapshot-"+id+"-"+strconv.FormatInt(revision, 10), 0)
}

func Publish(ctx context.Context, payload any, key string, delay int64) error {
	identity, _ := ctx.Value(requestKey{}).(identity)
	region := os.Getenv("VERCEL_REGION")
	if identity.token == "" || !regionName.MatchString(region) {
		return errors.New("queue identity unavailable")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+region+".vercel-queue.com/api/v3/topic/ech0-maintenance", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+identity.token)
	req.Header.Set("Vqs-Idempotency-Key", key)
	req.Header.Set("Vqs-Delay-Seconds", strconv.FormatInt(delay, 10))
	req.Header.Set("Vqs-Retention-Seconds", "259200")
	req.Header.Set("Vqs-Deployment-Id", os.Getenv("VERCEL_DEPLOYMENT_ID"))
	client := http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(req)
	if err != nil {
		return errors.New("queue publication unavailable")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusAccepted {
		return errors.New("queue refused publication")
	}
	return nil
}
