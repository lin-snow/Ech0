// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"bytes"
	"context"
	"io"
	"net/http"
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

func TestValidateMetadataRedirectOrigins(t *testing.T) {
	const clientID = "https://client.example.com/client.json"
	cases := []struct {
		name      string
		redirects []string
		wantErr   bool
	}{
		{"same origin", []string{"https://client.example.com/callback"}, false},
		{"same origin explicit default port", []string{"https://client.example.com:443/callback"}, false},
		{"same origin non-default port", []string{"https://client.example.com:8443/callback"}, true},
		{"loopback exception", []string{"http://127.0.0.1:52123/callback"}, false},
		{"ipv6 loopback exception", []string{"http://[::1]:52123/callback"}, false},
		{"cross origin", []string{"https://evil.example.com/callback"}, true},
		{"sibling subdomain", []string{"https://app.client.example.com/callback"}, true},
		{"one bad redirect among good", []string{
			"https://client.example.com/callback",
			"https://evil.example.com/callback",
		}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMetadataRedirectOrigins(clientID, tc.redirects)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateMetadataRedirectOrigins() error = %v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestFetchClientMetadataRequiresHTTPSAndJSON(t *testing.T) {
	const clientID = "https://client.example.com/client.json"
	body := []byte(`{"client_id":"https://client.example.com/client.json","redirect_uris":["https://client.example.com/callback"]}`)

	cases := []struct {
		name        string
		contentType string
		finalScheme string
		wantErr     bool
	}{
		{"json over https", "application/json; charset=utf-8", "https", false},
		{"problem+json over https", "application/problem+json", "https", false},
		{"html content", "text/html", "https", true},
		{"missing content type", "", "https", true},
		{"https downgraded after redirect", "application/json", "http", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, _, _, _ := testService()
			service.http = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				final := req
				if tc.finalScheme != "https" {
					final = req.Clone(req.Context())
					final.URL.Scheme = tc.finalScheme
				}
				header := make(http.Header)
				if tc.contentType != "" {
					header.Set("Content-Type", tc.contentType)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewReader(body)),
					Request:    final,
				}, nil
			})}

			_, oauthErr := service.fetchClientMetadata(context.Background(), clientID)
			if (oauthErr != nil) != tc.wantErr {
				t.Fatalf("fetchClientMetadata() error = %v, wantErr=%v", oauthErr, tc.wantErr)
			}
		})
	}
}
