package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	imgUtil "github.com/lin-snow/ech0/internal/util/img"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	externalKeyPrefix   = "external/"
	externalDefaultName = "external"
	treeNodeTypeFile    = "file"
	treeNodeTypeFolder  = "folder"
)

type FileService struct {
	transactor         transaction.Transactor
	commonRepository   CommonRepository
	storageManager     *storage.Manager
	keyvalueRepository KeyValueRepository
	fileRepository     FileRepository
	publisher          *publisher.Publisher
	keyGen             storage.KeyGenerator
	settingService     SettingService
}

func NewFileService(
	tx transaction.Transactor,
	commonRepository CommonRepository,
	kvRepo KeyValueRepository,
	fileRepo FileRepository,
	storageManager *storage.Manager,
	publisher *publisher.Publisher,
	settingService SettingService,
) *FileService {
	return &FileService{
		transactor:         tx,
		commonRepository:   commonRepository,
		keyvalueRepository: kvRepo,
		fileRepository:     fileRepo,
		storageManager:     storageManager,
		publisher:          publisher,
		keyGen:             storage.NewRandomKeyGenerator(),
		settingService:     settingService,
	}
}

func (s *FileService) UploadFile(
	ctx context.Context,
	file *multipart.FileHeader,
	category storage.Category,
	storageType storage.StorageType,
) (commonModel.FileDto, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userID)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	contentType := file.Header.Get("Content-Type")
	if !isAllowedType(contentType, config.Config().Upload.AllowedTypes) {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	limit := int64(config.Config().Upload.ImageMaxSize)
	if category == storage.CategoryAudio {
		limit = int64(config.Config().Upload.AudioMaxSize)
	}
	if file.Size > limit {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_SIZE_EXCEED_LIMIT)
	}

	gen := s.keyGenForCategory(category, file.Filename)
	key, err := gen.GenerateKey(category, user.ID, file.Filename)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	reader, err := file.Open()
	if err != nil {
		return commonModel.FileDto{}, err
	}
	defer func() { _ = reader.Close() }()

	var opts []virefs.PutOption
	if contentType != "" {
		opts = append(opts, virefs.WithContentType(contentType))
	}

	targetStorageType := storage.NormalizeStorageType(string(storageType))
	if targetStorageType == storage.StorageTypeExternal {
		targetStorageType = storage.StorageTypeLocal
	}
	selector := s.getSelector()
	if err := selector.Put(context.Background(), targetStorageType, key, reader, opts...); err != nil {
		return commonModel.FileDto{}, err
	}

	width, height := 0, 0
	if category.IsImageLike() {
		width, height, err = imgUtil.GetImageSizeFromFile(file)
		if err != nil {
			return commonModel.FileDto{}, err
		}
	}

	fileURL := selector.ResolveURL(targetStorageType, key)
	routeStorageType, provider, bucket := currentStorageRoute(selector, targetStorageType)

	fileRecord := &fileModel.File{
		Key:         key,
		StorageType: routeStorageType,
		Provider:    provider,
		Bucket:      bucket,
		URL:         fileURL,
		Name:        file.Filename,
		ContentType: contentType,
		Size:        file.Size,
		Category:    string(category),
		Width:       width,
		Height:      height,
		UserID:      user.ID,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return commonModel.FileDto{}, err
	}

	uploadType := commonModel.ImageType
	if category == storage.CategoryAudio {
		uploadType = commonModel.AudioType
	}

	user.Password = ""
	if err := s.publisher.ResourceUploaded(
		context.Background(),
		contracts.ResourceUploadedEvent{
			User:     user,
			FileName: file.Filename,
			URL:      fileURL,
			Size:     file.Size,
			Type:     string(uploadType),
		},
		key,
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish resource uploaded event", zap.Error(err))
	}

	return commonModel.FileDto{
		ID:          fileRecord.ID,
		Name:        file.Filename,
		Key:         key,
		StorageType: routeStorageType,
		URL:         fileURL,
		ContentType: contentType,
		Category:    string(category),
		Size:        file.Size,
		Width:       width,
		Height:      height,
	}, nil
}

func (s *FileService) CreateExternalFile(
	ctx context.Context,
	dto commonModel.CreateExternalFileDto,
) (commonModel.FileDto, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	rawURL := httpUtil.TrimURL(dto.URL)
	if rawURL == "" {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}

	contentType := strings.TrimSpace(dto.ContentType)
	if contentType == "" {
		contentType = httpUtil.GetMIMETypeFromFilenameOrURL(rawURL)
	}

	category := storage.NormalizeCategory(dto.Category)
	if category == "" {
		if strings.HasPrefix(contentType, "audio/") {
			category = storage.CategoryAudio
		} else {
			category = storage.CategoryImage
		}
	}

	normalizedURL := parsed.String()
	hash := sha256.Sum256([]byte(normalizedURL))
	key := externalKeyPrefix + string(category) + "/" + hex.EncodeToString(hash[:])

	const (
		externalStorageType = string(storage.StorageTypeExternal)
		externalProvider    = string(storage.StorageTypeExternal)
		externalBucket      = ""
	)
	existing, err := s.fileRepository.GetByRoute(
		context.Background(),
		externalStorageType,
		externalProvider,
		externalBucket,
		key,
	)
	if err == nil && existing != nil {
		return commonModel.FileDto{
			ID:          existing.ID,
			Name:        existing.Name,
			Key:         existing.Key,
			StorageType: existing.StorageType,
			URL:         existing.URL,
			ContentType: existing.ContentType,
			Category:    existing.Category,
			Size:        existing.Size,
			Width:       existing.Width,
			Height:      existing.Height,
		}, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return commonModel.FileDto{}, err
	}

	name := strings.TrimSpace(dto.Name)
	if name == "" {
		name = filepath.Base(parsed.Path)
		if name == "" || name == "." || name == "/" {
			name = externalDefaultName
		}
	}

	fileRecord := &fileModel.File{
		Key:         key,
		StorageType: externalStorageType,
		Provider:    externalProvider,
		Bucket:      externalBucket,
		URL:         normalizedURL,
		Name:        name,
		ContentType: contentType,
		Size:        0,
		Category:    string(category),
		Width:       dto.Width,
		Height:      dto.Height,
		UserID:      user.ID,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return commonModel.FileDto{}, err
	}

	return commonModel.FileDto{
		ID:          fileRecord.ID,
		Name:        fileRecord.Name,
		Key:         key,
		StorageType: fileRecord.StorageType,
		URL:         normalizedURL,
		ContentType: contentType,
		Category:    string(category),
		Size:        0,
		Width:       fileRecord.Width,
		Height:      fileRecord.Height,
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if id == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	fileRecord, err := s.fileRepository.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	if err := s.transactor.Run(context.Background(), func(ctx context.Context) error {
		return s.DeleteFileRecord(ctx, fileRecord.ID)
	}); err != nil {
		return err
	}
	if storage.NormalizeStorageType(fileRecord.StorageType) != storage.StorageTypeExternal {
		_ = s.DeleteStoredFile(fileRecord.StorageType, fileRecord.Key)
	}
	return nil
}

func (s *FileService) GetFileByID(ctx context.Context, id string) (commonModel.FileDto, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	fileRecord, err := s.fileRepository.GetByID(context.Background(), id)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	return commonModel.FileDto{
		ID:          fileRecord.ID,
		Name:        fileRecord.Name,
		Key:         fileRecord.Key,
		StorageType: fileRecord.StorageType,
		URL:         fileRecord.URL,
		ContentType: fileRecord.ContentType,
		Category:    fileRecord.Category,
		Size:        fileRecord.Size,
		Width:       fileRecord.Width,
		Height:      fileRecord.Height,
	}, nil
}

func (s *FileService) UpdateFileMeta(
	ctx context.Context,
	id string,
	dto commonModel.UpdateFileMetaDto,
) (commonModel.FileDto, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if id == "" || dto.Size < 0 {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}
	if dto.Width != nil && *dto.Width < 0 {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}
	if dto.Height != nil && *dto.Height < 0 {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}

	fileRecord, err := s.fileRepository.GetByID(context.Background(), id)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if storage.NormalizeStorageType(fileRecord.StorageType) != storage.StorageTypeObject {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}

	var contentTypePtr *string
	if contentType := strings.TrimSpace(dto.ContentType); contentType != "" {
		contentTypePtr = &contentType
	}

	updated, err := s.fileRepository.UpdateMetaByID(
		context.Background(),
		id,
		dto.Size,
		dto.Width,
		dto.Height,
		contentTypePtr,
	)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	return commonModel.FileDto{
		ID:          updated.ID,
		Name:        updated.Name,
		Key:         updated.Key,
		StorageType: updated.StorageType,
		URL:         updated.URL,
		ContentType: updated.ContentType,
		Category:    updated.Category,
		Size:        updated.Size,
		Width:       updated.Width,
		Height:      updated.Height,
	}, nil
}

func (s *FileService) ListFiles(
	ctx context.Context,
	query commonModel.FileListQueryDto,
) (commonModel.FileListResultDto, error) {
	result := commonModel.FileListResultDto{Items: []commonModel.FileListItemDto{}}
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	storageType := strings.TrimSpace(query.StorageType)
	if storageType != "" {
		normalized := storage.NormalizeStorageType(storageType)
		if normalized != storage.StorageTypeLocal && normalized != storage.StorageTypeObject {
			return result, errors.New(commonModel.INVALID_PARAMS)
		}
		storageType = string(normalized)
	}

	files, total, err := s.fileRepository.ListByStorageTypeAndSearch(
		context.Background(),
		storageType,
		query.Search,
		page,
		pageSize,
	)
	if err != nil {
		return result, err
	}

	items := make([]commonModel.FileListItemDto, 0, len(files))
	for _, f := range files {
		items = append(items, commonModel.FileListItemDto{
			ID:          f.ID,
			Name:        f.Name,
			Key:         f.Key,
			StorageType: f.StorageType,
			URL:         f.URL,
			ContentType: f.ContentType,
			Size:        f.Size,
			CreatedAt:   f.CreatedAt,
		})
	}
	result.Total = total
	result.Items = items
	return result, nil
}

func (s *FileService) ListFileTree(
	ctx context.Context,
	query commonModel.FileTreeQueryDto,
) (commonModel.FileTreeResultDto, error) {
	result := commonModel.FileTreeResultDto{Items: []commonModel.FileTreeNodeDto{}}
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	storageTypeRaw := strings.TrimSpace(query.StorageType)
	if storageTypeRaw == "" {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}
	storageType := storage.NormalizeStorageType(storageTypeRaw)
	if storageType != storage.StorageTypeLocal && storageType != storage.StorageTypeObject {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}
	prefix := strings.Trim(strings.TrimSpace(query.Prefix), "/")

	selector := s.getSelector()
	nodes, err := selector.ListNodes(context.Background(), storageType, prefix)
	if err != nil {
		return result, err
	}

	keyCandidatesByPath := make(map[string][]string, len(nodes))
	keySet := make(map[string]struct{}, len(nodes)*2)
	for _, node := range nodes {
		if node.IsDir {
			continue
		}
		candidates := selector.ResolveKeyCandidatesByPath(storageType, node.Path)
		if len(candidates) == 0 {
			continue
		}
		keyCandidatesByPath[node.Path] = candidates
		for _, key := range candidates {
			keySet[key] = struct{}{}
		}
	}
	fileKeys := make([]string, 0, len(keySet))
	for key := range keySet {
		fileKeys = append(fileKeys, key)
	}
	idByKey := map[string]string{}
	if len(fileKeys) > 0 {
		dbFiles, err := s.fileRepository.ListByStorageTypeAndKeys(context.Background(), string(storageType), fileKeys)
		if err != nil {
			return result, err
		}
		idByKey = make(map[string]string, len(dbFiles))
		for _, f := range dbFiles {
			idByKey[f.Key] = f.ID
		}
	}
	// Compatibility fallback: keep URL mapping as last resort.
	idByURL := map[string]string{}
	urlByPath := make(map[string]string, len(nodes))
	fileURLs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.IsDir {
			continue
		}
		url := selector.ResolveURLByPath(storageType, node.Path)
		if url == "" {
			continue
		}
		urlByPath[node.Path] = url
		fileURLs = append(fileURLs, url)
	}
	if len(fileURLs) > 0 {
		dbFiles, err := s.fileRepository.ListByStorageTypeAndURLs(context.Background(), string(storageType), fileURLs)
		if err != nil {
			return result, err
		}
		idByURL = make(map[string]string, len(dbFiles))
		for _, f := range dbFiles {
			idByURL[f.URL] = f.ID
		}
	}

	items := make([]commonModel.FileTreeNodeDto, 0, len(nodes))
	for _, node := range nodes {
		item := commonModel.FileTreeNodeDto{
			Name:        node.Name,
			Path:        node.Path,
			NodeType:    treeNodeTypeFile,
			HasChildren: false,
			Size:        node.Size,
			ContentType: node.ContentType,
			ModifiedAt:  node.LastModified,
		}
		if node.IsDir {
			item.NodeType = treeNodeTypeFolder
			item.HasChildren = true
			item.Size = 0
			item.ContentType = ""
		} else if candidates, ok := keyCandidatesByPath[node.Path]; ok {
			for _, key := range candidates {
				if id := idByKey[key]; id != "" {
					item.FileID = id
					break
				}
			}
			if item.FileID == "" {
				logUtil.GetLogger().Warn(
					"Tree key mapping missing file id, fallback to url mapping",
					zap.String("path", node.Path),
					zap.Strings("key_candidates", candidates),
					zap.String("storage_type", string(storageType)),
				)
				if url, ok := urlByPath[node.Path]; ok {
					item.FileID = idByURL[url]
				}
			}
		} else if url, ok := urlByPath[node.Path]; ok {
			item.FileID = idByURL[url]
		}
		items = append(items, item)
	}

	result.Items = items
	return result, nil
}

func (s *FileService) StreamFileByID(ctx *gin.Context, id string) {
	fileRecord, err := s.fileRepository.GetByID(context.Background(), id)
	if err != nil {
		ctx.String(http.StatusNotFound, "文件不存在")
		return
	}

	contentType := fileRecord.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx.Header("Content-Type", contentType)

	normalizedStorageType := storage.NormalizeStorageType(fileRecord.StorageType)
	if normalizedStorageType == storage.StorageTypeExternal {
		ctx.Redirect(http.StatusTemporaryRedirect, fileRecord.URL)
		return
	}

	reader, err := s.getSelector().Get(context.Background(), normalizedStorageType, fileRecord.Key)
	if err != nil {
		ctx.String(http.StatusNotFound, "文件不存在")
		return
	}
	s.streamReader(ctx, reader, fileRecord.Name, contentType, fileRecord.CreatedAt, fileRecord.ID, string(normalizedStorageType))
}

func (s *FileService) StreamFileByPath(ctx *gin.Context, query commonModel.FilePathStreamQueryDto) {
	userid := viewer.MustFromContext(ctx.Request.Context()).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil || !user.IsAdmin {
		ctx.String(http.StatusForbidden, "无权限")
		return
	}
	storageType := storage.NormalizeStorageType(query.StorageType)
	if storageType != storage.StorageTypeLocal && storageType != storage.StorageTypeObject {
		ctx.String(http.StatusBadRequest, "非法存储类型")
		return
	}
	filePath := strings.Trim(strings.TrimSpace(query.Path), "/")
	if filePath == "" {
		ctx.String(http.StatusBadRequest, "非法文件路径")
		return
	}
	contentType := strings.TrimSpace(query.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fileName := strings.TrimSpace(query.Name)
	if fileName == "" {
		fileName = path.Base(filePath)
	}
	if fileName == "" {
		fileName = "file"
	}
	selector := s.getSelector()
	if reader, pathErr := selector.GetByStoragePath(context.Background(), storageType, filePath); pathErr == nil {
		s.streamReader(ctx, reader, fileName, contentType, time.Now(), "", string(storageType)+":path:"+filePath)
		return
	}
	candidates := selector.ResolveKeyCandidatesByPath(storageType, filePath)
	if len(candidates) == 0 {
		ctx.String(http.StatusNotFound, "文件不存在")
		return
	}
	var reader io.ReadCloser
	var resolvedKey string
	for _, key := range candidates {
		reader, err = selector.Get(context.Background(), storageType, key)
		if err == nil {
			resolvedKey = key
			break
		}
	}
	if reader == nil {
		ctx.String(http.StatusNotFound, "文件不存在")
		return
	}
	s.streamReader(ctx, reader, fileName, contentType, time.Now(), "", string(storageType)+":"+resolvedKey)
}

func (s *FileService) streamReader(
	ctx *gin.Context,
	reader io.ReadCloser,
	fileName string,
	contentType string,
	modTime time.Time,
	fileID string,
	storageType string,
) {
	defer func() { _ = reader.Close() }()
	ctx.Header("Content-Type", contentType)

	readSeeker, ok := reader.(io.ReadSeeker)
	if ok {
		http.ServeContent(ctx.Writer, ctx.Request, fileName, modTime, readSeeker)
		return
	}
	if _, err := io.Copy(ctx.Writer, reader); err != nil {
		logUtil.GetLogger().Warn(
			"stream file copy failed",
			zap.String("file_id", fileID),
			zap.String("storage_type", storageType),
			zap.Error(err),
		)
	}
}

func (s *FileService) GetFilePresignURL(
	ctx context.Context,
	dto *commonModel.GetPresignURLDto,
) (commonModel.PresignDto, error) {
	var result commonModel.PresignDto
	userid := viewer.MustFromContext(ctx).UserID()

	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if dto.FileName == "" {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}
	if st := strings.TrimSpace(dto.StorageType); st != "" &&
		storage.NormalizeStorageType(st) != storage.StorageTypeObject {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}

	contentType := dto.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	category := storage.CategoryImage
	if strings.HasPrefix(contentType, "audio/") {
		category = storage.CategoryAudio
	}
	if !isAllowedType(contentType, config.Config().Upload.AllowedTypes) {
		return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	key, err := s.keyGen.GenerateKey(category, userid, dto.FileName)
	if err != nil {
		return result, err
	}

	selector := s.getSelector()
	presignURL, err := selector.PresignPutURL(context.Background(), key, 24*time.Hour)
	if err != nil {
		return result, err
	}

	fileURL := selector.ResolveURL(storage.StorageTypeObject, key)
	storageType, provider, bucket := currentStorageRoute(selector, storage.StorageTypeObject)
	fileRecord := &fileModel.File{
		Key:         key,
		StorageType: storageType,
		Provider:    provider,
		Bucket:      bucket,
		URL:         fileURL,
		Name:        dto.FileName,
		ContentType: contentType,
		Category:    string(category),
		UserID:      userid,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return result, err
	}

	result.ID = fileRecord.ID
	result.FileName = dto.FileName
	result.ContentType = contentType
	result.Key = key
	result.PresignURL = presignURL
	result.FileURL = fileURL
	return result, nil
}

func (s *FileService) CleanupOrphanFiles() error {
	ctx := context.Background()
	threshold := time.Now().UTC().Add(-24 * time.Hour)

	files, err := s.fileRepository.GetOrphanFiles(ctx, threshold)
	if err != nil {
		return err
	}

	var systemSetting settingModel.SystemSetting
	if err := s.settingService.GetSetting(&systemSetting); err != nil {
		return err
	}

	users, err := s.commonRepository.GetAllUsers(ctx)
	if err != nil {
		return err
	}
	isUserAvatar := make(map[string]bool)
	for _, user := range users {
		if user.Avatar != "" {
			isUserAvatar[user.Avatar] = true
		}
	}

	for _, file := range files {
		if file.Key != "" &&
			file.URL != systemSetting.ServerLogo &&
			!isUserAvatar[file.URL] &&
			storage.NormalizeStorageType(file.StorageType) != storage.StorageTypeExternal {
			_ = s.DeleteStoredFile(file.StorageType, file.Key)
		}
		_ = s.fileRepository.Delete(ctx, file.ID)
	}

	return nil
}

func (s *FileService) DeleteFileRecord(ctx context.Context, id string) error {
	return s.fileRepository.Delete(ctx, id)
}

func (s *FileService) DeleteStoredFile(storageType string, key string) error {
	if key == "" {
		return nil
	}
	normalizedStorageType := storage.NormalizeStorageType(storageType)
	if normalizedStorageType == storage.StorageTypeExternal {
		return nil
	}
	return s.getSelector().Delete(context.Background(), normalizedStorageType, key)
}

func (s *FileService) keyGenForCategory(category storage.Category, fileName string) storage.KeyGenerator {
	_ = category
	_ = fileName
	return s.keyGen
}

func isAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func (s *FileService) getSelector() *storage.StorageSelector {
	if s.storageManager == nil {
		return storage.NewStorageSelector(config.Config().Storage)
	}
	if selector := s.storageManager.GetSelector(); selector != nil {
		return selector
	}
	return storage.NewStorageSelector(config.Config().Storage)
}

func currentStorageRoute(
	selector *storage.StorageSelector,
	storageType storage.StorageType,
) (resolvedType, provider, bucket string) {
	switch storage.NormalizeStorageType(string(storageType)) {
	case storage.StorageTypeObject:
		if selector != nil {
			provider, bucket = selector.ObjectRoute()
		}
		return string(storage.StorageTypeObject), provider, bucket
	default:
		return string(storage.StorageTypeLocal), "", ""
	}
}
