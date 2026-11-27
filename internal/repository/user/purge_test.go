// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPurgeUserContent_RemovesOwnedAndAuthoredRows(t *testing.T) {
	repo, db, _ := newUserRepo(t)
	ctx := context.Background()

	owner := userModel.User{ID: "u-owner", Username: "owner"}
	other := userModel.User{ID: "u-other", Username: "other"}
	require.NoError(t, db.Create(&owner).Error)
	require.NoError(t, db.Create(&other).Error)

	ownedEcho := echoModel.Echo{ID: "e-owned", Content: "owned", UserID: owner.ID, Private: true}
	survivingEcho := echoModel.Echo{ID: "e-other", Content: "other", UserID: other.ID}
	require.NoError(t, db.Create(&ownedEcho).Error)
	require.NoError(t, db.Create(&survivingEcho).Error)

	tagOwned := echoModel.Tag{ID: "t-owned", Name: "owned-tag", UsageCount: 1}
	tagOther := echoModel.Tag{ID: "t-other", Name: "other-tag", UsageCount: 1}
	require.NoError(t, db.Create(&tagOwned).Error)
	require.NoError(t, db.Create(&tagOther).Error)
	require.NoError(t, db.Create(&echoModel.EchoTag{EchoID: ownedEcho.ID, TagID: tagOwned.ID}).Error)
	require.NoError(t, db.Create(&echoModel.EchoTag{EchoID: survivingEcho.ID, TagID: tagOther.ID}).Error)

	ownedFile := seedRefFile(t, db, "f-owned", owner.ID, "owned")
	attachedOtherFile := seedRefFile(t, db, "f-attached", other.ID, "attached")
	survivingFile := seedRefFile(t, db, "f-other", other.ID, "other")
	require.NoError(t, db.Create(&fileModel.Link{FileID: ownedFile.ID, OwnerKind: fileModel.OwnerKindEcho, OwnerID: ownedEcho.ID, Role: fileModel.RoleAttachment}).Error)
	require.NoError(t, db.Create(&fileModel.Link{FileID: attachedOtherFile.ID, OwnerKind: fileModel.OwnerKindEcho, OwnerID: ownedEcho.ID, Role: fileModel.RoleAttachment}).Error)
	require.NoError(t, db.Create(&fileModel.Link{FileID: survivingFile.ID, OwnerKind: fileModel.OwnerKindEcho, OwnerID: survivingEcho.ID, Role: fileModel.RoleAttachment}).Error)
	// The user's own file is linked to a surviving echo so detaching is visible.
	require.NoError(t, db.Create(&fileModel.Link{FileID: ownedFile.ID, OwnerKind: fileModel.OwnerKindEcho, OwnerID: survivingEcho.ID, Role: fileModel.RoleAttachment}).Error)
	require.NoError(t, db.Create(&echoModel.EchoExtension{ID: "ext-1", EchoID: ownedEcho.ID, Type: "weather", Payload: map[string]any{"t": 1}}).Error)
	require.NoError(t, db.Create(&authModel.Passkey{ID: "pk-1", UserID: owner.ID, CredentialID: "cred-1", CredentialJSON: "{}"}).Error)
	require.NoError(t, db.Create(&userModel.WebAuthnCredential{ID: "wa-1", UserID: owner.ID, CredentialID: "cred-2", CredentialJSON: "{}"}).Error)
	require.NoError(t, db.Create(&userModel.UserExternalIdentity{ID: "idp-1", UserID: owner.ID, Provider: "github", Subject: "s-1"}).Error)

	comments := []commentModel.Comment{
		{ID: "c-under-owned-by-other", EchoID: ownedEcho.ID, UserID: ptr(other.ID), Content: "x", Status: commentModel.StatusApproved},
		{ID: "c-under-owned-by-owner", EchoID: ownedEcho.ID, UserID: ptr(owner.ID), Content: "x", Status: commentModel.StatusApproved},
		{ID: "c-owner-elsewhere", EchoID: survivingEcho.ID, UserID: ptr(owner.ID), Content: "x", Status: commentModel.StatusApproved},
		{ID: "c-other-elsewhere", EchoID: survivingEcho.ID, UserID: ptr(other.ID), Content: "x", Status: commentModel.StatusApproved},
	}
	require.NoError(t, db.Create(&comments).Error)

	require.NoError(t, db.Create(&echoModel.Reaction{EchoID: survivingEcho.ID, Emoji: "like", Actor: "u:" + owner.ID}).Error)
	require.NoError(t, db.Create(&echoModel.Reaction{EchoID: survivingEcho.ID, Emoji: "like", Actor: "u:" + other.ID}).Error)
	require.NoError(t, db.Create(&settingModel.AccessTokenSetting{ID: "at-1", UserID: owner.ID, Token: "tok", JTI: "jti-1"}).Error)
	require.NoError(t, db.Create(&userModel.UserLocalAuth{UserID: owner.ID, PasswordHash: "h", PasswordAlgo: "bcrypt"}).Error)

	result, err := repo.PurgeUserContent(ctx, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{ownedEcho.ID}, result.EchoIDs)
	assert.ElementsMatch(t, []string{ownedFile.ID}, fileIDs(result))

	assertRowGone(t, db, &echoModel.Echo{}, "e-owned")
	assertRowPresent(t, db, &echoModel.Echo{}, "e-other")

	// Both media rows are still present at the repository layer; the service
	// deletes the user's own file via DeleteFileRecord after the transaction.
	assertRowPresent(t, db, &fileModel.File{}, ownedFile.ID)
	assertRowPresent(t, db, &fileModel.File{}, attachedOtherFile.ID)
	var links int64
	require.NoError(t, db.Model(&fileModel.Link{}).Where("owner_id = ?", ownedEcho.ID).Count(&links).Error)
	assert.Zero(t, links, "被删除 Echo 的附件关联必须清空")
	require.NoError(t, db.Model(&fileModel.Link{}).Where("file_id = ?", survivingFile.ID).Count(&links).Error)
	assert.Equal(t, int64(1), links, "其他用户 Echo 的附件不能受影响")
	require.NoError(t, db.Model(&fileModel.Link{}).Where("file_id = ?", ownedFile.ID).Count(&links).Error)
	assert.Zero(t, links, "被删用户的文件要从存活 Echo 上解绑")
	require.NoError(t, db.Model(&fileModel.Link{}).Where("file_id = ?", attachedOtherFile.ID).Count(&links).Error)
	assert.Zero(t, links, "被删除 Echo 下他人文件的链接随 Echo 一起清空")

	assertRowGone(t, db, &commentModel.Comment{}, "c-under-owned-by-other")
	assertRowGone(t, db, &commentModel.Comment{}, "c-under-owned-by-owner")
	assertRowGone(t, db, &commentModel.Comment{}, "c-owner-elsewhere")
	assertRowPresent(t, db, &commentModel.Comment{}, "c-other-elsewhere")

	var reactions int64
	require.NoError(t, db.Model(&echoModel.Reaction{}).Where("actor = ?", "u:"+owner.ID).Count(&reactions).Error)
	assert.Zero(t, reactions)
	require.NoError(t, db.Model(&echoModel.Reaction{}).Where("actor = ?", "u:"+other.ID).Count(&reactions).Error)
	assert.Equal(t, int64(1), reactions)

	assertRowGone(t, db, &echoModel.EchoExtension{}, "ext-1")
	assertRowGone(t, db, &authModel.Passkey{}, "pk-1")
	assertRowGone(t, db, &userModel.WebAuthnCredential{}, "wa-1")
	assertRowGone(t, db, &userModel.UserExternalIdentity{}, "idp-1")
	assertRowGone(t, db, &settingModel.AccessTokenSetting{}, "at-1")
	require.NoError(t, db.Model(&userModel.UserLocalAuth{}).Where("user_id = ?", owner.ID).Count(&links).Error)
	assert.Zero(t, links)
	assertRowGone(t, db, &userModel.User{}, owner.ID)

	var reloadedOwnedTag echoModel.Tag
	require.NoError(t, db.First(&reloadedOwnedTag, "id = ?", tagOwned.ID).Error)
	assert.Zero(t, reloadedOwnedTag.UsageCount, "孤立标签的引用计数必须重算")
	var reloadedOtherTag echoModel.Tag
	require.NoError(t, db.First(&reloadedOtherTag, "id = ?", tagOther.ID).Error)
	assert.Equal(t, 1, reloadedOtherTag.UsageCount, "其他标签不受影响")
}

func seedRefFile(t *testing.T, db *gorm.DB, id, uploaderID, name string) fileModel.File {
	t.Helper()
	refURL := "https://example.com/" + id
	f := fileModel.File{
		ID:          id,
		Kind:        string(fileModel.KindRef),
		Status:      string(fileModel.StatusReady),
		Access:      string(fileModel.AccessPublic),
		RefURL:      &refURL,
		Name:        name,
		ContentType: "image/png",
		UploaderID:  uploaderID,
	}
	require.NoError(t, db.Create(&f).Error)
	return f
}

func ptr[T any](v T) *T { return &v }

func fileIDs(result userModel.PurgeResult) []string {
	ids := make([]string, 0, len(result.Files))
	for _, f := range result.Files {
		ids = append(ids, f.ID)
	}
	return ids
}

func assertRowGone(t *testing.T, db *gorm.DB, model any, id string) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(model).Where("id = ?", id).Count(&count).Error)
	assert.Zero(t, count, "row %s must be deleted", id)
}

func assertRowPresent(t *testing.T, db *gorm.DB, model any, id string) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(model).Where("id = ?", id).Count(&count).Error)
	assert.Equal(t, int64(1), count, "row %s must survive", id)
}
