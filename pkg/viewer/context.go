// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

import (
	"context"
	"net/http"
)

type contextKey struct{}

func WithContext(ctx context.Context, v Context) context.Context {
	return context.WithValue(ctx, contextKey{}, v)
}

func FromContext(ctx context.Context) (Context, bool) {
	if ctx == nil {
		return nil, false
	}
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil, false
	}
	vc, ok := v.(Context)
	return vc, ok
}

// MustFromContext returns the principal attached to ctx, or Anonymous when no
// middleware ran. Missing middleware therefore cannot grant identity: the
// fallback is the anonymous value, never a synthetic user.
func MustFromContext(ctx context.Context) Context {
	if ctx == nil {
		return Anonymous
	}
	if v := ctx.Value(contextKey{}); v != nil {
		if vc, ok := v.(Context); ok && vc != nil {
			return vc
		}
	}
	return Anonymous
}

func WithRequest(req *http.Request, v Context) *http.Request {
	if req == nil {
		return nil
	}
	reqCtx := req.Context()
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	return req.WithContext(WithContext(reqCtx, v))
}

func AttachToRequest(req **http.Request, v Context) {
	if req == nil || *req == nil {
		return
	}
	*req = WithRequest(*req, v)
}
