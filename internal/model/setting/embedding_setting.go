// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type EmbeddingSetting struct {
	Enable    bool   `json:"enable"`
	Model     string `json:"model"`
	ApiKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Dim       int    `json:"dim"`
	BatchSize int    `json:"batch_size"`
}

type EmbeddingSettingDto struct {
	Enable    bool   `json:"enable"`
	Model     string `json:"model"`
	ApiKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Dim       int    `json:"dim"`
	BatchSize int    `json:"batch_size"`
}
