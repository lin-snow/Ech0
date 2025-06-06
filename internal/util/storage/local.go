package util

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func UploadFileToLocal(file *multipart.FileHeader, fileType commonModel.UploadFileType) (string, error) {
	// 创建图片存储目录
	if err := createImageDirIfNotExist(config.Config.Upload.ImagePath); err != nil {
		return "", err
	}

	// 获取原始文件名和扩展名
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(file.Filename, ext)

	// 使用 UUID 和原始文件名生成新的文件名
	newFileName := fmt.Sprintf("%s_%s%s", baseName, uuid.New().String(), ext)
	// 保存文件到指定目录
	savePath := filepath.Join(config.Config.Upload.ImagePath, newFileName)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(savePath), 0750); err != nil {
		return "", err
	}

	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	imageURL := fmt.Sprintf("/images/%s", newFileName)
	return imageURL, nil
}

func DeleteFileFromLocal(filePath string) error {
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		// 只有当错误不是"文件不存在"时才返回错误
		return err
	}
	return nil
}
