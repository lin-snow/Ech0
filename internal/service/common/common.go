package service

import (
	"errors"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/common"
	storageUtil "github.com/lin-snow/ech0/internal/util/storage"
	"mime/multipart"
)

type CommonService struct {
	commonRepository repository.CommonRepositoryInterface
}

func NewCommonService(commonRepository repository.CommonRepositoryInterface) CommonServiceInterface {
	return &CommonService{
		commonRepository: commonRepository,
	}
}

func (commonService *CommonService) CommonGetUserByUserId(userId uint) (userModel.User, error) {
	return commonService.commonRepository.GetUserByUserId(userId)
}

func (commonService *CommonService) UploadImage(userId uint, file *multipart.FileHeader) (string, error) {
	user, err := commonService.commonRepository.GetUserByUserId(userId)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查文件类型是否合法
	if !storageUtil.IsAllowedType(file.Header.Get("Content-Type"), config.Config.Upload.AllowedTypes) {
		return "", errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	// 检查文件大小是否合法
	if file.Size > int64(config.Config.Upload.ImageMaxSize) {
		return "", errors.New(commonModel.FILE_SIZE_EXCEED_LIMIT)
	}

	// 调用存储函数存储图片
	imageUrl, err := storageUtil.UploadFile(file, commonModel.ImageType, commonModel.LOCAL_FILE)
	if err != nil {
		return "", err
	}

	return imageUrl, nil
}
