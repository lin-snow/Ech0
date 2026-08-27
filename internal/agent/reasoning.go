// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"regexp"
	"strings"
)

var thinkBlockRe = regexp.MustCompile(`(?is)<think(?:ing)?\b[^>]*>.*?</think(?:ing)?\s*>`)

func stripReasoning(s string) string {
	return strings.TrimSpace(thinkBlockRe.ReplaceAllString(s, ""))
}

var (
	thinkOpenMarkers  = []string{"<think>", "<thinking>"}
	thinkCloseMarkers = []string{"</think>", "</thinking>"}
)

type reasoningSplitter struct {
	inThink bool
	pending string
}

func (r *reasoningSplitter) feed(s string) (answer, reasoning string) {
	r.pending += s
	var ans, rea strings.Builder
	for {
		if r.inThink {
			idx, mlen := firstFoldMarker(r.pending, thinkCloseMarkers)
			if idx < 0 {
				hold := foldPrefixHold(r.pending, thinkCloseMarkers)
				rea.WriteString(r.pending[:len(r.pending)-hold])
				r.pending = r.pending[len(r.pending)-hold:]
				break
			}
			rea.WriteString(r.pending[:idx])
			r.pending = r.pending[idx+mlen:]
			r.inThink = false
			continue
		}
		idx, mlen := firstFoldMarker(r.pending, thinkOpenMarkers)
		if idx < 0 {
			hold := foldPrefixHold(r.pending, thinkOpenMarkers)
			ans.WriteString(r.pending[:len(r.pending)-hold])
			r.pending = r.pending[len(r.pending)-hold:]
			break
		}
		ans.WriteString(r.pending[:idx])
		r.pending = r.pending[idx+mlen:]
		r.inThink = true
	}
	return ans.String(), rea.String()
}

func (r *reasoningSplitter) flush() (answer, reasoning string) {
	rest := r.pending
	r.pending = ""
	if r.inThink {
		return "", rest
	}
	return rest, ""
}

func firstFoldMarker(s string, markers []string) (idx, mlen int) {
	idx = -1
	for _, m := range markers {
		if i := foldIndex(s, m); i >= 0 && (idx < 0 || i < idx) {
			idx, mlen = i, len(m)
		}
	}
	return idx, mlen
}

func foldIndex(s, marker string) int {
	for i := 0; i+len(marker) <= len(s); i++ {
		if foldHasPrefix(s[i:], marker) {
			return i
		}
	}
	return -1
}

func foldPrefixHold(s string, markers []string) int {
	hold := 0
	for _, m := range markers {
		n := min(len(m), len(s))
		for k := n; k > hold; k-- {
			if foldHasPrefix(m, s[len(s)-k:]) {
				hold = k
				break
			}
		}
	}
	return hold
}

func foldHasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if asciiLower(s[i]) != asciiLower(prefix[i]) {
			return false
		}
	}
	return true
}

func asciiLower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
