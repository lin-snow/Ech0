// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"strings"
	"testing"
)

func runGuard(chunks ...string) (emitted string, tripped bool) {
	g := &toolCallLeakGuard{}
	var b strings.Builder
	for _, c := range chunks {
		safe, hit := g.feed(c)
		b.WriteString(safe)
		if hit {
			return b.String(), true
		}
	}
	b.WriteString(g.flush())
	return b.String(), false
}

func TestLeakGuard_NormalTextPassesThrough(t *testing.T) {
	in := []string{"根据你的 Echo，", "你最近在读《", "三体》，", "状态不错 🙂"}
	emitted, tripped := runGuard(in...)
	if tripped {
		t.Fatalf("normal text should not trip the guard")
	}
	if want := strings.Join(in, ""); emitted != want {
		t.Fatalf("text should pass through losslessly:\n want %q\n got  %q", want, emitted)
	}
}

func TestLeakGuard_FullMarkerInOneChunk(t *testing.T) {
	if _, tripped := runGuard("好的<tool_call> <function=search_echos>"); !tripped {
		t.Fatalf("a full <tool_call> marker should trip the guard")
	}
	if _, tripped := runGuard("x<function=search_echos>"); !tripped {
		t.Fatalf("a full <function= marker should trip the guard")
	}
}

func TestLeakGuard_MarkerSplitAcrossChunks(t *testing.T) {
	emitted, tripped := runGuard("前文abc<tool", "_call>技术")
	if !tripped {
		t.Fatalf("a marker split across chunks should still trip")
	}
	if strings.Contains(emitted, "<tool") {
		t.Fatalf("held partial marker must not leak before tripping, got %q", emitted)
	}
}

func TestLeakGuard_PartialPrefixReleasedWhenNotMarker(t *testing.T) {
	in := []string{"完成<fun", "ny 想法"}
	emitted, tripped := runGuard(in...)
	if tripped {
		t.Fatalf("a partial prefix that resolves to non-marker should not trip")
	}
	if want := strings.Join(in, ""); emitted != want {
		t.Fatalf("released partial prefix should be lossless:\n want %q\n got  %q", want, emitted)
	}
}

func TestLeakGuard_MultibyteSafe(t *testing.T) {
	in := []string{"a < b 且 c < d，", "继续写想法"}
	emitted, tripped := runGuard(in...)
	if tripped {
		t.Fatalf("plain '<' in prose should not trip")
	}
	if want := strings.Join(in, ""); emitted != want {
		t.Fatalf("multibyte text should be lossless:\n want %q\n got  %q", want, emitted)
	}
}
