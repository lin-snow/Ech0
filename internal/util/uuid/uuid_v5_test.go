// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package uuid

import (
	"testing"
	"uuid"
)

var nameSpaceDNS = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func TestNewV5Golden(t *testing.T) {
	capsuleNS := NewV5(NameSpaceURL, []byte("https://github.com/lin-snow/Ech0/capsule/build"))

	tests := []struct {
		name      string
		namespace uuid.UUID
		input     string
		want      string
	}{
		{name: "rfc 9562 dns vector", namespace: nameSpaceDNS, input: "www.example.com", want: "2ed6657d-e927-568b-95e1-2665a8aea6a2"},
		{
			name:      "capsule build namespace",
			namespace: NameSpaceURL,
			input:     "https://github.com/lin-snow/Ech0/capsule/build",
			want:      "d2d5a2f6-2044-58fb-b6d7-72a102a3a72d",
		},
		{name: "derived tag id", namespace: capsuleNS, input: "tag\x00hello", want: "e81d7a70-2b4e-526e-823e-b8eb3677de82"},
		{name: "derived file id", namespace: capsuleNS, input: "file\x00a\x00b", want: "c20887ea-5e05-5dbd-9bfc-940bbf6ea167"},
		{name: "empty name", namespace: capsuleNS, input: "\x00", want: "edb90640-d667-5e35-ac18-c6b8cd8097c0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewV5(tt.namespace, []byte(tt.input)).String(); got != tt.want {
				t.Fatalf("NewV5(%s, %q) = %s, want %s", tt.namespace, tt.input, got, tt.want)
			}
		})
	}
}

func TestNewV5DistinctNamespaces(t *testing.T) {
	a := NewV5(NameSpaceURL, []byte("x"))
	b := NewV5(nameSpaceDNS, []byte("x"))
	if a == b {
		t.Fatalf("NewV5 ignored the namespace: both sides = %s", a)
	}
	if again := NewV5(NameSpaceURL, []byte("x")); again != a {
		t.Fatalf("NewV5 is not deterministic: %s then %s", a, again)
	}
}
