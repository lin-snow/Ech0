// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"fmt"
	"strings"
	"time"
)

// 时间是胶囊中唯一的表示层差异（spec §11）：库里是 int64 Unix 秒，胶囊里是
// RFC3339。二者无损双射——语义都是「时刻」，不携带时区意图。

// FormatUnix 把库中的 Unix 秒渲染成导出用的 RFC3339 UTC（Z 后缀）。
func FormatUnix(sec int64) string {
	return time.Unix(sec, 0).UTC().Format(time.RFC3339)
}

// ParseTime 解析 RFC3339 时间为 Unix 秒。输入接受任意合法偏移（手写胶囊常写
// +08:00），语义为时刻，故偏移信息在转成 Unix 秒后自然丢弃且无损。
func ParseTime(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty timestamp")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, fmt.Errorf("invalid RFC3339 timestamp %q: %w", raw, err)
	}
	return t.Unix(), nil
}
