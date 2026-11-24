// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// recentCache is the stored summary together with what it was generated
// under. A summary is only served while the setting that shaped it is still
// the setting in force: switching the model, prompt or endpoint makes a new
// one, without any code having to remember to invalidate the old.
type recentCache struct {
	Fingerprint string `json:"fingerprint"`
	Text        string `json:"text"`
}

// recentFingerprint covers the setting fields that change what a summary says.
func recentFingerprint(setting settingModel.AgentSetting) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		setting.Protocol, setting.BaseURL, setting.Model, setting.Prompt,
	}, "\x00")))
	return hex.EncodeToString(sum[:8])
}

// GetRecent returns the public "what the author has been up to" summary.
//
// The enable switch is checked before the cache, not after: this endpoint is
// public, and an operator who turns the agent off expects the summary to stop
// being served, not to linger until the next post clears it.
func (s *CopilotService) GetRecent(ctx context.Context) (string, error) {
	const cacheKey = string(agent.GEN_RECENT)

	setting, err := s.agentSetting(ctx)
	if err != nil {
		return "", err
	}
	if !setting.Enable {
		return "", apperr.ErrAgentDisabled
	}
	fingerprint := recentFingerprint(setting)

	if value, ok := s.getRecentFromCache(ctx, cacheKey, fingerprint); ok {
		return value, nil
	}

	value, err, _ := s.recentGenGroup.Do(cacheKey+":"+fingerprint, func() (any, error) {
		if cached, ok := s.getRecentFromCache(ctx, cacheKey, fingerprint); ok {
			return cached, nil
		}

		output, err := s.buildRecentSummary(ctx, setting)
		if err != nil {
			return "", err
		}

		payload, _ := json.Marshal(recentCache{Fingerprint: fingerprint, Text: output})
		if err := s.durableKV.Set(ctx, cacheKey, string(payload)); err != nil {
			logUtil.GetLogger().
				Error("Failed to add or update key value", logUtil.Err(err))
		}

		return output, nil
	})
	if err != nil {
		return "", err
	}

	recent, ok := value.(string)
	if !ok {
		return "", errors.New("recent summary type assertion failed")
	}

	return recent, nil
}

func (s *CopilotService) getRecentFromCache(ctx context.Context, cacheKey, fingerprint string) (string, bool) {
	raw, err := s.durableKV.Get(ctx, cacheKey)
	if err != nil {
		return "", false
	}
	var cached recentCache
	if json.Unmarshal([]byte(raw), &cached) != nil || cached.Fingerprint != fingerprint {
		return "", false
	}
	return cached.Text, true
}

func (s *CopilotService) buildRecentSummary(ctx context.Context, setting settingModel.AgentSetting) (string, error) {
	systemCtx := viewer.WithContext(ctx, viewer.System)
	echos, err := s.echoService.GetEchosByPage(
		systemCtx,
		commonModel.PageQueryDto{
			Page:     1,
			PageSize: 10,
		},
	)
	if err != nil {
		return "", err
	}

	var memos []agent.Message
	for i, e := range echos.Items {
		content := fmt.Sprintf(
			"用户 %s 在 %s 发布了内容 %d ：%s 。 内容标签为：%v。",
			e.Username,
			time.Unix(e.CreatedAt, 0).UTC().Format("2006-01-02 15:04"),
			i+1,
			e.Content,
			e.Tags,
		)

		memos = append(memos, agent.Message{
			Role:    agent.RoleUser,
			Content: content,
		})
	}

	locale := i18n.SystemDefaultLocale()
	in := []agent.Message{
		{
			Role:    agent.RoleSystem,
			Content: summarySystemPromptFor(locale),
		},
		{
			Role:    agent.RoleUser,
			Content: summaryUserPromptFor(locale),
		},
	}

	in = append(in, memos...)

	output, err := agent.Generate(ctx, setting, in, true)
	if err != nil {
		return "", err
	}

	return output, nil
}
