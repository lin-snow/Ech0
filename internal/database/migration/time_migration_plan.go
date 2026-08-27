// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

type TimeColumnPlan struct {
	Table  string
	Column string
}

type TimeMigrationStat struct {
	Table        string
	Column       string
	CandidateRow int64
	UpdatedRow   int64
}

type TimeMigrationInvalidSample struct {
	Table  string
	Column string
	RowID  int64
	Value  string
}

func StorageTimeColumnPlans() []TimeColumnPlan {
	return []TimeColumnPlan{
		{Table: "user_local_auth", Column: "updated_at"},
		{Table: "user_external_identities", Column: "created_at"},
		{Table: "user_external_identities", Column: "updated_at"},
		{Table: "webauthn_credentials", Column: "last_used_at"},
		{Table: "webauthn_credentials", Column: "created_at"},
		{Table: "webauthn_credentials", Column: "updated_at"},
		{Table: "echos", Column: "created_at"},
		{Table: "echo_extensions", Column: "created_at"},
		{Table: "echo_extensions", Column: "updated_at"},
		{Table: "tags", Column: "created_at"},
		{Table: "files", Column: "created_at"},
		{Table: "temp_files", Column: "expire_at"},
		{Table: "temp_files", Column: "created_at"},
		{Table: "comments", Column: "created_at"},
		{Table: "comments", Column: "updated_at"},
		{Table: "webhooks", Column: "last_trigger"},
		{Table: "webhooks", Column: "created_at"},
		{Table: "webhooks", Column: "updated_at"},
		{Table: "access_token_settings", Column: "expiry"},
		{Table: "access_token_settings", Column: "last_used_at"},
		{Table: "access_token_settings", Column: "created_at"},
		{Table: "passkeys", Column: "last_used_at"},
		{Table: "passkeys", Column: "created_at"},
		{Table: "passkeys", Column: "updated_at"},
	}
}

func StorageTimeColumnsByTable() map[string][]string {
	tableColumns := make(map[string][]string)
	for _, plan := range StorageTimeColumnPlans() {
		tableColumns[plan.Table] = append(tableColumns[plan.Table], plan.Column)
	}
	return tableColumns
}
