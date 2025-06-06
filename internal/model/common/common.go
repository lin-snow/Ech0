package model

// File 相关
type UploadFileType string
type FileStorageType string

const (
	ImageType UploadFileType = "image"
	AudioType UploadFileType = "audio"
)

const (
	LOCAL_FILE FileStorageType = "local"
	S3_FILE    FileStorageType = "s3"
	R2_FILE    FileStorageType = "r2"
)
