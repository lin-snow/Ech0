// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"reflect"
)

type MetadataBuilder func(PublishMetadataInput) map[string]string

type PublishMetadataInput struct {
	Context   context.Context
	EventType reflect.Type
	Topic     string
	Key       string
	Headers   map[string]string
	Value     any
}
