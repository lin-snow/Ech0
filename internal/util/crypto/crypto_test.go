// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMD5Encrypt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "d41d8cd98f00b204e9800998ecf8427e"},
		{name: "abc", in: "abc", want: "900150983cd24fb0d6963f7d28e17f72"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MD5Encrypt(tc.in))
		})
	}
}

func TestHashPassword_RoundTrip(t *testing.T) {
	const plain = "s3cr3t-pw"
	hash, err := HashPassword(plain)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(hash, "$2"), "want bcrypt hash, got %q", hash)
	other, err := HashPassword(plain)
	require.NoError(t, err)
	assert.NotEqual(t, hash, other, "bcrypt 每次应产生不同盐")

	assert.True(t, CheckPassword(AlgoBcrypt, hash, plain))
	assert.False(t, CheckPassword(AlgoBcrypt, hash, "wrong-pw"))
}

func TestCheckPassword_LegacyMD5(t *testing.T) {
	const plain = "old-pw"
	md5Hash := MD5Encrypt(plain)

	assert.True(t, CheckPassword(AlgoMD5, md5Hash, plain))
	assert.False(t, CheckPassword(AlgoMD5, md5Hash, "nope"))

	assert.True(t, CheckPassword("", md5Hash, plain))
}

func TestGenerateRandomString_Length(t *testing.T) {
	for _, n := range []int{1, 8, 16, 32, 64, 256} {
		got := GenerateRandomString(n)
		assert.Len(t, got, n, "length %d", n)
	}
}

func TestGenerateRandomString_NonPositiveReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", GenerateRandomString(0))
	assert.Equal(t, "", GenerateRandomString(-5))
}

func TestGenerateRandomString_CharsetOnly(t *testing.T) {
	s := GenerateRandomString(10000)
	for i, r := range s {
		require.True(t, strings.ContainsRune(randomCharset, r),
			"index %d produced out-of-charset rune %q", i, r)
	}
}

func TestGenerateRandomString_NoCollisions(t *testing.T) {
	const (
		count  = 50000
		length = 32
	)
	seen := make(map[string]struct{}, count)
	for i := range count {
		s := GenerateRandomString(length)
		_, dup := seen[s]
		require.Falsef(t, dup, "collision after %d generations: %q", i, s)
		seen[s] = struct{}{}
	}
}

func TestGenerateRandomString_UsesAllCharsetSymbols(t *testing.T) {
	s := GenerateRandomString(100000)
	for _, c := range randomCharset {
		assert.Truef(t, strings.ContainsRune(s, c), "charset symbol %q never appeared", c)
	}
}
