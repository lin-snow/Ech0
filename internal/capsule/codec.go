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

const unknownFieldMarker = "not found in type"

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

func newLaxDecoder(data []byte) *yaml.Decoder {
	return yaml.NewDecoder(bytes.NewReader(data))
}

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

	if err := newLaxDecoder(data).Decode(out); err != nil {
		return unknown, err
	}
	return unknown, nil
}

var ErrEmptyDocument = errors.New("capsule: empty yaml document")

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

func splitFrontmatter(data []byte) (front, body []byte, ok bool) {
	rest, ok := bytes.CutPrefix(data, []byte(frontmatterFence+"\n"))
	if !ok {
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
	if trimmed, cut := bytes.CutSuffix(rest, []byte("\n"+frontmatterFence)); cut {
		return append(trimmed, '\n'), nil, true
	}
	return nil, nil, false
}
