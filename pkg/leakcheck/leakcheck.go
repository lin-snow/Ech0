// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package leakcheck

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"testing"
)

func Run(m *testing.M) int {
	code := m.Run()
	if code != 0 {
		return code
	}

	report, count, err := profile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL\tgoroutine leak check: %v\n", err)
		return 1
	}
	if count == 0 {
		return 0
	}

	fmt.Fprintf(os.Stderr, "FAIL\tgoroutine leak check: %d leaked goroutine(s)\n%s", count, report)
	return 1
}

func profile() (string, int, error) {
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		return "", 0, errors.New(`pprof profile "goroutineleak" is unavailable`)
	}

	var buf bytes.Buffer
	if err := p.WriteTo(&buf, 1); err != nil {
		return "", 0, fmt.Errorf("write goroutineleak profile: %w", err)
	}

	report := buf.String()
	header, _, _ := strings.Cut(report, "\n")
	count, err := parseTotal(header)
	if err != nil {
		return report, 0, err
	}
	return report, count, nil
}

func parseTotal(header string) (int, error) {
	prefix, total, ok := strings.CutLast(header, " ")
	if !ok || !strings.HasSuffix(prefix, "total") {
		return 0, fmt.Errorf("unexpected goroutineleak profile header %q", header)
	}
	count, err := strconv.Atoi(total)
	if err != nil {
		return 0, fmt.Errorf("unexpected goroutineleak profile header %q", header)
	}
	return count, nil
}
