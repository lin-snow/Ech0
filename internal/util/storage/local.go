package util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

// UploadFileToLocal 根据文件类型上传文件到本地存储
func UploadFileToLocal(file *multipart.FileHeader, fileType commonModel.UploadFileType, userID uint) (string, error) {
	// 根据文件类型选择上传方式
	switch fileType {
	case commonModel.ImageType:
		// 上传图片到本地
		return UploadImageToLocal(file, userID)
	case commonModel.AudioType:
		// 上传音频到本地
		return UploadAudioToLocal(file, userID)
	case commonModel.ModelType:
		// 上传3D模型到本地
		return UploadModelToLocal(file, userID)
	default:
		// 不支持的文件类型
		return "", errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}
}

// UploadImageToLocal 将图片上传到本地存储
func UploadImageToLocal(file *multipart.FileHeader, userID uint) (string, error) {
	// 创建图片存储目录
	if err := createDirIfNotExist(config.Config.Upload.ImagePath); err != nil {
		return "", err
	}

	// 获取原始文件名和扩展名
	ext := filepath.Ext(file.Filename)
	// baseName := strings.TrimSuffix(file.Filename, ext)

	// 生成新的文件名,格式为[userID]_[timestamp]_[random].[ext]
	newFileName, err := GenerateRandomFilename(userID, ext)
	if err != nil {
		return "", err
	}
	// 保存文件到指定目录
	savePath := filepath.Join(config.Config.Upload.ImagePath, newFileName)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			log.Println("Failed to close file source:", closeErr)
		}
	}()

	if err = os.MkdirAll(filepath.Dir(savePath), 0o750); err != nil {
		return "", err
	}

	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer func() {
		// 确保文件被正确关闭
		if closeErr := out.Close(); closeErr != nil {
			log.Println("Failed to close destination file:", closeErr)
		}
	}()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	// 返回图片的 URL
	imageURL := fmt.Sprintf("/images/%s", newFileName)
	return imageURL, nil
}

// UploadAudioToLocal 将音频上传到本地存储
func UploadAudioToLocal(file *multipart.FileHeader, userID uint) (string, error) {
	// 创建音频存储目录
	if err := createDirIfNotExist(config.Config.Upload.AudioPath); err != nil {
		return "", err
	}

	// 获取扩展名
	ext := filepath.Ext(file.Filename)

	// 重名音频文件名（暂时使用固定名字 music + 扩展名）
	newFileName := fmt.Sprintf("music%s", ext)
	savePath := filepath.Join(config.Config.Upload.AudioPath, newFileName)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		// 确保文件被正确关闭
		if closeErr := src.Close(); closeErr != nil {
			log.Println("Failed to close file source:", closeErr)
		}
	}()

	if err = os.MkdirAll(filepath.Dir(savePath), 0o750); err != nil {
		return "", err
	}

	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer func() {
		// 确保文件被正确关闭
		if closeErr := out.Close(); closeErr != nil {
			log.Println("Failed to close destination file:", closeErr)
		}
	}()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	// 返回音频的 URL
	audioURL := fmt.Sprintf("/audios/%s", newFileName)
	return audioURL, nil
}

// DeleteFileFromLocal 删除本地文件
func DeleteFileFromLocal(filePath string) error {
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		// 只有当错误不是"文件不存在"时才返回错误
		return err
	}
	return nil
}

func GenerateRandomFilename(userID uint, ext string) (string, error) {
	timestamp := time.Now().Unix()
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	randomStr := hex.EncodeToString(bytes)

	// 确保扩展名前带点
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	newFileName := fmt.Sprintf("%d_%d_%s%s", userID, timestamp, randomStr, ext)
	return newFileName, nil
}

// UploadModelToLocal 将3D模型上传到本地存储
func UploadModelToLocal(file *multipart.FileHeader, userID uint) (string, error) {
	// 创建模型存储目录
	if err := createDirIfNotExist(config.Config.Upload.ModelPath); err != nil {
		return "", err
	}

	// 获取原始文件名和扩展名
	ext := filepath.Ext(file.Filename)

	// 生成新的文件名,格式为[userID]_[timestamp]_[random].[ext]
	newFileName, err := GenerateRandomFilename(userID, ext)
	if err != nil {
		return "", err
	}

	// 保存文件到指定目录
	savePath := filepath.Join(config.Config.Upload.ModelPath, newFileName)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			log.Println("Failed to close file source:", closeErr)
		}
	}()

	if err = os.MkdirAll(filepath.Dir(savePath), 0o750); err != nil {
		return "", err
	}

	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			log.Println("Failed to close destination file:", closeErr)
		}
	}()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	// 返回模型的 URL
	modelURL := fmt.Sprintf("/models/%s", newFileName)
	return modelURL, nil
}
