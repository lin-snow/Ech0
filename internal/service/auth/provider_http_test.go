// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func swapOIDCClient(t *testing.T, c *http.Client) {
	t.Helper()
	old := oidcHTTPClient
	oidcHTTPClient = c
	t.Cleanup(func() { oidcHTTPClient = old })
}

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func cannedResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func TestExchangeOAuthCode_Direct(t *testing.T) {
	t.Run("success returns token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{"access_token":"acc-tok","token_type":"Bearer","expires_in":3600}`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
		setting.TokenURL = ts.URL
		tok, err := exchangeOAuthCode(&setting, "code-1")
		require.NoError(t, err)
		assert.Equal(t, "acc-tok", tok.AccessToken)
	})

	t.Run("server error returns error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "bad request", http.StatusBadRequest)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
		setting.TokenURL = ts.URL
		_, err := exchangeOAuthCode(&setting, "code-1")
		require.Error(t, err)
	})
}

func TestGitHubProvider(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, `{"access_token":"gh-token","token_type":"bearer","scope":"read:user"}`)
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer gh-token", r.Header.Get("Authorization"))
		writeJSON(w, http.StatusOK, `{"id":12345,"login":"octocat","email":"octo@example.com"}`)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	swapOIDCClient(t, ts.Client())

	setting := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
	setting.TokenURL = ts.URL + "/token"
	setting.UserInfoURL = ts.URL + "/userinfo"

	tokResp, err := exchangeGithubCodeForToken(&setting, "code")
	require.NoError(t, err)
	assert.Equal(t, "gh-token", tokResp.AccessToken)
	assert.Contains(t, tokResp.Scope, "read:user")

	user, err := fetchGitHubUserInfo(&setting, "gh-token")
	require.NoError(t, err)
	assert.Equal(t, int64(12345), user.ID)
	assert.Equal(t, "octocat", user.Login)
}

func TestFetchGitHubUserInfo_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()
	swapOIDCClient(t, ts.Client())

	setting := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
	setting.UserInfoURL = ts.URL
	_, err := fetchGitHubUserInfo(&setting, "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub")
}

func TestGoogleProvider(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK,
			`{"access_token":"g-token","token_type":"Bearer","expires_in":3600,"id_token":"g-id-token","scope":"openid email"}`)
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, `{"sub":"g-sub","email":"g@example.com","verified_email":true}`)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	swapOIDCClient(t, ts.Client())

	setting := fullOAuth2Setting(string(commonModel.OAuth2GOOGLE))
	setting.TokenURL = ts.URL + "/token"
	setting.UserInfoURL = ts.URL + "/userinfo"

	tokResp, err := exchangeGoogleCodeForToken(&setting, "code")
	require.NoError(t, err)
	assert.Equal(t, "g-token", tokResp.AccessToken)
	assert.Equal(t, "g-id-token", tokResp.IDToken)
	assert.Greater(t, tokResp.ExpiresIn, int64(0))

	user, err := fetchGoogleUserInfo(&setting, "g-token")
	require.NoError(t, err)
	assert.Equal(t, "g-sub", user.Sub)
}

func TestFetchGoogleUserInfo_Errors(t *testing.T) {
	t.Run("non-200", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusBadGateway)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := fullOAuth2Setting(string(commonModel.OAuth2GOOGLE))
		setting.UserInfoURL = ts.URL
		_, err := fetchGoogleUserInfo(&setting, "tok")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Google")
	})

	t.Run("invalid json body", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{not-json`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := fullOAuth2Setting(string(commonModel.OAuth2GOOGLE))
		setting.UserInfoURL = ts.URL
		_, err := fetchGoogleUserInfo(&setting, "tok")
		require.Error(t, err)
	})
}

func TestExchangeQQCodeForToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/jsonp", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `callback({"access_token":"qq-tok","expires_in":7776000,"openid":"qq-open"});`)
	})
	mux.HandleFunc("/query", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `access_token=qq-tok2&expires_in=100&refresh_token=rt&openid=qq-open2`)
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	})
	mux.HandleFunc("/garbage", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `total-garbage-no-token`)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	swapOIDCClient(t, ts.Client())

	base := fullOAuth2Setting(string(commonModel.OAuth2QQ))

	t.Run("jsonp wrapped json", func(t *testing.T) {
		setting := base
		setting.TokenURL = ts.URL + "/jsonp"
		resp, err := exchangeQQCodeForToken(&setting, "code")
		require.NoError(t, err)
		assert.Equal(t, "qq-tok", resp.AccessToken)
		assert.Equal(t, "qq-open", resp.OpenID)
	})

	t.Run("query string fallback", func(t *testing.T) {
		setting := base
		setting.TokenURL = ts.URL + "/query"
		resp, err := exchangeQQCodeForToken(&setting, "code")
		require.NoError(t, err)
		assert.Equal(t, "qq-tok2", resp.AccessToken)
		assert.Equal(t, "qq-open2", resp.OpenID)
		assert.Equal(t, "rt", resp.RefreshToken)
		assert.Equal(t, int64(100), resp.ExpiresIn)
	})

	t.Run("non-200 errors", func(t *testing.T) {
		setting := base
		setting.TokenURL = ts.URL + "/error"
		_, err := exchangeQQCodeForToken(&setting, "code")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "QQ token")
	})

	t.Run("unparseable body errors", func(t *testing.T) {
		setting := base
		setting.TokenURL = ts.URL + "/garbage"
		_, err := exchangeQQCodeForToken(&setting, "code")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "无法解析")
	})
}

func TestFetchQQUserInfo(t *testing.T) {
	t.Run("success parses openid", func(t *testing.T) {
		swapOIDCClient(t, &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "graph.qq.com", r.URL.Host)
			return cannedResponse(http.StatusOK, `{"client_id":"appid","openid":"qq-openid-xyz"}`), nil
		})})

		resp, err := fetchQQUserInfo("qq-tok")
		require.NoError(t, err)
		assert.Equal(t, "qq-openid-xyz", resp.OpenID)
	})

	t.Run("non-200 errors", func(t *testing.T) {
		swapOIDCClient(t, &http.Client{Transport: rtFunc(func(_ *http.Request) (*http.Response, error) {
			return cannedResponse(http.StatusInternalServerError, "boom"), nil
		})})

		_, err := fetchQQUserInfo("qq-tok")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "QQ openid")
	})
}

func TestExchangeCustomCodeForToken(t *testing.T) {
	t.Run("non-oidc returns access token only", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{"access_token":"c-tok","token_type":"Bearer"}`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
		setting.IsOIDC = false
		setting.TokenURL = ts.URL
		acc, idt, err := exchangeCustomCodeForToken(&setting, "code")
		require.NoError(t, err)
		assert.Equal(t, "c-tok", acc)
		assert.Empty(t, idt)
	})

	t.Run("oidc returns access and id token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{"access_token":"c-tok","token_type":"Bearer","id_token":"c-idtok"}`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
		setting.IsOIDC = true
		setting.TokenURL = ts.URL
		acc, idt, err := exchangeCustomCodeForToken(&setting, "code")
		require.NoError(t, err)
		assert.Equal(t, "c-tok", acc)
		assert.Equal(t, "c-idtok", idt)
	})
}

func TestFetchCustomUserInfo_NonOIDC(t *testing.T) {
	newSetting := func() settingModel.OAuth2Setting {
		s := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
		s.IsOIDC = false
		return s
	}

	t.Run("extracts sub field", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer acc-tok", r.Header.Get("Authorization"))
			writeJSON(w, http.StatusOK, `{"sub":"custom-sub","email":"c@example.com"}`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := newSetting()
		setting.UserInfoURL = ts.URL
		id, err := fetchCustomUserInfo(&setting, "acc-tok", "", "")
		require.NoError(t, err)
		assert.Equal(t, "custom-sub", id)
	})

	t.Run("id field takes priority", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{"id":98765,"sub":"ignored"}`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := newSetting()
		setting.UserInfoURL = ts.URL
		id, err := fetchCustomUserInfo(&setting, "acc-tok", "", "")
		require.NoError(t, err)
		assert.Equal(t, "98765", id)
	})

	t.Run("missing identity field errors", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{"name":"bob"}`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := newSetting()
		setting.UserInfoURL = ts.URL
		_, err := fetchCustomUserInfo(&setting, "acc-tok", "", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "唯一标识")
	})

	t.Run("non-200 errors", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := newSetting()
		setting.UserInfoURL = ts.URL
		_, err := fetchCustomUserInfo(&setting, "acc-tok", "", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Custom")
	})

	t.Run("invalid json body errors", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, `{broken`)
		}))
		defer ts.Close()
		swapOIDCClient(t, ts.Client())

		setting := newSetting()
		setting.UserInfoURL = ts.URL
		_, err := fetchCustomUserInfo(&setting, "acc-tok", "", "")
		require.Error(t, err)
	})

	t.Run("oidc with empty id token errors", func(t *testing.T) {
		setting := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
		setting.IsOIDC = true
		_, err := fetchCustomUserInfo(&setting, "acc-tok", "", "nonce")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id_token")
	})
}
