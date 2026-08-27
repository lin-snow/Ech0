// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	defaultSafeResponseBodyLimitBytes int64 = 1 << 20
	maxSafeRedirects                        = 3
)

var blockedCIDRs = mustParseCIDRs(
	[]string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"224.0.0.0/4",
		"240.0.0.0/4",
		"::/128",
		"::1/128",
		"fe80::/10",
		"fc00::/7",
		"ff00::/8",
	},
)

func mustParseCIDRs(cidrs []string) []*net.IPNet {
	result := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid cidr %s: %v", cidr, err))
		}
		result = append(result, ipNet)
	}
	return result
}

func isPrivateOrReservedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	for _, ipNet := range blockedCIDRs {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func isBlockedHostname(hostname string) bool {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	return hostname == "localhost" ||
		strings.HasSuffix(hostname, ".localhost") ||
		hostname == "host.docker.internal" ||
		hostname == "gateway.docker.internal"
}

func Validate(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL 格式无效: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("URL 必须包含协议和主机")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("仅支持 http/https 协议")
	}
	if parsed.User != nil {
		return errors.New("URL 不允许包含用户信息")
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return errors.New("URL 主机无效")
	}
	if isBlockedHostname(hostname) {
		return errors.New("目标主机不被允许")
	}
	if ip := net.ParseIP(hostname); ip != nil && isPrivateOrReservedIP(ip) {
		return errors.New("目标 IP 不被允许")
	}
	return nil
}

func secureDialContext(timeout time.Duration) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
		if !ok || tcpAddr == nil || isPrivateOrReservedIP(tcpAddr.IP) {
			_ = conn.Close()
			return nil, errors.New("连接目标地址不被允许")
		}
		return conn, nil
	}
}

func readBodyWithLimit(reader io.Reader, maxBytes int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("响应体超过限制: %d bytes", maxBytes)
	}
	return body, nil
}
