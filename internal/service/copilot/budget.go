// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import settingModel "github.com/lin-snow/ech0/internal/model/setting"

const (
	defaultContextWindow = 256_000
	contextUsableRatio   = 0.6
	contextReserveTokens = 8_000
	minAggregateBudget   = 2_000
)

func aggregateBudgetTokens(setting settingModel.AgentSetting) int {
	window := setting.ContextWindow
	if window <= 0 {
		window = defaultContextWindow
	}
	budget := int(float64(window)*contextUsableRatio) - contextReserveTokens
	if budget < minAggregateBudget {
		return minAggregateBudget
	}
	return budget
}

func chatContextBudgetTokens(setting settingModel.AgentSetting) int {
	return aggregateBudgetTokens(setting)
}
