// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"fmt"
	"sort"
	"strings"
)

// apiFilesMarker 是 serve 模式的媒体路由前缀。它在胶囊里出现意味着内容里
// 硬编码了「某个活实例的文件接口」——目标实例上未必成立，故一律告警（spec §7）。
const apiFilesMarker = "/api/files/"

// instanceMarker 报告文本是否内嵌了源实例相关 URL，返回命中的标记。
// server_url 为空（手写胶囊常缺省）时退化为只查 /api/files/。
func instanceMarker(text, serverURL string) string {
	if text == "" {
		return ""
	}
	if base := strings.TrimRight(serverURL, "/"); base != "" && strings.Contains(text, base) {
		return base
	}
	if strings.Contains(text, apiFilesMarker) {
		return apiFilesMarker
	}
	return ""
}

// payloadHit 是 extension.payload 内一处内嵌实例 URL 的定位。
type payloadHit struct {
	field  string
	marker string
}

// scanPayload 递归扫描 payload 的全部字符串叶子。payload 结构随 type 而异、
// 规格不逐一约束（spec §4.2），所以只能无差别地扫，不能按已知键名去挑。
func scanPayload(payload map[string]any, serverURL string) []payloadHit {
	var hits []payloadHit
	scanValue("extension.payload", payload, serverURL, &hits)
	return hits
}

func scanValue(field string, v any, serverURL string, hits *[]payloadHit) {
	switch t := v.(type) {
	case string:
		if marker := instanceMarker(t, serverURL); marker != "" {
			*hits = append(*hits, payloadHit{field: field, marker: marker})
		}
	case map[string]any:
		// map 遍历顺序随机，排序后再下钻，报告才稳定。
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			scanValue(field+"."+k, t[k], serverURL, hits)
		}
	case []any:
		for i := range t {
			scanValue(fmt.Sprintf("%s[%d]", field, i), t[i], serverURL, hits)
		}
	}
}
