// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package i18n resolves which language a request is in.
//
// The server does not produce user-facing text: API error details are English
// and the SPA owns every translated string. The resolved locale is still
// needed for two things that genuinely happen server-side: validating the
// configured default locale and telling the copilot which language to answer
// in.
package i18n

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/database"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"golang.org/x/text/language"
)

const ContextLocaleKey = "locale"

var (
	supportedLocales = []language.Tag{
		language.MustParse("zh-CN"),
		language.MustParse("en-US"),
		language.MustParse("de-DE"),
		language.MustParse("ja-JP"),
	}
	matcher = language.NewMatcher(supportedLocales)
)

const maxAcceptLanguageSeparators = 32

func sanitizeAcceptLanguage(v string) string {
	if strings.Count(v, "-")+strings.Count(v, "_") > maxAcceptLanguageSeparators {
		return ""
	}
	return v
}

func ResolveLocale(raw ...string) string {
	parts := make([]string, 0, len(raw))
	for _, v := range raw {
		v = sanitizeAcceptLanguage(strings.TrimSpace(v))
		if v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return string(commonModel.FallbackLocale)
	}
	tag, _, _ := language.ParseAcceptLanguage(strings.Join(parts, ","))
	best, _, confidence := matcher.Match(tag...)
	if confidence == language.No {
		return string(commonModel.FallbackLocale)
	}
	base, _ := best.Base()
	region, _ := best.Region()
	if region.IsCountry() {
		return fmt.Sprintf("%s-%s", strings.ToLower(base.String()), strings.ToUpper(region.String()))
	}
	return best.String()
}

func LocaleFromGin(ctx *gin.Context) string {
	if ctx == nil {
		return "zh-CN"
	}
	if v, ok := ctx.Get(ContextLocaleKey); ok {
		if locale, ok := v.(string); ok && locale != "" {
			return locale
		}
	}
	return "zh-CN"
}

func hasExplicitLocale(ctx *gin.Context) bool {
	if ctx == nil {
		return false
	}
	if strings.TrimSpace(ctx.Query("lang")) != "" {
		return true
	}
	return strings.TrimSpace(ctx.GetHeader("X-Locale")) != ""
}

func explicitLocaleFromRequest(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}
	explicit := strings.TrimSpace(ctx.Query("lang"))
	if explicit != "" {
		return explicit
	}
	return strings.TrimSpace(ctx.GetHeader("X-Locale"))
}

// defaultLocaleProvider is installed by the composition root, which resolves
// the instance default locale through the settings kernel. Keeping it a hook
// means this package never reads the settings table itself.
var defaultLocaleProvider func() string

// SetDefaultLocaleProvider installs the default-locale reader. Pass nil to
// restore the compiled-in fallback (tests).
func SetDefaultLocaleProvider(fn func() string) { defaultLocaleProvider = fn }

func SystemDefaultLocale() string { return systemDefaultLocale() }

func systemDefaultLocale() string {
	if defaultLocaleProvider != nil {
		if locale := strings.TrimSpace(defaultLocaleProvider()); locale != "" {
			return ResolveLocale(locale)
		}
	}
	return string(commonModel.DefaultLocale)
}

func userPreferredLocale(userID string) string {
	if strings.TrimSpace(userID) == "" {
		return ""
	}
	db := database.GetDB()
	if db == nil {
		return ""
	}

	var user userModel.User
	if err := db.Select("locale").Where("id = ?", userID).First(&user).Error; err != nil {
		return ""
	}
	if strings.TrimSpace(user.Locale) == "" {
		return ""
	}
	return ResolveLocale(user.Locale)
}

func setLocaleContext(ctx *gin.Context, locale string) {
	if ctx == nil {
		return
	}
	ctx.Set(ContextLocaleKey, ResolveLocale(locale))
}

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		explicit := explicitLocaleFromRequest(ctx)
		acceptLanguage := strings.TrimSpace(ctx.GetHeader("Accept-Language"))
		locale := systemDefaultLocale()
		if explicit != "" {
			locale = ResolveLocale(explicit, acceptLanguage)
		}
		setLocaleContext(ctx, locale)
		ctx.Next()
	}
}

func ApplyUserLocaleFromUserID(ctx *gin.Context, userID string) {
	if ctx == nil || hasExplicitLocale(ctx) {
		return
	}
	locale := userPreferredLocale(userID)
	if strings.TrimSpace(locale) == "" {
		locale = systemDefaultLocale()
	}
	setLocaleContext(ctx, locale)
}
