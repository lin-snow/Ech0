// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	model "github.com/lin-snow/ech0/internal/model/user"
	actorUtil "github.com/lin-snow/ech0/internal/util/actor"
	"gorm.io/gorm"
)

// PurgeUserContent removes every row owned by, authored by, or addressed to a
// user. It intentionally does not touch object storage: the caller deletes the
// returned files after the transaction commits, so a failed upload deletion can
// never resurrect half-deleted metadata. Callers must run it inside the user
// deletion transaction.
func (userRepository *UserRepository) PurgeUserContent(ctx context.Context, userID string) (model.PurgeResult, error) {
	db := userRepository.getDB(ctx)
	result := model.PurgeResult{}

	user, err := userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return result, err
	}

	var echoIDs []string
	if err := db.Model(&echoModel.Echo{}).
		Where("user_id = ?", userID).
		Pluck("id", &echoIDs).Error; err != nil {
		return result, err
	}

	var tagIDs []string
	if len(echoIDs) > 0 {
		if err := db.Table("echo_tags").
			Where("echo_id IN ?", echoIDs).
			Distinct().
			Pluck("tag_id", &tagIDs).Error; err != nil {
			return result, err
		}
	}

	// Only files uploaded by the user are deleted; files other people uploaded
	// and attached to this user's echoes belong to them (D3), so they survive
	// even though the link rows under the deleted echoes are removed.
	var fileIDs []string
	if err := db.Model(&fileModel.File{}).
		Where("user_id = ?", userID).
		Pluck("id", &fileIDs).Error; err != nil {
		return result, err
	}
	if len(fileIDs) > 0 {
		if err := db.Where("id IN ?", fileIDs).Find(&result.Files).Error; err != nil {
			return result, err
		}
	}

	if len(echoIDs) > 0 {
		// EchoFile / EchoExtension cascade in the schema, but SQLite foreign
		// keys are not guaranteed to be enforced, so delete explicitly.
		for _, table := range []any{
			&fileModel.EchoFile{},
			&echoModel.EchoExtension{},
			&echoModel.Reaction{},
			&commentModel.Comment{},
			&echoModel.EchoTag{},
		} {
			if err := db.Where("echo_id IN ?", echoIDs).Delete(table).Error; err != nil {
				return result, err
			}
		}
		if err := db.Where("id IN ?", echoIDs).Delete(&echoModel.Echo{}).Error; err != nil {
			return result, err
		}
	}

	// Comments the user wrote on surviving echoes. Comments under their own
	// echoes are already gone above.
	if err := db.Where("user_id = ?", userID).Delete(&commentModel.Comment{}).Error; err != nil {
		return result, err
	}
	// Reactions are addressed by an opaque actor string.
	if err := db.Where("actor = ?", actorUtil.UserActor(userID)).Delete(&echoModel.Reaction{}).Error; err != nil {
		return result, err
	}

	// Detach files from surviving echoes before the caller removes the rows,
	// otherwise the cascade would silently drop other users' echoes' media.
	if len(fileIDs) > 0 {
		if err := db.Where("file_id IN ?", fileIDs).Delete(&fileModel.EchoFile{}).Error; err != nil {
			return result, err
		}
	}
	if len(fileIDs) > 0 {
		if err := db.Where("uploader_id = ? OR file_id IN ?", userID, fileIDs).Delete(&fileModel.TempFile{}).Error; err != nil {
			return result, err
		}
	} else if err := db.Where("uploader_id = ?", userID).Delete(&fileModel.TempFile{}).Error; err != nil {
		return result, err
	}

	for _, table := range []any{
		&settingModel.AccessTokenSetting{},
		&model.WebAuthnCredential{},
		&model.UserExternalIdentity{},
		&authModel.Passkey{},
		&model.UserLocalAuth{},
	} {
		if err := db.Where("user_id = ?", userID).Delete(table).Error; err != nil {
			return result, err
		}
	}

	if err := db.Where("id = ?", userID).Delete(&model.User{}).Error; err != nil {
		return result, err
	}
	if err := recomputeTagUsage(db, tagIDs); err != nil {
		return result, err
	}

	userRepository.cache.Delete(GetUserIDKey(user.ID))
	userRepository.cache.Delete(GetUsernameKey(user.Username))
	userRepository.cache.Delete(GetAdminKey(user.ID))

	result.EchoIDs = echoIDs
	return result, nil
}

func recomputeTagUsage(db *gorm.DB, tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}
	return db.Exec(
		"UPDATE tags SET usage_count = (SELECT COUNT(*) FROM echo_tags WHERE echo_tags.tag_id = tags.id) WHERE id IN ?",
		tagIDs,
	).Error
}
