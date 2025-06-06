package util

import (
	"errors"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"mime/multipart"
	"os"
)

func UploadFile(file *multipart.FileHeader, fileType commonModel.UploadFileType, storageType commonModel.FileStorageType) (string, error) {
	if file == nil {
		return "", errors.New(commonModel.NO_FILE_UPLOAD_ERROR)
	}

	if storageType == commonModel.LOCAL_FILE {
		return UploadFileToLocal(file, fileType)
	}

	if storageType == commonModel.S3_FILE {

	}

	if storageType == commonModel.R2_FILE {

	}

	return "", errors.New(commonModel.NO_FILE_STORAGE_ERROR)
}

func IsAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// createImageDirIfNotExist 创建存储图片的目录
func createImageDirIfNotExist(imagePath string) error {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(imagePath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// FileExists 文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
