package repository

import (
	"context"

	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type UserRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewUserRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
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
	return cache.ReadThroughTypedUnlessTx[model.User](
		ctx,
		userRepository.cache,
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
	if user.IsOwner {
		userRepository.cache.Set(GetOwnerKey(), *user, 1)
	}
	return nil
}

func (userRepository *UserRepository) GetUserByID(ctx context.Context, id string) (model.User, error) {
	cacheKey := GetUserIDKey(id)
	return cache.ReadThroughTypedUnlessTx[model.User](
		ctx,
		userRepository.cache,
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

func (userRepository *UserRepository) GetOwner(ctx context.Context) (model.User, error) {
	cacheKey := GetOwnerKey()
	return cache.ReadThroughTypedUnlessTx[model.User](
		ctx,
		userRepository.cache,
		cacheKey,
		1,
		func(ctx context.Context) (model.User, error) {
			user := model.User{}
			err := userRepository.getDB(ctx).Where("is_owner = ?", true).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
		func() (model.User, error) {
			user := model.User{}
			err := userRepository.db().Where("is_owner = ?", true).First(&user).Error
			if err != nil {
				return model.User{}, err
			}
			return user, nil
		},
	)
}

func (userRepository *UserRepository) IsInitialized(ctx context.Context) (bool, error) {
	var kv commonModel.KeyValue
	err := userRepository.getDB(ctx).Where("key = ?", commonModel.InstallInitializedKey).First(&kv).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return kv.Value == "true", nil
}

func (userRepository *UserRepository) MarkInitialized(ctx context.Context) error {
	result := userRepository.getDB(ctx).
		Model(&commonModel.KeyValue{}).
		Where("key = ?", commonModel.InstallInitializedKey).
		Update("value", "true")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	return userRepository.getDB(ctx).Create(&commonModel.KeyValue{
		Key:   commonModel.InstallInitializedKey,
		Value: "true",
	}).Error
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
	if existing.IsAdmin && !user.IsAdmin {
		userRepository.cache.Delete(GetAdminKey(user.ID))
	}
	if user.IsAdmin {
		userRepository.cache.Set(GetAdminKey(user.ID), *user, 1)
	}
	if existing.IsOwner && !user.IsOwner {
		userRepository.cache.Delete(GetOwnerKey())
	}
	if user.IsOwner {
		userRepository.cache.Set(GetOwnerKey(), *user, 1)
	}

	return nil
}

func (userRepository *UserRepository) DeleteUser(ctx context.Context, id string) error {
	userToDel, err := userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	err = userRepository.getDB(ctx).Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	userRepository.cache.Delete(GetUserIDKey(userToDel.ID))
	userRepository.cache.Delete(GetUsernameKey(userToDel.Username))
	if userToDel.IsAdmin {
		userRepository.cache.Delete(GetAdminKey(userToDel.ID))
	}
	if userToDel.IsOwner {
		userRepository.cache.Delete(GetOwnerKey())
	}

	return nil
}
