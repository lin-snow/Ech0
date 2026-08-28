// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package setting

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/kvstore"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
)

var (
	System = Spec[settingModel.SystemSetting]{
		Key: commonModel.SystemSettingsKey,
		Default: func() settingModel.SystemSetting {
			c := config.Config().Setting
			return settingModel.SystemSetting{
				SiteTitle:     c.SiteTitle,
				ServerLogo:    c.ServerLogo,
				ServerName:    c.Servername,
				ServerURL:     urlUtil.TrimURL(c.Serverurl),
				HomeLayout:    "single",
				AllowRegister: c.AllowRegister,
				DefaultLocale: string(commonModel.DefaultLocale),
				ICPNumber:     c.Icpnumber,
				FooterContent: c.FooterContent,
				FooterLink:    urlUtil.TrimURL(c.FooterLink),
				MetingAPI:     urlUtil.TrimURL(c.MetingAPI),
				CustomCSS:     c.CustomCSS,
				CustomJS:      c.CustomJS,
			}
		},
		Normalize: func(s *settingModel.SystemSetting) {
			s.DefaultLocale = i18nUtil.ResolveLocale(s.DefaultLocale)
			if s.HomeLayout != "three" {
				s.HomeLayout = "single"
			}
		},
	}

	OAuth2 = Spec[settingModel.OAuth2Setting]{
		Key: commonModel.OAuth2SettingKey,
		Default: func() settingModel.OAuth2Setting {
			return settingModel.OAuth2Setting{
				Enable:                        false,
				Provider:                      string(commonModel.OAuth2GITHUB),
				AuthURL:                       "https://github.com/login/oauth/authorize",
				TokenURL:                      "https://github.com/login/oauth/access_token",
				UserInfoURL:                   "https://api.github.com/user",
				Scopes:                        []string{"read:user"},
				AuthRedirectAllowedReturnURLs: append([]string{}, config.Config().Auth.Redirect.AllowedReturnURLs...),
				CORSAllowedOrigins:            append([]string{}, config.Config().Web.CORS.AllowedOrigins...),
			}
		},
		Normalize: normalizeOAuth2Boundary,
	}

	S3 = Spec[settingModel.S3Setting]{
		Key: commonModel.S3SettingKey,
		Default: func() settingModel.S3Setting {
			c := config.Config().Storage
			return settingModel.S3Setting{
				Enable:       c.ObjectEnabled,
				Provider:     strings.TrimSpace(c.Provider),
				Endpoint:     stripScheme(strings.TrimSpace(c.Endpoint)),
				AccessKey:    c.AccessKey,
				SecretKey:    c.SecretKey,
				BucketName:   c.BucketName,
				Region:       strings.TrimSpace(c.Region),
				UseSSL:       c.UseSSL,
				CDNURL:       strings.TrimRight(strings.TrimSpace(c.CDNURL), "/"),
				PathPrefix:   strings.Trim(strings.TrimSpace(c.PathPrefix), "/"),
				PublicRead:   true,
				UsePathStyle: c.UsePathStyle,
			}
		},
	}

	Passkey = Spec[settingModel.PasskeySetting]{
		Key: commonModel.PasskeySettingKey,
		Default: func() settingModel.PasskeySetting {
			return settingModel.PasskeySetting{
				WebAuthnRPID:           strings.TrimSpace(config.Config().Auth.WebAuthn.RPID),
				WebAuthnAllowedOrigins: append([]string{}, config.Config().Auth.WebAuthn.Origins...),
			}
		},
		Normalize: normalizePasskeyBoundary,
		Migrate:   migratePasskeyFromLegacy,
	}

	Agent = Spec[settingModel.AgentSetting]{
		Key: commonModel.AgentSettingKey,
		Default: func() settingModel.AgentSetting {
			return settingModel.AgentSetting{
				Enable:   false,
				Protocol: string(commonModel.OpenAI),
			}
		},
	}

	Snapshot = Spec[settingModel.SnapshotSchedule]{
		Key: commonModel.SnapshotScheduleKey,
		Default: func() settingModel.SnapshotSchedule {
			return settingModel.SnapshotSchedule{
				Enable:         false,
				CronExpression: "0 2 * * 0",
			}
		},
	}

	Embedding = Spec[settingModel.EmbeddingSetting]{
		Key: commonModel.EmbeddingSettingKey,
		Default: func() settingModel.EmbeddingSetting {
			return settingModel.EmbeddingSetting{Enable: false}
		},
	}

	Comment = Spec[commentModel.SystemSetting]{
		Key: commentModel.CommentSystemSettingKey,
		Default: func() commentModel.SystemSetting {
			s := commentModel.SystemSetting{
				EnableComment:   true,
				RequireApproval: true,
				CaptchaEnabled:  false,
			}
			normalizeComment(&s)
			return s
		},
		Normalize: normalizeComment,
	}
)

var registry = []seedable{
	System,
	serverURLSeed{},
	OAuth2,
	S3,
	Passkey,
	Agent,
	Snapshot,
	Embedding,
	Comment,
}

func normalizeOAuth2Boundary(s *settingModel.OAuth2Setting) {
	if len(s.AuthRedirectAllowedReturnURLs) == 0 {
		s.AuthRedirectAllowedReturnURLs = append([]string{}, config.Config().Auth.Redirect.AllowedReturnURLs...)
	}
	if len(s.CORSAllowedOrigins) == 0 {
		s.CORSAllowedOrigins = append([]string{}, config.Config().Web.CORS.AllowedOrigins...)
	}
}

func normalizePasskeyBoundary(s *settingModel.PasskeySetting) {
	if strings.TrimSpace(s.WebAuthnRPID) == "" {
		s.WebAuthnRPID = strings.TrimSpace(config.Config().Auth.WebAuthn.RPID)
	}
	if len(s.WebAuthnAllowedOrigins) == 0 {
		s.WebAuthnAllowedOrigins = append([]string{}, config.Config().Auth.WebAuthn.Origins...)
	}
}

func normalizeComment(s *commentModel.SystemSetting) {
	if s.EmailNotify.SMTPPort <= 0 {
		s.EmailNotify.SMTPPort = 587
	}
}

func migratePasskeyFromLegacy(ctx context.Context, kv kvstore.Store) (settingModel.PasskeySetting, bool) {
	var result settingModel.PasskeySetting
	raw, err := kv.Get(ctx, commonModel.OAuth2SettingKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return result, false
	}
	var legacy struct {
		WebAuthnRPID           string   `json:"webauthn_rp_id"`
		WebAuthnAllowedOrigins []string `json:"webauthn_allowed_origins"`
	}
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return result, false
	}
	result.WebAuthnRPID = strings.TrimSpace(legacy.WebAuthnRPID)
	result.WebAuthnAllowedOrigins = sanitizeURLList(legacy.WebAuthnAllowedOrigins)
	return result, result.WebAuthnRPID != "" || len(result.WebAuthnAllowedOrigins) > 0
}

type serverURLSeed struct{}

func (serverURLSeed) seed(ctx context.Context, kv kvstore.Store) error {
	if _, err := kv.Get(ctx, commonModel.ServerURLKey); err == nil {
		return nil
	} else if !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}
	sys := System.Default()
	if System.Normalize != nil {
		System.Normalize(&sys)
	}
	return kv.Set(ctx, commonModel.ServerURLKey, sys.ServerURL)
}

func stripScheme(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, "http://"), "https://")
}

func sanitizeURLList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := urlUtil.TrimURL(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
