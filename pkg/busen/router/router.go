// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"errors"
	"slices"
	"strings"
)

var ErrInvalidPattern = errors.New("invalid topic pattern")

type Matcher interface {
	Match(topic string) bool
}

func Compile(pattern string) (Matcher, error) {
	segments, err := split(pattern)
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return exactMatcher(""), nil
	}

	hasWildcard := false
	for i, segment := range segments {
		switch segment {
		case "*":
			hasWildcard = true
		case ">":
			hasWildcard = true
			if i != len(segments)-1 {
				return nil, ErrInvalidPattern
			}
		default:
			if strings.Contains(segment, "*") || strings.Contains(segment, ">") {
				return nil, ErrInvalidPattern
			}
		}
	}

	if !hasWildcard {
		return exactMatcher(pattern), nil
	}

	return wildcardMatcher(segments), nil
}

type exactMatcher string

func (m exactMatcher) Match(topic string) bool {
	return string(m) == topic
}

type wildcardMatcher []string

func (m wildcardMatcher) Match(topic string) bool {
	pi := 0
	ti := 0
	for pi < len(m) {
		switch m[pi] {
		case "*":
			_, next, ok := nextSegment(topic, ti)
			if !ok {
				return false
			}
			pi++
			ti = next
		case ">":
			return ti < len(topic)
		default:
			segment, next, ok := nextSegment(topic, ti)
			if !ok || m[pi] != segment {
				return false
			}
			pi++
			ti = next
		}
	}

	return ti == len(topic)
}

func split(topic string) ([]string, error) {
	if topic == "" {
		return nil, nil
	}

	parts := strings.Split(topic, ".")
	if slices.Contains(parts, "") {
		return nil, ErrInvalidPattern
	}

	return parts, nil
}

func nextSegment(topic string, start int) (segment string, next int, ok bool) {
	if start >= len(topic) {
		return "", start, false
	}

	rest := topic[start:]
	end := strings.IndexByte(rest, '.')
	if end < 0 {
		segment = rest
		if segment == "" {
			return "", start, false
		}
		return segment, len(topic), true
	}

	segment = rest[:end]
	if segment == "" || start+end+1 >= len(topic) {
		return "", start, false
	}

	return segment, start + end + 1, true
}
