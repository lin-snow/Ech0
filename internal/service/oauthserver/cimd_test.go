// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"testing"
)

func TestValidateMetadataDocumentURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"public https", "https://client.example.com/oauth/metadata.json", false},
		{"public ip https", "https://93.184.216.34/metadata.json", false},
		{"http rejected", "http://client.example.com/metadata.json", true},
		{"loopback rejected", "https://127.0.0.1/metadata.json", true},
		{"localhost rejected", "https://localhost/metadata.json", true},
		{"metadata ip rejected", "https://169.254.169.254/latest/meta-data", true},
		{"private ip rejected", "https://10.0.0.8/metadata.json", true},
		{"userinfo rejected", "https://user:pass@client.example.com/metadata.json", true},
		{"fragment rejected", "https://client.example.com/metadata.json#frag", true},
		{"root path rejected", "https://client.example.com/", true},
		{"no path rejected", "https://client.example.com", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMetadataDocumentURL(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateMetadataDocumentURL(%q) error = %v, wantErr=%v", tc.raw, err, tc.wantErr)
			}
		})
	}
}
