// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	"github.com/lin-snow/ech0/internal/cache"
	model "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db    func() *gorm.DB
	cache *cache.Cache
}

func NewUserRepository(
	dbProvider func() *gorm.DB,
	cache *cache.Cache,
) *UserRepository {
	return &UserRepository{
		db:    dbProvider,
		cache: cache,
	}
}

func (userRepository *UserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return userRepository.db()
}

func (userRepository *UserRepository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	cacheKey := GetUsernameKey(username)
	return userRepository.cache.ReadThroughUnlessTx(
		ctx,
		cacheKey,
		1,
		func(ctx context.Context) (model.User, error) {
			user := model.User{}
			err := userRepository.getDB(ctx).Where("username = ?", username).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
		func() (model.User, error) {
			user := model.User{}
			err := userRepository.db().Where("username = ?", username).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
	)
}

func (userRepository *UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := userRepository.getDB(ctx).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (userRepository *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := userRepository.getDB(ctx).Create(user).Error
	if err != nil {
		return err
	}

	userRepository.cache.Set(GetUserIDKey(user.ID), *user, 1)
	userRepository.cache.Set(GetUsernameKey(user.Username), *user, 1)
	return nil
}

func (userRepository *UserRepository) UpsertLocalAuth(ctx context.Context, localAuth *model.UserLocalAuth) error {
	return userRepository.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"password_hash", "password_algo", "updated_at"}),
	}).Create(localAuth).Error
}

func (userRepository *UserRepository) GetUserByID(ctx context.Context, id string) (model.User, error) {
	cacheKey := GetUserIDKey(id)
	return userRepository.cache.ReadThroughUnlessTx(
		ctx,
		cacheKey,
		1,
		func(ctx context.Context) (model.User, error) {
			var user model.User
			if err := userRepository.getDB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
				return user, err
			}
			return user, nil
		},
		func() (model.User, error) {
			var user model.User
			if err := userRepository.db().Where("id = ?", id).First(&user).Error; err != nil {
				return user, err
			}
			return user, nil
		},
	)
}

// GetOwner returns the single account the space belongs to. Errors are
// propagated so callers can tell "not initialized yet" from a database
// failure. It is deliberately uncached — an ownership transfer changes the
// answer, and the lookup is a single indexed row.
func (userRepository *UserRepository) GetOwner(ctx context.Context) (model.User, error) {
	user := model.User{}
	err := userRepository.getDB(ctx).
		Where("role = ?", model.RoleOwner).
		First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// IsInitialized answers whether the space has been claimed, which is exactly
// "an owner exists". There is no separate install marker to keep in sync.
func (userRepository *UserRepository) IsInitialized(ctx context.Context) (bool, error) {
	var count int64
	err := userRepository.getDB(ctx).
		Model(&model.User{}).
		Where("role = ?", model.RoleOwner).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateUserRole moves one account between member and admin. It never touches
// the owner row: ownership changes only through TransferOwnership.
func (userRepository *UserRepository) UpdateUserRole(ctx context.Context, id string, newRole model.Role) error {
	user, err := userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	user.Role = newRole
	return userRepository.UpdateUser(ctx, &user)
}

// TransferOwnership demotes the current owner and promotes the next one. The
// order is load-bearing: SQLite checks the partial unique index row by row, so
// promoting first collides with the existing owner. Both statements run in one
// transaction, and this is the only place in the codebase that may swap roles
// with two updates.
func (userRepository *UserRepository) TransferOwnership(
	ctx context.Context,
	currentOwnerID string,
	newOwnerID string,
) error {
	var previousOwner, newOwner model.User
	err := userRepository.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", currentOwnerID).First(&previousOwner).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", newOwnerID).First(&newOwner).Error; err != nil {
			return err
		}
		demoted := tx.Model(&model.User{}).
			Where("id = ? AND role = ?", previousOwner.ID, model.RoleOwner).
			Update("role", model.RoleAdmin)
		if demoted.Error != nil {
			return demoted.Error
		}
		if demoted.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		promoted := tx.Model(&model.User{}).
			Where("id = ? AND role = ?", newOwner.ID, model.RoleAdmin).
			Update("role", model.RoleOwner)
		if promoted.Error != nil {
			return promoted.Error
		}
		if promoted.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Role is cached with the rest of the row under both keys; raw updates do
	// not refresh it, so both accounts are evicted instead of left stale.
	for _, u := range []model.User{previousOwner, newOwner} {
		userRepository.cache.Delete(GetUserIDKey(u.ID))
		userRepository.cache.Delete(GetUsernameKey(u.Username))
	}
	return nil
}

func (userRepository *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	var existing model.User
	if err := userRepository.getDB(ctx).Where("id = ?", user.ID).First(&existing).Error; err != nil {
		return err
	}

	err := userRepository.getDB(ctx).Save(user).Error
	if err != nil {
		return err
	}

	userRepository.cache.Set(GetUserIDKey(user.ID), *user, 1)
	if existing.Username != "" && existing.Username != user.Username {
		userRepository.cache.Delete(GetUsernameKey(existing.Username))
	}
	userRepository.cache.Set(GetUsernameKey(user.Username), *user, 1)

	return nil
}
