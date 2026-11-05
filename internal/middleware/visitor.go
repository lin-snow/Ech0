// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	actorUtil "github.com/lin-snow/ech0/internal/util/actor"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

// VisitorIDHeader carries the client-generated anonymous identity used for
// reactions. It is not a credential: it only keeps one browser's reactions
// separable, and abuse is bounded by the per-IP rate limit.
const VisitorIDHeader = "X-Ech0-Visitor"

// VisitorIdentity resolves the request's anonymous actor once and stores it in
// the request context. The echo service prefers the authenticated user and
// only falls back to this value.
func VisitorIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil {
			c.Next()
			return
		}

		value := ""
		if visitorID := strings.TrimSpace(c.GetHeader(VisitorIDHeader)); uuidUtil.IsValid(visitorID) {
			value = "v:" + visitorID
		} else if ip := strings.TrimSpace(c.ClientIP()); ip != "" {
			sum := sha256.Sum256([]byte(ip))
			value = "ip:" + hex.EncodeToString(sum[:])
		}

		if value != "" {
			c.Request = c.Request.WithContext(actorUtil.WithContext(c.Request.Context(), value))
		}
		c.Next()
	}
}
