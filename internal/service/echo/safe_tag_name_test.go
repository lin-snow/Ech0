// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSafeTagName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"plain ascii", "golang", true},
		{"chinese", "技术分享", true},
		{"with space", "hello world", true},
		{"with digits and dash", "go-1-26", true},
		{"empty is safe", "", true},
		{"open angle bracket", "a<b", false},
		{"close angle bracket", "a>b", false},
		{"script payload", "<script>alert(1)</script>", false},
		{"double quote", `a"b`, false},
		{"single quote", "a'b", false},
		{"ampersand", "a&b", false},
		{"html entity attempt", "&lt;img&gt;", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isSafeTagName(tc.in))
		})
	}
}
