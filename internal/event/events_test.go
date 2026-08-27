// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package event

import (
	"testing"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventName(t *testing.T) {
	cases := []struct {
		name string
		ev   Named
		want string
	}{
		{"UserCreated", UserCreated{}, "user.created"},
		{"UserUpdated", UserUpdated{}, "user.updated"},
		{"UserDeleted", UserDeleted{}, "user.deleted"},
		{"EchoCreated", EchoCreated{}, "echo.created"},
		{"EchoUpdated", EchoUpdated{}, "echo.updated"},
		{"EchoDeleted", EchoDeleted{}, "echo.deleted"},
		{"CommentCreated", CommentCreated{}, "comment.created"},
		{"CommentStatusUpdated", CommentStatusUpdated{}, "comment.status.updated"},
		{"CommentDeleted", CommentDeleted{}, "comment.deleted"},
		{"ResourceUploaded", ResourceUploaded{}, "resource.uploaded"},
		{"SystemSnapshot", SystemSnapshot{}, "system.snapshot"},
		{"SystemExport", SystemExport{}, "system.export"},
		{"UpdateSnapshotSchedule", UpdateSnapshotSchedule{}, "system.snapshot_schedule.updated"},
	}

	require.Len(t, cases, 13, "expected exactly 13 named events")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ev.EventName())
		})
	}
}

func TestEventNamesUnique(t *testing.T) {
	names := []Named{
		UserCreated{},
		UserUpdated{},
		UserDeleted{},
		EchoCreated{},
		EchoUpdated{},
		EchoDeleted{},
		CommentCreated{},
		CommentStatusUpdated{},
		CommentDeleted{},
		ResourceUploaded{},
		SystemSnapshot{},
		SystemExport{},
		UpdateSnapshotSchedule{},
	}

	seen := make(map[string]string, len(names))
	for _, n := range names {
		topic := n.EventName()
		assert.NotEmpty(t, topic, "topic must not be empty")
		if prev, dup := seen[topic]; dup {
			t.Errorf("duplicate topic %q used by multiple events (also %q)", topic, prev)
			continue
		}
		seen[topic] = topic
	}
	assert.Len(t, seen, len(names), "every event must have a distinct topic")
}

func TestOrderingKey(t *testing.T) {
	const (
		userID    = "user-key-0001"
		echoID    = "echo-key-0001"
		commentID = "comment-key-0001"
		storeKey  = "uploads/2026/abc.png"
	)

	cases := []struct {
		name string
		ev   Keyed
		want string
	}{
		{"UserCreated", UserCreated{User: userModel.User{ID: userID}}, userID},
		{"UserUpdated", UserUpdated{User: userModel.User{ID: userID}}, userID},
		{"UserDeleted", UserDeleted{User: userModel.User{ID: userID}}, userID},
		{"EchoCreated", EchoCreated{Echo: echoModel.Echo{ID: echoID}}, echoID},
		{"EchoUpdated", EchoUpdated{Echo: echoModel.Echo{ID: echoID}}, echoID},
		{"EchoDeleted", EchoDeleted{Echo: echoModel.Echo{ID: echoID}}, echoID},
		{"CommentCreated", CommentCreated{Comment: commentModel.Comment{ID: commentID}}, commentID},
		{"CommentStatusUpdated", CommentStatusUpdated{Comment: commentModel.Comment{ID: commentID}}, commentID},
		{"CommentDeleted", CommentDeleted{Comment: commentModel.Comment{ID: commentID}}, commentID},
		{"ResourceUploaded", ResourceUploaded{Key: storeKey}, storeKey},
	}

	require.Len(t, cases, 10, "expected exactly 10 keyed events")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ev.OrderingKey())
		})
	}
}

func TestOrderingKey_EmptyWhenIDUnset(t *testing.T) {
	assert.Empty(t, EchoCreated{}.OrderingKey())
	assert.Empty(t, UserCreated{}.OrderingKey())
	assert.Empty(t, CommentCreated{}.OrderingKey())
	assert.Empty(t, ResourceUploaded{FileName: "a.png", URL: "http://x/a.png"}.OrderingKey())
}

func TestSystemEventsNotKeyed(t *testing.T) {
	cases := []struct {
		name string
		ev   Named
	}{
		{"SystemSnapshot", SystemSnapshot{}},
		{"SystemExport", SystemExport{}},
		{"UpdateSnapshotSchedule", UpdateSnapshotSchedule{Schedule: settingModel.SnapshotSchedule{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tc.ev.(Keyed)
			assert.False(t, ok, "%s must NOT implement Keyed", tc.name)
		})
	}
}

func TestKeyedEventsAlsoNamed(t *testing.T) {
	keyed := []Keyed{
		UserCreated{},
		UserUpdated{},
		UserDeleted{},
		EchoCreated{},
		EchoUpdated{},
		EchoDeleted{},
		CommentCreated{},
		CommentStatusUpdated{},
		CommentDeleted{},
		ResourceUploaded{},
	}
	for _, k := range keyed {
		n, ok := k.(Named)
		require.True(t, ok, "keyed event must also be Named")
		assert.NotEmpty(t, n.EventName())
	}
}
