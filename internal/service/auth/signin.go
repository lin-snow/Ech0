// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"context"
	"net/http"

	"github.com/lin-snow/ech0/internal/clientip"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/ratelimit"
)

// signInAttempt is budget reserved for one sign-in try. It is refunded only
// after the credentials are proven, so failures keep their cost while a burst
// of correct sign-ins does not exhaust the account's allowance.
type signInAttempt struct {
	limits      *ratelimit.MemoryLimiter
	addressKey  string
	accountKey  string
	chargedAddr bool
	chargedAcct bool
}

func (authService *AuthService) reserveSignIn(ctx context.Context, username string) (*signInAttempt, error) {
	attempt := &signInAttempt{
		limits:     authService.limits,
		accountKey: ratelimit.AccountKey(username),
	}

	if address := clientip.FromContext(ctx); address != "" {
		attempt.addressKey = address
		decision := attempt.limits.Consume(ratelimit.ScopeSignInIP, address, 1)
		if !decision.Allowed {
			return nil, signInRateLimited(decision)
		}
		attempt.chargedAddr = true
	}

	if attempt.accountKey != "" {
		decision := attempt.limits.Consume(ratelimit.ScopeSignInAccount, attempt.accountKey, 1)
		if !decision.Allowed {
			// Do not punish the address for a denied account budget.
			attempt.refundAddress()
			return nil, signInRateLimited(decision)
		}
		attempt.chargedAcct = true
	}

	return attempt, nil
}

func (attempt *signInAttempt) succeeded() {
	attempt.refundAddress()
	if attempt.chargedAcct {
		attempt.limits.Refund(ratelimit.ScopeSignInAccount, attempt.accountKey, 1)
		attempt.chargedAcct = false
	}
}

func (attempt *signInAttempt) refundAddress() {
	if attempt.chargedAddr {
		attempt.limits.Refund(ratelimit.ScopeSignInIP, attempt.addressKey, 1)
		attempt.chargedAddr = false
	}
}

// signInRateLimited reports a 429. Huma renders the problem document itself,
// and huma.StatusError has no header channel, so the cool-down is conveyed by
// the status alone; the middleware-set Retry-After covers the transport layer.
func signInRateLimited(_ ratelimit.Decision) error {
	return commonModel.Fail(http.StatusTooManyRequests, commonModel.RATE_LIMIT_EXCEEDED)
}
