package service

import (
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"mime/multipart"
)

type CommonServiceInterface interface {
	CommonGetUserByUserId(userId uint) (userModel.User, error)
	UploadImage(userid uint, file *multipart.FileHeader) (string, error)
	DeleteImage(userid uint, url, source string) error
	DirectDeleteImage(url, source string) error
}
