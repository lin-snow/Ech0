package auth

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
)

func TestParseAndValidateClientRedirect_Allowed(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	u, err := svc.parseAndValidateClientRedirect("https://app.example.com/auth?from=test")
	if err != nil {
		t.Fatalf("expected allow redirect, got err: %v", err)
	}
	if u.Host != "app.example.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
}

func TestParseAndValidateClientRedirect_Denied(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	_, err := svc.parseAndValidateClientRedirect("https://evil.example.net/auth")
	if err == nil {
		t.Fatalf("expected deny redirect")
	}
}
