// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// AddConnectDto is the request body for adding a peer. The record id is
// server-managed, so it is deliberately absent from the input schema.
type AddConnectDto struct {
	ConnectURL string `json:"connect_url" required:"true" doc:"Peer instance root URL; the server calls <url>/api/connect"`
}
