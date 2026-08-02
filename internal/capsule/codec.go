// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	frontmatterFence = "---"
	yamlIndent       = 2
)

// unknownFieldMarker 是 yaml.v3 严格解码对「结构体没有这个字段」的措辞。
// 未知字段按 spec §8 属警告（前向兼容），其余类型错误才是硬错误。
const unknownFieldMarker = "not found in type"

// EncodeYAML 以两空格缩进序列化，输出稳定、diff 友好。
func EncodeYAML(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(yamlIndent)
	if err := enc.Encode(v); err != nil {
		_ = enc.Close()
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// newLaxDecoder 返回一个忽略未知字段的解码器。
func newLaxDecoder(data []byte) *yaml.Decoder {
	return yaml.NewDecoder(bytes.NewReader(data))
}

// DecodeYAML 解析 YAML 并把未知字段降级为警告返回，其余错误照常上抛。
// 消费者必须忽略未知字段（spec §8），但 check 需要能报出来。
func DecodeYAML(data []byte, out any) (unknown []string, err error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	err = dec.Decode(out)
	if err == nil {
		return nil, nil
	}
	if errors.Is(err, io.EOF) {
		return nil, ErrEmptyDocument
	}

	var typeErr *yaml.TypeError
	if !errors.As(err, &typeErr) {
		return nil, err
	}

	var fatal []string
	for _, e := range typeErr.Errors {
		if strings.Contains(e, unknownFieldMarker) {
			unknown = append(unknown, e)
			continue
		}
		fatal = append(fatal, e)
	}
	if len(fatal) > 0 {
		return unknown, fmt.Errorf("yaml: %s", strings.Join(fatal, "; "))
	}

	// 只有未知字段：宽松地重解一次，拿到可用的值。
	if err := newLaxDecoder(data).Decode(out); err != nil {
		return unknown, err
	}
	return unknown, nil
}

// ErrEmptyDocument 标记空的 YAML 文档（yaml.v3 用 io.EOF 表达，语义太弱）。
var ErrEmptyDocument = errors.New("capsule: empty yaml document")

// EncodeEcho 把一个 Echo 渲染成 frontmatter-markdown 字节。
// 正文逐字写出，不做任何转换或转义（spec §4.3）。
func EncodeEcho(doc *EchoDoc) ([]byte, error) {
	front, err := EncodeYAML(doc)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Grow(len(front) + len(doc.Content) + 2*len(frontmatterFence) + 4)
	buf.WriteString(frontmatterFence + "\n")
	buf.Write(front)
	buf.WriteString(frontmatterFence + "\n")
	buf.WriteString(doc.Content)
	return buf.Bytes(), nil
}

// DecodeEcho 解析 frontmatter-markdown。闭合围栏之后的全部字节即正文，
// 逐字保留（含首尾空白与换行）。
func DecodeEcho(data []byte) (*EchoDoc, []string, error) {
	body, rest, ok := splitFrontmatter(data)
	if !ok {
		return nil, nil, errors.New("missing frontmatter: file must start with a --- fence")
	}
	doc := &EchoDoc{}
	unknown, err := DecodeYAML(body, doc)
	if err != nil {
		return nil, unknown, err
	}
	doc.Content = string(rest)
	return doc, unknown, nil
}

// splitFrontmatter 切出围栏之间的 YAML 与其后的正文。
func splitFrontmatter(data []byte) (front, body []byte, ok bool) {
	rest, ok := bytes.CutPrefix(data, []byte(frontmatterFence+"\n"))
	if !ok {
		// 允许 CRLF 开头的手写文件。
		rest, ok = bytes.CutPrefix(data, []byte(frontmatterFence+"\r\n"))
		if !ok {
			return nil, nil, false
		}
	}
	for _, closer := range []string{"\n" + frontmatterFence + "\n", "\n" + frontmatterFence + "\r\n"} {
		if idx := bytes.Index(rest, []byte(closer)); idx >= 0 {
			return rest[:idx+1], rest[idx+len(closer):], true
		}
	}
	// 无正文时文件可以以围栏结尾且不带换行。
	if trimmed, cut := bytes.CutSuffix(rest, []byte("\n"+frontmatterFence)); cut {
		return append(trimmed, '\n'), nil, true
	}
	return nil, nil, false
}
