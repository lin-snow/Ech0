// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoRepository_DeleteEchoById_RemovesExtension(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e1", "with extension", false, 100)
	seedEcho(t, db, "e2", "kept", false, 200)
	require.NoError(t, db.Create(&echoModel.EchoExtension{
		ID:     "ext1",
		EchoID: "e1",
		Type:   "demo",
	}).Error)
	require.NoError(t, db.Create(&echoModel.EchoExtension{
		ID:     "ext2",
		EchoID: "e2",
		Type:   "demo",
	}).Error)

	require.NoError(t, db.Create(&echoModel.Reaction{
		ID: "r1", EchoID: "e1", Emoji: "like", Actor: "v:actor",
	}).Error)

	require.NoError(t, repo.DeleteEchoById(context.Background(), "e1"))

	var reactionCount int64
	require.NoError(t, db.Model(&echoModel.Reaction{}).Where("echo_id = ?", "e1").Count(&reactionCount).Error)
	assert.Zero(t, reactionCount, "deleted echo's reactions should be removed")

	var count int64
	require.NoError(t, db.Model(&echoModel.EchoExtension{}).Where("echo_id = ?", "e1").Count(&count).Error)
	assert.Zero(t, count, "deleted echo's extension should be removed")

	require.NoError(t, db.Model(&echoModel.EchoExtension{}).Where("echo_id = ?", "e2").Count(&count).Error)
	assert.EqualValues(t, 1, count, "other echo's extension should stay")
}
