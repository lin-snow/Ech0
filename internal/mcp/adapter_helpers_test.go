// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"encoding/json"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringArg(t *testing.T) {
	cases := []struct {
		name string
		args map[string]any
		key  string
		want string
	}{
		{"present string", map[string]any{"k": "v"}, "k", "v"},
		{"present empty string", map[string]any{"k": ""}, "k", ""},
		{"missing key", map[string]any{"other": "v"}, "k", ""},
		{"wrong type int", map[string]any{"k": 7}, "k", ""},
		{"wrong type float", map[string]any{"k": 1.5}, "k", ""},
		{"wrong type bool", map[string]any{"k": true}, "k", ""},
		{"nil value", map[string]any{"k": nil}, "k", ""},
		{"empty map", map[string]any{}, "k", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, stringArg(tc.args, tc.key))
		})
	}
}

func TestIntArg(t *testing.T) {
	cases := []struct {
		name     string
		args     map[string]any
		key      string
		fallback int
		want     int
	}{
		{"float64 truncates", map[string]any{"k": 3.9}, "k", -1, 3},
		{"float64 zero", map[string]any{"k": 0.0}, "k", -1, 0},
		{"float64 negative", map[string]any{"k": -2.0}, "k", 9, -2},
		{"json.Number valid", map[string]any{"k": json.Number("42")}, "k", -1, 42},
		{"json.Number negative", map[string]any{"k": json.Number("-5")}, "k", 0, -5},
		{"json.Number non-integer falls back", map[string]any{"k": json.Number("1.5")}, "k", 7, 7},
		{"json.Number garbage falls back", map[string]any{"k": json.Number("abc")}, "k", 7, 7},
		{"missing key falls back", map[string]any{}, "k", 11, 11},
		{"plain int type falls back", map[string]any{"k": 5}, "k", 11, 11},
		{"string type falls back", map[string]any{"k": "5"}, "k", 11, 11},
		{"nil value falls back", map[string]any{"k": nil}, "k", 11, 11},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, intArg(tc.args, tc.key, tc.fallback))
		})
	}
}

func TestBoolArg(t *testing.T) {
	cases := []struct {
		name string
		args map[string]any
		key  string
		want bool
	}{
		{"true", map[string]any{"k": true}, "k", true},
		{"false", map[string]any{"k": false}, "k", false},
		{"missing key", map[string]any{}, "k", false},
		{"wrong type string", map[string]any{"k": "true"}, "k", false},
		{"wrong type int", map[string]any{"k": 1}, "k", false},
		{"nil value", map[string]any{"k": nil}, "k", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, boolArg(tc.args, tc.key))
		})
	}
}

func TestBuildTags(t *testing.T) {
	t.Run("missing tags returns nil", func(t *testing.T) {
		assert.Nil(t, buildTags(map[string]any{}))
	})
	t.Run("tags wrong type returns nil", func(t *testing.T) {
		assert.Nil(t, buildTags(map[string]any{"tags": "not-an-array"}))
	})
	t.Run("tags map type returns nil", func(t *testing.T) {
		assert.Nil(t, buildTags(map[string]any{"tags": map[string]any{"a": "b"}}))
	})
	t.Run("filters empty and non-string entries", func(t *testing.T) {
		got := buildTags(map[string]any{"tags": []any{"go", "", 42, nil, "vue"}})
		assert.Equal(t, []echoModel.Tag{{Name: "go"}, {Name: "vue"}}, got)
	})
	t.Run("all-invalid yields nil slice", func(t *testing.T) {
		got := buildTags(map[string]any{"tags": []any{"", 1, true}})
		assert.Nil(t, got)
	})
}

func TestBuildEchoFiles(t *testing.T) {
	t.Run("missing echo_files returns nil", func(t *testing.T) {
		assert.Nil(t, buildEchoFiles(map[string]any{}))
	})
	t.Run("echo_files wrong type returns nil", func(t *testing.T) {
		assert.Nil(t, buildEchoFiles(map[string]any{"echo_files": "nope"}))
	})
	t.Run("default sort_order uses index and skips invalid entries", func(t *testing.T) {
		got := buildEchoFiles(map[string]any{"echo_files": []any{
			map[string]any{"file_id": "f0"},
			map[string]any{"sort_order": 9.0},
			"not-a-map",
			map[string]any{"file_id": "f3"},
		}})
		require.Len(t, got, 2)
		assert.Equal(t, echoModel.EchoFile{FileID: "f0", SortOrder: 0}, got[0])
		assert.Equal(t, echoModel.EchoFile{FileID: "f3", SortOrder: 3}, got[1])
	})
	t.Run("explicit sort_order overrides index default", func(t *testing.T) {
		got := buildEchoFiles(map[string]any{"echo_files": []any{
			map[string]any{"file_id": "f0", "sort_order": 5.0},
		}})
		require.Len(t, got, 1)
		assert.Equal(t, echoModel.EchoFile{FileID: "f0", SortOrder: 5}, got[0])
	})
	t.Run("empty file_id string is skipped", func(t *testing.T) {
		got := buildEchoFiles(map[string]any{"echo_files": []any{
			map[string]any{"file_id": ""},
		}})
		assert.Nil(t, got)
	})
}

func TestBuildExtension(t *testing.T) {
	t.Run("missing extension returns nil", func(t *testing.T) {
		assert.Nil(t, buildExtension(map[string]any{}))
	})
	t.Run("extension wrong type returns nil", func(t *testing.T) {
		assert.Nil(t, buildExtension(map[string]any{"extension": "nope"}))
	})
	t.Run("missing type returns nil", func(t *testing.T) {
		assert.Nil(t, buildExtension(map[string]any{"extension": map[string]any{"payload": map[string]any{}}}))
	})
	t.Run("empty type returns nil", func(t *testing.T) {
		assert.Nil(t, buildExtension(map[string]any{"extension": map[string]any{"type": ""}}))
	})
	t.Run("type with payload", func(t *testing.T) {
		got := buildExtension(map[string]any{"extension": map[string]any{
			"type":    "music",
			"payload": map[string]any{"url": "x"},
		}})
		require.NotNil(t, got)
		assert.Equal(t, "music", got.Type)
		assert.Equal(t, map[string]any{"url": "x"}, got.Payload)
	})
	t.Run("type without payload yields nil payload", func(t *testing.T) {
		got := buildExtension(map[string]any{"extension": map[string]any{"type": "music"}})
		require.NotNil(t, got)
		assert.Equal(t, "music", got.Type)
		assert.Nil(t, got.Payload)
	})
	t.Run("type with non-map payload yields nil payload", func(t *testing.T) {
		got := buildExtension(map[string]any{"extension": map[string]any{
			"type":    "music",
			"payload": "not-a-map",
		}})
		require.NotNil(t, got)
		assert.Nil(t, got.Payload)
	})
}

func TestJSONResult(t *testing.T) {
	t.Run("marshals value into single text content item", func(t *testing.T) {
		res, err := jsonResult(map[string]any{"a": 1})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Content, 1)
		assert.Equal(t, "text", res.Content[0].Type)
		assert.JSONEq(t, `{"a":1}`, res.Content[0].Text)
		assert.False(t, res.IsError)
	})
	t.Run("unmarshalable value returns error", func(t *testing.T) {
		res, err := jsonResult(make(chan int))
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "marshal result")
	})
}

func TestTextResult(t *testing.T) {
	res := textResult("hello")
	require.NotNil(t, res)
	require.Len(t, res.Content, 1)
	assert.Equal(t, "text", res.Content[0].Type)
	assert.Equal(t, "hello", res.Content[0].Text)
	assert.False(t, res.IsError)
}

func TestTextError(t *testing.T) {
	res := textError("boom")
	require.NotNil(t, res)
	require.Len(t, res.Content, 1)
	assert.Equal(t, "text", res.Content[0].Type)
	assert.Equal(t, "boom", res.Content[0].Text)
	assert.True(t, res.IsError)
}
