// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package clientip resolves the address a request should be attributed to.
//
// The peer address on the TCP connection is trusted; forwarding headers are
// only believed when the peer is an internal address, because that is the shape
// of a same-host or container-network reverse proxy. Walking the header from
// the right means a client cannot prepend a fake hop: the entries a trusted
// proxy appended are the ones that survive.
package clientip

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type contextKey struct{}

// WithClientIP stores the resolved address on a request context.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, contextKey{}, ip)
}

// FromContext returns the resolved address, or "" when unset.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	ip, _ := ctx.Value(contextKey{}).(string)
	return ip
}

// Resolve picks the client address for one request.
func Resolve(remoteAddr string, header http.Header) string {
	peer := parsePeer(remoteAddr)
	if peer == nil || !isInternal(peer) {
		return peerString(peer, remoteAddr)
	}

	forwarded := header.Values("X-Forwarded-For")
	for i := len(forwarded) - 1; i >= 0; i-- {
		entries := strings.Split(forwarded[i], ",")
		for j := len(entries) - 1; j >= 0; j-- {
			candidate := strings.TrimSpace(entries[j])
			ip := net.ParseIP(candidate)
			if ip == nil {
				// A malformed hop means the chain is not trustworthy from
				// here leftward; stop instead of guessing past it.
				return realIPOrPeer(header, peer)
			}
			if isInternal(ip) {
				continue
			}
			return candidate
		}
	}

	return realIPOrPeer(header, peer)
}

func realIPOrPeer(header http.Header, peer net.IP) string {
	realIP := strings.TrimSpace(header.Get("X-Real-Ip"))
	if ip := net.ParseIP(realIP); ip != nil && !isInternal(ip) {
		return realIP
	}
	return peer.String()
}

func parsePeer(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}
	return net.ParseIP(host)
}

func peerString(peer net.IP, remoteAddr string) string {
	if peer != nil {
		return peer.String()
	}
	return strings.TrimSpace(remoteAddr)
}

// isInternal reports whether an address belongs to the ranges a local reverse
// proxy or container runtime uses.
func isInternal(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
