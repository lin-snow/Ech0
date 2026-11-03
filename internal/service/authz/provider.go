// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAuthzService,
	wire.Bind(new(Service), new(*AuthzService)),
)
