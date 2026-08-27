// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router_test

import (
	"errors"
	"testing"

	"github.com/lin-snow/ech0/pkg/busen/router"
)

func TestCompileInvalidPatterns(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
	}{
		{name: "greater than in middle", pattern: "a.>.c"},
		{name: "greater than not last among many", pattern: ">.a"},
		{name: "empty middle segment", pattern: "a..b"},
		{name: "leading dot", pattern: ".a"},
		{name: "trailing dot", pattern: "a."},
		{name: "only dot", pattern: "."},
		{name: "embedded star in literal", pattern: "a*"},
		{name: "embedded greater than in literal", pattern: "a>"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := router.Compile(tc.pattern)
			if !errors.Is(err, router.ErrInvalidPattern) {
				t.Fatalf("Compile(%q) err = %v, want ErrInvalidPattern", tc.pattern, err)
			}
			if m != nil {
				t.Fatalf("Compile(%q) matcher = %v, want nil on error", tc.pattern, m)
			}
		})
	}
}

func TestCompileMatch(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		topic   string
		want    bool
	}{
		{name: "empty matches empty", pattern: "", topic: "", want: true},
		{name: "empty rejects nonempty", pattern: "", topic: "a", want: false},

		{name: "exact equal", pattern: "a.b.c", topic: "a.b.c", want: true},
		{name: "exact last differs", pattern: "a.b.c", topic: "a.b.d", want: false},
		{name: "exact too short", pattern: "a.b.c", topic: "a.b", want: false},
		{name: "single literal", pattern: "a", topic: "a", want: true},

		{name: "star matches one segment", pattern: "*", topic: "a", want: true},
		{name: "star rejects two segments", pattern: "*", topic: "a.b", want: false},
		{name: "star rejects empty", pattern: "*", topic: "", want: false},

		{name: "mid star match", pattern: "a.*.c", topic: "a.b.c", want: true},
		{name: "mid star other value", pattern: "a.*.c", topic: "a.x.c", want: true},
		{name: "mid star last differs", pattern: "a.*.c", topic: "a.b.d", want: false},
		{name: "mid star too few segments", pattern: "a.*.c", topic: "a.b", want: false},
		{name: "mid star too many segments", pattern: "a.*.c", topic: "a.b.c.d", want: false},

		{name: "trailing star match", pattern: "a.*", topic: "a.b", want: true},
		{name: "trailing star too many", pattern: "a.*", topic: "a.b.c", want: false},
		{name: "trailing star too few", pattern: "a.*", topic: "a", want: false},

		{name: "tail gt single segment", pattern: "a.>", topic: "a.b", want: true},
		{name: "tail gt multi segment", pattern: "a.>", topic: "a.b.c", want: true},
		{name: "tail gt empty tail rejected", pattern: "a.>", topic: "a", want: false},
		{name: "tail gt prefix differs", pattern: "a.>", topic: "x.b", want: false},

		{name: "bare gt single", pattern: ">", topic: "a", want: true},
		{name: "bare gt multi", pattern: ">", topic: "a.b.c", want: true},
		{name: "bare gt empty rejected", pattern: ">", topic: "", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := router.Compile(tc.pattern)
			if err != nil {
				t.Fatalf("Compile(%q) unexpected err: %v", tc.pattern, err)
			}
			if m == nil {
				t.Fatalf("Compile(%q) returned nil matcher", tc.pattern)
			}
			if got := m.Match(tc.topic); got != tc.want {
				t.Fatalf("Compile(%q).Match(%q) = %v, want %v", tc.pattern, tc.topic, got, tc.want)
			}
		})
	}
}
