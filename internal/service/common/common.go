package service

import (
	"errors"
	"fmt"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
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

func (commonService *CommonService) DeleteImage(userid uint, url, source string) error {
	user, err := commonService.commonRepository.GetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查图片是否存在
	if url == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	if source == echoModel.ImageSourceLocal {
		// 获取图片名字（去除前面的/images/)
		imageName := url[len("/images/"):]

		// 构造图片路径
		imagePath := fmt.Sprintf("data/images/%s", imageName)

		// 删除图片
		return storageUtil.DeleteFileFromLocal(imagePath)
	} else if source == echoModel.ImageSourceURL {
		// 无需处理
	} else if source == echoModel.ImageSourceS3 {

	} else if source == echoModel.ImageSourceR2 {

	} else {
		// 未知图片来源按本地图片处理
		// 获取图片名字（去除前面的/images/)
		imageName := url[len("/images/"):]

		// 构造图片路径
		imagePath := fmt.Sprintf("data/images/%s", imageName)

		// 删除图片
		return storageUtil.DeleteFileFromLocal(imagePath)
	}

	return nil
}

func (commonService *CommonService) DirectDeleteImage(url, source string) error {
	// 检查图片是否存在
	if url == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	if source == echoModel.ImageSourceLocal {
		// 获取图片名字（去除前面的/images/)
		imageName := url[len("/images/"):]

		// 构造图片路径
		imagePath := fmt.Sprintf("data/images/%s", imageName)

		// 删除图片
		return storageUtil.DeleteFileFromLocal(imagePath)
	} else if source == echoModel.ImageSourceURL {
		// 无需处理
	} else if source == echoModel.ImageSourceS3 {

	} else if source == echoModel.ImageSourceR2 {

	} else {
		// 未知图片来源按本地图片处理
		// 获取图片名字（去除前面的/images/)
		imageName := url[len("/images/"):]

		// 构造图片路径
		imagePath := fmt.Sprintf("data/images/%s", imageName)

		// 删除图片
		return storageUtil.DeleteFileFromLocal(imagePath)
	}

	return nil
}
