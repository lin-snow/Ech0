package service

import (
	"mime/multipart"
)

type CommonServiceInterface interface {
	UploadImage(userid uint, file *multipart.FileHeader) (string, error)
}
