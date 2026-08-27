// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package event

import (
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type Named interface{ EventName() string }

type Keyed interface{ OrderingKey() string }

type (
	UserCreated struct{ User userModel.User }
	UserUpdated struct{ User userModel.User }
	UserDeleted struct{ User userModel.User }

	EchoCreated struct {
		Echo echoModel.Echo
		User userModel.User
	}
	EchoUpdated struct {
		Echo echoModel.Echo
		User userModel.User
	}
	EchoDeleted struct {
		Echo echoModel.Echo
		User userModel.User
	}

	CommentCreated       struct{ Comment commentModel.Comment }
	CommentStatusUpdated struct{ Comment commentModel.Comment }
	CommentDeleted       struct{ Comment commentModel.Comment }

	ResourceUploaded struct {
		User     userModel.User
		FileName string
		URL      string
		Size     int64
		Type     string
		Key      string `json:"-"`
	}

	SystemSnapshot struct {
		Info string
		Size int64
	}
	SystemExport struct {
		Info string
		Size int64
	}

	UpdateSnapshotSchedule struct {
		Schedule settingModel.SnapshotSchedule
	}
)

func (UserCreated) EventName() string            { return "user.created" }
func (UserUpdated) EventName() string            { return "user.updated" }
func (UserDeleted) EventName() string            { return "user.deleted" }
func (EchoCreated) EventName() string            { return "echo.created" }
func (EchoUpdated) EventName() string            { return "echo.updated" }
func (EchoDeleted) EventName() string            { return "echo.deleted" }
func (CommentCreated) EventName() string         { return "comment.created" }
func (CommentStatusUpdated) EventName() string   { return "comment.status.updated" }
func (CommentDeleted) EventName() string         { return "comment.deleted" }
func (ResourceUploaded) EventName() string       { return "resource.uploaded" }
func (SystemSnapshot) EventName() string         { return "system.snapshot" }
func (SystemExport) EventName() string           { return "system.export" }
func (UpdateSnapshotSchedule) EventName() string { return "system.snapshot_schedule.updated" }

func (e UserCreated) OrderingKey() string          { return e.User.ID }
func (e UserUpdated) OrderingKey() string          { return e.User.ID }
func (e UserDeleted) OrderingKey() string          { return e.User.ID }
func (e EchoCreated) OrderingKey() string          { return e.Echo.ID }
func (e EchoUpdated) OrderingKey() string          { return e.Echo.ID }
func (e EchoDeleted) OrderingKey() string          { return e.Echo.ID }
func (e CommentCreated) OrderingKey() string       { return e.Comment.ID }
func (e CommentStatusUpdated) OrderingKey() string { return e.Comment.ID }
func (e CommentDeleted) OrderingKey() string       { return e.Comment.ID }
func (e ResourceUploaded) OrderingKey() string     { return e.Key }
