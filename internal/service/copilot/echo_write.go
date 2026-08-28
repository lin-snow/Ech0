// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

// How much of an Echo a confirmation shows. Enough to recognise which one is
// about to change, short enough that the picker stays a picker.
const maxConfirmContentRunes = 240

// The three write tools all take content the same way, so they share one shape.
// Pointers on update: a partial edit has to tell "leave it alone" apart from
// "set it to empty", and a bool cannot do that.
type createEchoArgs struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	Private bool     `json:"private"`
}

type updateEchoArgs struct {
	ID      string    `json:"id"`
	Content *string   `json:"content"`
	Tags    *[]string `json:"tags"`
	Private *bool     `json:"private"`
}

type deleteEchoArgs struct {
	ID string `json:"id"`
}

// createEchoTool posts a new Echo, once the person has seen it and said yes.
func (s *CopilotService) createEchoTool(a *asker, locale string, loc *time.Location) agent.Tool {
	ws := writeStringsFor(locale)
	return mutatingTool(a, agent.ToolDef{
		Name:        "create_echo",
		Description: createEchoDescriptionFor(locale),
		Parameters:  json.RawMessage(createEchoSchema),
	}, func(_ context.Context, args json.RawMessage) (AskQuestion, applyFunc, error) {
		var in createEchoArgs
		if err := decodeWriteArgs(args, &in, ws); err != nil {
			return AskQuestion{}, nil, err
		}
		content := strings.TrimSpace(in.Content)
		if content == "" {
			return AskQuestion{}, nil, errors.New(ws.NoChange)
		}
		tags := normalizeTagNames(in.Tags)

		question := AskQuestion{
			ID:     "create_echo",
			Header: ws.CreateHeader,
			Text:   ws.CreateText,
			Detail: detailBlock(
				field{ws.FieldContent, clampRunes(content, maxConfirmContentRunes)},
				field{ws.FieldTags, tagLine(ws, tags)},
				field{ws.FieldVisibility, visibility(ws, in.Private)},
			),
			Options: []AskOption{{Label: ws.CreateYes}, {Label: ws.Cancel}},
		}

		return question, func(ctx context.Context) (agent.ToolOutput, error) {
			echo := &echoModel.Echo{
				Content: content,
				Private: in.Private,
				Tags:    tagModels(tags),
			}
			if err := s.echoService.PostEcho(ctx, echo); err != nil {
				return agent.ToolOutput{}, err
			}
			return agent.ToolOutput{Content: fmt.Sprintf(ws.Created, echo.ID)}, nil
		}, nil
	})
}

// updateEchoTool edits an existing Echo.
//
// The current row is loaded to build the confirmation, and the same values are
// carried into the write: UpdateEcho replaces files, layout and extension
// wholesale, so anything this tool does not show the person it must hand back
// unchanged. A chat that silently dropped an Echo's images would be a chat that
// destroyed them.
func (s *CopilotService) updateEchoTool(a *asker, locale string, loc *time.Location) agent.Tool {
	ws := writeStringsFor(locale)
	return mutatingTool(a, agent.ToolDef{
		Name:        "update_echo",
		Description: updateEchoDescriptionFor(locale),
		Parameters:  json.RawMessage(updateEchoSchema),
	}, func(ctx context.Context, args json.RawMessage) (AskQuestion, applyFunc, error) {
		var in updateEchoArgs
		if err := decodeWriteArgs(args, &in, ws); err != nil {
			return AskQuestion{}, nil, err
		}
		id, err := echoIDArg(in.ID, ws)
		if err != nil {
			return AskQuestion{}, nil, err
		}
		current, err := s.echoService.GetEchoById(ctx, id)
		if err != nil {
			return AskQuestion{}, nil, err
		}

		next := *current
		fields := []field{
			{ws.FieldID, current.ID},
			{ws.FieldContent, clampRunes(current.Content, maxConfirmContentRunes)},
		}
		changed := false

		if in.Content != nil {
			content := strings.TrimSpace(*in.Content)
			if content != "" && content != current.Content {
				next.Content = content
				fields = append(fields, field{ws.FieldNewContent, clampRunes(content, maxConfirmContentRunes)})
				changed = true
			}
		}
		if in.Tags != nil {
			tags := normalizeTagNames(*in.Tags)
			was := currentTagNames(current.Tags)
			if !sameStrings(was, tags) {
				next.Tags = tagModels(tags)
				fields = append(fields, field{ws.FieldTags, tagLine(ws, was) + ws.Arrow + tagLine(ws, tags)})
				changed = true
			}
		}
		if in.Private != nil && *in.Private != current.Private {
			next.Private = *in.Private
			fields = append(fields, field{
				ws.FieldVisibility,
				visibility(ws, current.Private) + ws.Arrow + visibility(ws, *in.Private),
			})
			changed = true
		}

		if !changed {
			return AskQuestion{}, nil, errors.New(ws.NoChange)
		}

		question := AskQuestion{
			ID:      "update_echo",
			Header:  ws.UpdateHeader,
			Text:    ws.UpdateText,
			Detail:  detailBlock(fields...),
			Options: []AskOption{{Label: ws.UpdateYes}, {Label: ws.Cancel}},
		}

		return question, func(ctx context.Context) (agent.ToolOutput, error) {
			if err := s.echoService.UpdateEcho(ctx, &next); err != nil {
				return agent.ToolOutput{}, err
			}
			return agent.ToolOutput{Content: fmt.Sprintf(ws.Updated, next.ID)}, nil
		}, nil
	})
}

// deleteEchoTool removes an Echo. The confirmation shows the whole thing that
// is about to disappear, because "delete Echo 019a…" is not something a person
// can consent to.
func (s *CopilotService) deleteEchoTool(a *asker, locale string, loc *time.Location) agent.Tool {
	ws := writeStringsFor(locale)
	return mutatingTool(a, agent.ToolDef{
		Name:        "delete_echo",
		Description: deleteEchoDescriptionFor(locale),
		Parameters:  json.RawMessage(deleteEchoSchema),
	}, func(ctx context.Context, args json.RawMessage) (AskQuestion, applyFunc, error) {
		var in deleteEchoArgs
		if err := decodeWriteArgs(args, &in, ws); err != nil {
			return AskQuestion{}, nil, err
		}
		id, err := echoIDArg(in.ID, ws)
		if err != nil {
			return AskQuestion{}, nil, err
		}
		current, err := s.echoService.GetEchoById(ctx, id)
		if err != nil {
			return AskQuestion{}, nil, err
		}

		if loc == nil {
			loc = time.UTC
		}
		detail := detailBlock(
			field{ws.FieldID, current.ID},
			field{ws.FieldContent, clampRunes(current.Content, maxConfirmContentRunes)},
			field{ws.FieldTags, tagLine(ws, currentTagNames(current.Tags))},
			field{ws.FieldPostedAt, time.Unix(current.CreatedAt, 0).In(loc).Format("2006-01-02")},
		)

		question := AskQuestion{
			ID:     "delete_echo",
			Header: ws.DeleteHeader,
			Text:   ws.DeleteText,
			Detail: detail,
			Options: []AskOption{
				{Label: ws.DeleteYes, Description: ws.DeleteNote},
				{Label: ws.Cancel},
			},
		}

		return question, func(ctx context.Context) (agent.ToolOutput, error) {
			if err := s.echoService.DeleteEchoById(ctx, id); err != nil {
				return agent.ToolOutput{}, err
			}
			return agent.ToolOutput{Content: fmt.Sprintf(ws.Deleted, id)}, nil
		}, nil
	})
}

// echoIDArg reads the Echo id a write tool was called with, and refuses
// anything that is not one.
//
// The refusal is worded as an instruction rather than as "not found", because
// the mistake it catches has a specific shape: search results are numbered
// 【1】【2】 for the model to refer to in prose, and a model told to pass "the
// id" reaches for that number. Handed to GetEchoById it would resolve to
// nothing, and the model would report the Echo as missing — a wrong answer,
// where the truth is that it quoted the wrong field. Told what it did instead,
// it re-reads the id= value and retries inside the same run.
//
// This is the enforcing half of a rule the schema and the prompt also state. It
// is here because those two are read by a model whose attention is finite, and
// this is not.
func echoIDArg(raw string, ws writeStrings) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" || !uuidUtil.IsValid(id) {
		return "", errors.New(ws.BadEchoID)
	}
	return id, nil
}

// decodeWriteArgs reads a write tool's arguments and refuses any field the tool
// cannot honour.
//
// encoding/json ignores unknown fields, which is the wrong default here: a
// model that passes files or extension would be told the Echo was posted, and
// the person would read that their images went with it. Neither is true — this
// tool writes text, tags and visibility and nothing else. Refusing is the only
// answer that is not a lie.
func decodeWriteArgs(args json.RawMessage, dst any, ws writeStrings) error {
	raw := bytes.TrimSpace(args)
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	// Lenient first. Malformed JSON, a truncated call and a wrong type are the
	// model's own mistakes, and they have to reach it worded as themselves.
	if err := json.Unmarshal(raw, dst); err != nil {
		return err
	}
	// Strict second. The bytes already parsed and type-checked, so the only
	// thing left to find is a field the tool does not have — the one mistake
	// that needs explaining, because the field it reached for does exist on an
	// Echo and simply is not writable from here.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%s: %w", ws.UnsupportedArg, err)
	}
	return nil
}

// field is one line of a confirmation's detail block.
type field struct {
	label string
	value string
}

// detailBlock renders the change as plain text, one field per line. Not
// Markdown: the client shows it preformatted, and content that happens to
// contain Markdown must read as the content it is.
func detailBlock(fields ...field) string {
	var b strings.Builder
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(f.label)
		b.WriteString("：")
		b.WriteString(f.value)
	}
	return b.String()
}

func visibility(ws writeStrings, private bool) string {
	if private {
		return ws.Private
	}
	return ws.Public
}

func tagLine(ws writeStrings, names []string) string {
	if len(names) == 0 {
		return ws.NoTags
	}
	return "#" + strings.Join(names, " #")
}

// normalizeTagNames trims, strips the leading hash the model tends to include,
// and drops duplicates while keeping the order the model chose.
func normalizeTagNames(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, name := range raw {
		name = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(name), "#"))
		if name == "" || slices.Contains(out, name) {
			continue
		}
		out = append(out, name)
	}
	return out
}

func tagModels(names []string) []echoModel.Tag {
	if len(names) == 0 {
		return nil
	}
	tags := make([]echoModel.Tag, 0, len(names))
	for _, name := range names {
		tags = append(tags, echoModel.Tag{Name: name})
	}
	return tags
}

func currentTagNames(tags []echoModel.Tag) []string {
	names := make([]string, 0, len(tags))
	for i := range tags {
		names = append(names, tags[i].Name)
	}
	return names
}

// sameStrings compares tag sets as sets: reordering the same tags is not an
// edit, and offering it as one would ask a person to approve nothing.
func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left := slices.Clone(a)
	right := slices.Clone(b)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}

const createEchoSchema = `{"type":"object","description":"只能发布正文、标签和可见性。图片、附件、音乐/视频/网站/位置/GitHub 项目等扩展卡片一律不支持，传了会直接报错——用户要发带图或带卡片的 Echo，请让 ta 在界面里发。","properties":{"content":{"type":"string","description":"Echo 的正文，用用户想发布的原话/内容，不要加你自己的解说"},"tags":{"type":"array","items":{"type":"string"},"description":"标签名（不带 #）；优先复用系统提示里列出的已有标签，不要凭空造新标签"},"private":{"type":"boolean","description":"true 为仅自己可见；用户没说就传 false"}},"required":["content"],"additionalProperties":false}`

const updateEchoSchema = `{"type":"object","description":"只能改正文、标签和可见性。这条 Echo 原有的图片、附件和扩展卡片会原样保留，但也无法用本工具修改或删除，传 files / extension 之类的字段会直接报错——用户要动这些，请让 ta 在界面里改。","properties":{"id":{"type":"string","description":"要修改的 Echo 的 ID：照抄 search_echos 结果里那条的 id= 后面的 UUID（形如 019ce0ea-82dd-774f-ae2d-5445512d42ad）。绝对不要传【1】这类序号，也不要自己编"},"content":{"type":"string","description":"新的正文；只在要改正文时传，传就是整条替换"},"tags":{"type":"array","items":{"type":"string"},"description":"新的标签名列表；只在要改标签时传，传就是整组替换（要保留原有标签就把它们一起写进来）"},"private":{"type":"boolean","description":"新的可见性；只在要改可见性时传"}},"required":["id"],"additionalProperties":false}`

const deleteEchoSchema = `{"type":"object","description":"整条删除，包括它的图片、附件和扩展卡片。无法只删其中一部分。","properties":{"id":{"type":"string","description":"要删除的 Echo 的 ID：照抄 search_echos 结果里那条的 id= 后面的 UUID（形如 019ce0ea-82dd-774f-ae2d-5445512d42ad）。绝对不要传【1】这类序号，也不要自己编"}},"required":["id"],"additionalProperties":false}`
