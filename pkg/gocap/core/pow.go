// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

func PRNG(seed string, length int) string {
	if length <= 0 {
		return ""
	}

	state := fnv1a32(seed)
	var out strings.Builder
	out.Grow(length + 8)

	for out.Len() < length {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		out.WriteString(leftPadHex8(state))
	}
	s := out.String()
	return s[:length]
}

func BuildChallengePairs(seed string, count, saltSize, difficulty int) [][2]string {
	pairs := make([][2]string, count)
	for i := range count {
		idx := strconv.Itoa(i + 1)
		salt := PRNG(seed+idx, saltSize)
		target := PRNG(seed+idx+"d", difficulty)
		pairs[i] = [2]string{salt, target}
	}
	return pairs
}

func VerifySolutions(seed string, count, saltSize, difficulty int, solutions []int) bool {
	if len(solutions) != count {
		return false
	}
	pairs := BuildChallengePairs(seed, count, saltSize, difficulty)
	for i, pair := range pairs {
		sum := sha256.Sum256([]byte(pair[0] + strconv.Itoa(solutions[i])))
		h := hex.EncodeToString(sum[:])
		if !strings.HasPrefix(h, pair[1]) {
			return false
		}
	}
	return true
}

func fnv1a32(s string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash += (hash << 1) + (hash << 4) + (hash << 7) + (hash << 8) + (hash << 24)
	}
	return hash
}

func leftPadHex8(v uint32) string {
	const hexdigits = "0123456789abcdef"
	buf := [8]byte{}
	for i := 7; i >= 0; i-- {
		buf[i] = hexdigits[v&0xF]
		v >>= 4
	}
	return string(buf[:])
}
