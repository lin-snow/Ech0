// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// Delivery is a recoverable side effect. ID is also the recipient's replay key.
// A separate lane serializes each webhook and the shared embedding index.
type Delivery struct {
	ID          string `gorm:"primaryKey"`
	Lane        string `gorm:"not null;index:idx_delivery_lane"`
	Subject     string
	Destination string
	Payload     []byte
	Ready       int64  `gorm:"not null;index"`
	Attempt     int64  `gorm:"not null"`
	Lease       string `gorm:"not null"`
}

func (Delivery) TableName() string { return "ech0_deliveries" }
