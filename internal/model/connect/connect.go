// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type Connect struct {
	ServerName  string `json:"server_name"`
	ServerURL   string `json:"server_url"`
	Logo        string `json:"logo"`
	TotalEchos  int    `json:"total_echos"`
	TodayEchos  int    `json:"today_echos"`
	SysUsername string `json:"sys_username"`
	Version     string `json:"version"`
}

type Connected struct {
	ID         string `gorm:"type:char(36);primaryKey" json:"id"`
	ConnectURL string `                  json:"connect_url"`
}

type ConnectedHealth struct {
	ID         string `json:"id"`
	ConnectURL string `json:"connect_url"`
	Status     string `json:"status"`
	Version    string `json:"version"`
}

func (c *Connected) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuidUtil.NewV7()
	}
	return nil
}
