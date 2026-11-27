// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// List is the single list payload of the API. Offset pagination fills Total;
// cursor pagination fills NextCursor. A client tells the two apart by which
// field the response carries, never by out-of-band knowledge.
type List[T any] struct {
	Items      []T    `json:"items"                doc:"当前页数据"`
	Total      *int64 `json:"total,omitempty"      doc:"符合查询条件的总条数，仅偏移分页返回"`
	NextCursor string `json:"nextCursor,omitempty" doc:"下一页游标，仅游标分页返回；为空表示没有更多"`
}
